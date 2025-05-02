// MIT License
//
// Copyright (c) 2025 chaunsin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package cookiecloud

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var ErrCookieNotfound = fmt.Errorf("cookie not found")

type Body struct {
	Uuid      string `json:"uuid"`
	Encrypted string `json:"encrypted"`
}

type Cookie struct {
	CookieData       map[string][]CookieData      `json:"cookie_data"`
	LocalStorageData map[string]map[string]string `json:"local_storage_data"`
	UpdateTime       time.Time                    `json:"update_time"`
}

type CookieData struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HttpOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreId        string  `json:"storeId"`
	Value          string  `json:"value"`
}

func (c CookieData) GetExpired() time.Time {
	sec := int64(c.ExpirationDate)
	nsec := int64((c.ExpirationDate - float64(sec)) * 1e9)
	return time.Unix(sec, nsec)
}

type GetReq struct {
	Uuid            string `json:"-"`
	Password        string `json:"password"`
	CloudDecryption bool   `json:"-"` // 如果指定了ture则在服务端解密并返回,为了安全不建议这么使用。
}

type GetResp struct {
	Body
	Cookie
}

type UpdateReq struct {
	Uuid     string `json:"uuid"`
	Password string `json:"password"`
	Cookie
}

type UpdateResp struct {
	Action string `json:"action"` // done error
}

type Config struct {
	Url     string        `json:"url" yaml:"url"`
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	Retry   int           `json:"retry" yaml:"retry"`
	Debug   bool          `json:"debug" yaml:"debug"`
}

type Client struct {
	cfg *Config
	cli *resty.Client
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg.Url == "" {
		log.Printf("cookiecloud: waring use default host")
		cfg.Url = "http://localhost:8088"
	}

	cli := resty.New()
	cli.SetBaseURL(cfg.Url)
	if cfg.Retry > 0 {
		cli.SetRetryCount(cfg.Retry)
	}
	if cfg.Timeout > 0 {
		cli.SetTimeout(cfg.Timeout)
	}
	cli.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	cli.SetDebug(cfg.Debug)
	return &Client{
		cfg: cfg,
		cli: cli,
	}, nil
}

func (c *Client) SetLogger(l resty.Logger) {
	c.cli.SetLogger(l)
}

func (c *Client) SetTransport(transport http.RoundTripper) {
	c.cli.SetTransport(transport)
}

func (c *Client) SetTLSClientConfig(config *tls.Config) {
	c.cli.SetTLSClientConfig(config)
}

func (c *Client) SetCookies(cs []*http.Cookie) {
	c.cli.SetCookies(cs)
}

func (c *Client) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		c.cli.SetHeader(k, v)
	}
}

// X returns the internal resty.Client
func (c *Client) X() *resty.Client {
	return c.cli
}

// Get sends a request to the server to get the cookie.
// see: https://github.com/easychen/CookieCloud/blob/master/api/app.js#L46
func (c *Client) Get(ctx context.Context, req *GetReq) (*GetResp, error) {
	if req.Uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	var (
		resp GetResp
		cli  = c.cli.R().SetContext(ctx)
	)

	// 云端解密
	if req.CloudDecryption {
		cli = cli.SetBody(req)
	}

	res, err := cli.SetResult(&resp).Post("/get/" + req.Uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to request server: %v", err)
	}
	if res.StatusCode() == 404 {
		return nil, ErrCookieNotfound
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("server return status %d body %+v", res.StatusCode(), resp)
	}
	if req.CloudDecryption {
		return &resp, nil
	}

	// 本地解密逻辑
	keyPassword := Md5String(req.Uuid, "-", req.Password)[:16]
	decrypted, err := Decrypt(keyPassword, resp.Encrypted)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %v", err)
	}
	var cookie Cookie
	if err := json.Unmarshal(decrypted, &cookie); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted data as json: %v", err)
	}
	resp.Cookie = cookie
	return &resp, nil
}

// Update sends a request to the server to update the cookie.
// see: https://github.com/easychen/CookieCloud/blob/master/api/app.js#L28
func (c *Client) Update(ctx context.Context, req *UpdateReq) (*UpdateResp, error) {
	if req.Uuid == "" {
		return nil, fmt.Errorf("uuid is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}
	keyPassword := Md5String(req.Uuid, "-", req.Password)[:16]
	encrypted, err := Encrypt(keyPassword, string(data))
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	var (
		request = Body{
			Uuid:      req.Uuid,
			Encrypted: encrypted,
		}
		body UpdateResp
	)
	res, err := c.cli.R().SetContext(ctx).SetBody(&request).SetResult(&body).Post("/update")
	if err != nil {
		return nil, fmt.Errorf("failed to request server: %v", err)
	}
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("server return status %d body %+v", res.StatusCode(), body)
	}
	if body.Action != "done" {
		return nil, fmt.Errorf("server return action %s", body.Action)
	}
	return &body, nil
}
