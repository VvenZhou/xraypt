package main

import (
	"fmt"
	"log"
	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
	//"time"
	"sync"
	"sort"
	"strconv"
)

const pTimeout = 1500 //ms
const pCount = 5
const sTimeout = 20000 //ms

var subs = []string{"https://raw.githubusercontent.com/ssrsub/ssr/master/v2ray", "https://jiang.netlify.com", "https://raw.githubusercontent.com/freefq/free/master/v2"}
var subJ = []string{"https://raw.githubusercontent.com/freefq/free/master/v2"}

func main() {
	var goodPingNodes []*tools.Node
	var wgPing sync.WaitGroup
	pingJob := make(chan string, 500)
	pingResult := make(chan *tools.Node, 500)

	for i := 1; i <= 50; i++ {
		go ping.XrayPing(&wgPing, pingJob, pingResult, pCount, pTimeout)
	}

	var vmLinks []string
	vmLinks = tools.SubGetVms(subJ)
	for _, s := range vmLinks {
		pingJob <- s
		wgPing.Add(1)
	}
	close(pingJob)

	log.Println("waitting for ping to be done...")
	wgPing.Wait()
	log.Println("ping finished")
	goodPingCnt := len(pingResult)
	log.Printf("there are %d good pings\n", goodPingCnt)

	for i := 1; i <= goodPingCnt; i++  {
		n := <-pingResult
		goodPingNodes = append(goodPingNodes, n)
	}
//	sort.Sort(tools.ByDelay(goodPingNodes))
	for _, n := range goodPingNodes {
		fmt.Println("avgDelay:", (*n).AvgDelay)
	}

	var goodSpeedNodes []*tools.Node
	var wgSpeed sync.WaitGroup
	speedJob := make(chan *tools.Node, goodPingCnt)
	speedResult := make(chan *tools.Node, goodPingCnt)

	for i := 1; i <= 4; i++ {
		go speedtest.XraySpeedTest(&wgSpeed, speedJob, speedResult, sTimeout)
	}
	for _, n := range goodPingNodes {
		speedJob <- n
		wgSpeed.Add(1)
	}
	close(speedJob)
	wgSpeed.Wait()
	goodSpeeds := len(speedResult)
	for i := 1; i <= goodSpeeds; i++ {
		n := <-speedResult
		goodSpeedNodes = append(goodSpeedNodes, n)
		//fmt.Println((*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
	}

	sort.Sort(tools.ByDelay(goodSpeedNodes))
	sort.Stable(tools.ByDLSpeed(goodSpeedNodes))
	sort.Stable(tools.ByULSpeed(goodSpeedNodes))

	for i, n := range goodSpeedNodes {
		fmt.Println((*n).AvgDelay, (*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
		(*n).Id = strconv.Itoa(i)
		(*n).CreateJson("jsons/")
	}
}

