package main

import (
	"fmt"
	"github.com/VvenZhou/xraypt/src/ping"
)

func main() {
	avgDelay, err := ping.XrayPing("x.json")
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println(avgDelay)
	}
}
