package main

import (
	"fmt"
	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
)

func main() {
	avgDelay, err := ping.XrayPing("x.json", 5, 1500)
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println("avgDelay:", avgDelay)
	}

	country, DLSpeed, ULSpeed := speedtest.XraySpeedTest("x.json", 15000)
	fmt.Println(country, " ", DLSpeed, " ", ULSpeed)
}

