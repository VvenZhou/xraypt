package main

import (
	"fmt"
	//"encoding/json"
	//"io/ioutil"
	"github.com/VvenZhou/xraypt/src/ping"
	//"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
	//"time"
	"sync"
)

const pTimeout = 1500 //ms
const pCount = 5
const sTimeout = 15000 //ms

var subs = []string{"https://raw.githubusercontent.com/ssrsub/ssr/master/v2ray", "https://jiang.netlify.com", "https://raw.githubusercontent.com/freefq/free/master/v2"}

func main() {
	//var vmLinks []string
	//vmLinks = tools.SubGetVms(subs)
	//for _, s := range vmLinks {
	//}

	vm := "vmess://eyJ2IjoiMiIsInBzIjoid3d3LnlvdW5lZWQud2luIiwiYWRkIjoiMjMuMjI1LjU3LjIwMyIsInBvcnQiOiI0NDMiLCJpZCI6IjgxMTc4MmQ5LTZjZGItNDVkZC05NDQ4LTFlYzRjNDdhZDU2NCIsImFpZCI6IjY0IiwibmV0Ijoid3MiLCJ0eXBlIjoibm9uZSIsImhvc3QiOiJ3d3cuMzQ0MjgzOTQueHl6IiwicGF0aCI6Ii9wYXRoLzMxMDkxMDIxMTkxNiIsInRscyI6InRscyJ9"

	vm1 := "vmess://eyJ2IjogIjIiLCAicHMiOiAiZ2l0aHViLmNvbS9mcmVlZnEgLSBcdTdmOGVcdTU2ZmRDbG91ZGlubm92YXRpb25cdTY1NzBcdTYzNmVcdTRlMmRcdTVmYzMgMzUiLCAiYWRkIjogIjE1NC44NC4xLjM1IiwgInBvcnQiOiAiNDQzIiwgImlkIjogIjA0MTU3NDZjLTRkNmItNDlmYi05YThhLWU3NGFkNjE3MmQzZCIsICJhaWQiOiAiNjQiLCAibmV0IjogIndzIiwgInR5cGUiOiAibm9uZSIsICJob3N0IjogInd3dy4wMDcyMjU0Mi54eXoiLCAicGF0aCI6ICIvcGF0aC8zMTA5MTAyMTE5MTYiLCAidGxzIjogInRscyJ9"

	var wgPing sync.WaitGroup
	pingJob := make(chan string, 100)
	result := make(chan *tools.Node, 100)

	pingJob <- vm
	pingJob <- vm1
	wgPing.Add(1)
	wgPing.Add(1)
	go ping.XrayPing(&wgPing, pingJob, result, pCount, pTimeout)
	go ping.XrayPing(&wgPing, pingJob, result, pCount, pTimeout)
	close(pingJob)
	wgPing.Wait()
	n := <-result
	n2 := <-result
	fmt.Println("avg0:", (*n).AvgDelay)
	fmt.Println("avg0:", (*n2).AvgDelay)

	//var wgSpeed sync.WaitGroup
	//speedJob := make(chan *tools.Node, 100)
	//speedResult := make(chan *tools.Node, 100)
	//speedJob <- n
	//wgSpeed.Add(1)
	//go speedtest.XraySpeedTest(&wgSpeed, speedJob, speedResult, sTimeout)
	//wgSpeed.Wait()
	//fmt.Println((*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)

	//go speedtest.XraySpeedTest(&wgSpeed, , sTimeout)
	//fmt.Println(country, " ", DLSpeed, " ", ULSpeed)

	//time.Sleep(10*time.Second)
}

