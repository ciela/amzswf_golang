package main

import (
	"fmt"

	"github.com/sclasen/swf4go"
)

func main() {
	client := swf.NewClient(swf.MustGetenv("AWS_ACCESS_KEY_ID"), swf.MustGetenv("AWS_SECRET_ACCESS_KEY"), swf.APNorthEast1)
	client.Debug = true

	// ドメインリスト取得
	req := swf.ListDomainsRequest{
		RegistrationStatus: "REGISTERED",
	}
	res, err := client.ListDomains(req)
	if err != nil {
		fmt.Println(err)
		return
	}

}
