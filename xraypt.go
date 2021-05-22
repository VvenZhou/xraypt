package main

import (
	"fmt"
	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
)

func main() {
	avgDelay, err := ping.XrayPing("x.json")
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println("avgDelay:", avgDelay)
	}
}

