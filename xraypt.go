package main

import (
	"fmt"
	ping "github.com/VvenZhou/xraypt/src"
)

func main() {
	avgDelay, err := ping.XrayPing("x.json")
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println(avgDelay)
	}
}
