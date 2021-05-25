package main

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	//"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
	//"time"
)

const pTimeout = 1500 //ms
const pCount = 5
const sTimeout = 15000 //ms

var subs = []string{"https://raw.githubusercontent.com/ssrsub/ssr/master/v2ray", "https://jiang.netlify.com", "https://raw.githubusercontent.com/freefq/free/master/v2"}

type Node struct {
	ShareLink string
	JsonPath string
	AvgDelay int
	Country string
	DLSpeed float64
	ULSpeed float64
}

func main() {
	//var vmLinks []string
	//vmLinks = tools.SubGetVms(subs)
	//for _, s := range vmLinks {
	//}

	vm := "vmess://eyJ2IjoiMiIsInBzIjoid3d3LnlvdW5lZWQud2luIiwiYWRkIjoiMjMuMjI1LjU3LjIwMyIsInBvcnQiOiI0NDMiLCJpZCI6IjgxMTc4MmQ5LTZjZGItNDVkZC05NDQ4LTFlYzRjNDdhZDU2NCIsImFpZCI6IjY0IiwibmV0Ijoid3MiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiJ3d3cuMzQ0MjgzOTQueHl6IiwicGF0aCI6Ii9wYXRoLzMxMDkxMDIxMTkxNiIsInRscyI6InRscyJ9"

	var vmout tools.VmessOut
	var con tools.Config
	tools.VmLinkToVmOut(&vmout, vm)
	tools.VmOutToConfig(&con, vmout)

	byteValue, err := json.MarshalIndent(con, "", "    ")
	if err != nil {
		fmt.Println(err)
	}

	err = ioutil.WriteFile("o.json", byteValue, 0644)
	if err != nil {
		fmt.Println(err)
	}

	//pingJob := make(chan string, 2)
	//result := make(chan int, 2)
	//go ping.XrayPing(pingJob, result, pCount, pTimeout)
	//go ping.XrayPing(pingJob, result, pCount, pTimeout)
	//pingJob <- "0.json"
	//pingJob <- "o.json"
	//close(pingJob)
	//for a := 1; a <= 2; a++ {
	//	avgDelay := <-result
	//	fmt.Println("avg0:", avgDelay)
	//}
//
	country, DLSpeed, ULSpeed := speedtest.XraySpeedTest("o.json", sTimeout)
	fmt.Println(country, " ", DLSpeed, " ", ULSpeed)

	//time.Sleep(10*time.Second)
}

