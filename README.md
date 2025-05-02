# cookiecloud-go-sdk

[![GoDoc](https://godoc.org/github.com/chaunsin/cookiecloud-go-sdk?status.svg)](https://godoc.org/github.com/chaunsin/cookiecloud-go-sdk) [![Go Report Card](https://goreportcard.com/badge/github.com/chaunsin/cookiecloud-go-sdk)](https://goreportcard.com/report/github.com/chaunsin/cookiecloud-go-sdk)

一个支持CookieCloud的Go SDK

https://github.com/easychen/CookieCloud

## ✨ 功能

- 获取cookie
- 更新cookie
- 支持加解密
- 支持重试、限流、hook等功能，基于[resty](github.com/go-resty/resty)

## 🔨 安装

- go1.21+

```shell
go get github.com/chaunsin/cookiecloud-go-sdk@latest
```

## 🚀 示例

```go
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

```
