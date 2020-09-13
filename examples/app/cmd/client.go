package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/UnderTreeTech/waterdrop/pkg/log"

	"github.com/UnderTreeTech/protobuf/demo"
	"github.com/UnderTreeTech/waterdrop/pkg/server/rpc"
	"github.com/UnderTreeTech/waterdrop/pkg/status"

	"github.com/UnderTreeTech/waterdrop/pkg/server/http"

	"github.com/UnderTreeTech/waterdrop/pkg/trace/jaeger"

	"google.golang.org/grpc/resolver"

	"github.com/UnderTreeTech/waterdrop/pkg/registry/etcd"

	_ "github.com/UnderTreeTech/waterdrop/examples/app/internal/ecode"
	"github.com/UnderTreeTech/waterdrop/pkg/conf"
)

func main() {
	flag.Parse()

	conf.Init()
	defer log.New(nil).Sync()
	defer jaeger.Init()()

	etcdConf := &etcd.Config{}
	if err := conf.Unmarshal("etcd", etcdConf); err != nil {
		panic(fmt.Sprintf("unmarshal etcd config fail, err msg %s", err.Error()))
	}
	etcd := etcd.New(etcdConf)
	resolver.Register(etcd)

	httpCliConf := &http.ClientConfig{}
	if err := conf.Unmarshal("client.http.app", httpCliConf); err != nil {
		panic(fmt.Sprintf("unmarshal http client config fail, err msg %s", err.Error()))
	}
	fmt.Println("http client conf", httpCliConf)
	httpClient := http.NewClient(httpCliConf)
	r := &http.Request{
		URI:        "/api/app/validate/{id}",
		PathParams: map[string]string{"id": "1"},
		Body: `{
					"email": "example@example.com",
					"name": "John&Sun",
					"password": "styd.cn",
					"sex": 2,
					"age": 12,
					"addr": [
						{
							"mobile": "上海市徐汇区",
							"address": "<a onblur='alert(secret)' href='http://www.google.com'>Google</a>",
							"app": {
								"sappkey": "<p>md5hash</p>"
							},
							"reply": {
								"urls": [
									"www.&baidu.com",
									"www.g&oogle.com",
									"&#x6a;&#x61;&#x76;&#x61;&#x73;&#x63;&#x72;&#x69;&#x70;&#x74;&#x3a;&#x61;&#x6c;&#x65;&#x72;&#x74;&#x28;&#x31;&#x29;&#x3b;",
									"u003cimg src=1 onerror=alert(/xss/)u003e"
								]
							},
							"resp": [
								{
									"app_key": "sha1hash",
									"app_secret": "<href>rsa</href>"
								}
							]
						}
					]
				}`,
	}

	var result interface{}
	err := httpClient.Post(context.Background(), r, &result)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}

	cliConf := &rpc.ClientConfig{}
	if err := conf.Unmarshal("client.rpc.stardust", cliConf); err != nil {
		panic(fmt.Sprintf("unmarshal demo client config fail, err msg %s", err.Error()))
	}
	fmt.Println(cliConf)
	client := demo.NewDemoClient(rpc.NewClient(cliConf))
	now := time.Now()
	for i := 0; i < 1; i++ {
		_, err := client.SayHelloURL(context.Background(), &demo.HelloReq{Name: "John Sun"})
		if err != nil {
			fmt.Println("err", status.ExtractStatus(err))
		}
	}
	fmt.Println(time.Since(now))

	time.Sleep(time.Hour * 30)
}
