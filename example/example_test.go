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

package example

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/chaunsin/cookiecloud-go-sdk"
)

func Example() {
	cfg := cookiecloud.Config{
		Url: "http://localhost:8088",
	}
	cli, err := cookiecloud.NewClient(&cfg)
	if err != nil {
		log.Fatalf("init cookiecloud err: %s", err)
	}

	// 获取cookie
	getResp, err := cli.Get(context.Background(), &cookiecloud.GetReq{
		Uuid:            "example-uuid",
		Password:        "example-passsword",
		CloudDecryption: false, // 是否使用云端解密，ture:是(为了安全起见不建议使用云端解密),false:本地解密
	})
	if err != nil {
		if errors.Is(err, cookiecloud.ErrCookieNotfound) {
			log.Println("cookie not found")
			return
		}
		log.Fatalf("cookiecloud get cookie invald err: %s", err)
	}
	fmt.Printf("cookie: %+v\n", getResp.Cookie)
	fmt.Printf("local storage: %+v\n", getResp.LocalStorageData)

	// 更新cookie
	updateResp, err := cli.Update(context.Background(), &cookiecloud.UpdateReq{
		Cookie: cookiecloud.Cookie{
			CookieData: getResp.Cookie.CookieData,
		},
		Password: "example-password",
		Uuid:     "example-uuid",
	})
	if err != nil {
		log.Fatalf("更新失败: %s", err)
	}
	log.Println("返回:", updateResp.Action)
}
