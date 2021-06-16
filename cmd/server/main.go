package main

import (
	"github.com/mmfkupl/testtelegramgo"
)

func main() {
	appClient, err := testtelegramgo.GetAppClient("config.json")
	if err != nil {
		panic(err)
	}

	if err = appClient.StartAppClient(); err != nil {
		panic(err)
	}
}
