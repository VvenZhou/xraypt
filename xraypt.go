package main

import (
	"fmt"
	"log"
	"sync"
	"sort"
	"strconv"
	"strings"
	"os"
	"time"

	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
)

const pCount = 9
const pT = 1500 //ms
const pRealCount = 3
const pRealT = 2000 //ms

const sT = 20000 //ms

const threadPingCnt = 50
const threadSpeedCnt = 4
const DSLine = 5.0

const pTimeout = time.Duration(pT) * time.Millisecond
const pRealTimeout = time.Duration(pRealT) * time.Millisecond
const sTimeout = time.Duration(sT) * time.Millisecond


var subs = []string{"https://raw.githubusercontent.com/ssrsub/ssr/master/v2ray", "https://jiang.netlify.com", "https://raw.githubusercontent.com/freefq/free/master/v2"}
var subJ = []string{"https://raw.githubusercontent.com/freefq/free/master/v2"}
var subA = []string{""}


func main() {

	os.RemoveAll("jit/")
	os.MkdirAll("jit/", 0755)
	//vless := []byte("vless://bdc07b5f-dd93-4c29-8fcf-25327ac2a55a@v2rayge1.free3333.xyz:443?encryption=none&security=tls&type=ws&host=v2rayge1.free3333.xyz&path=%2fray#https%3a%2f%2fgithub.com%2fAlvin9999%2fnew-pac%2fwiki%2bVLESS%e5%be%b7%e5%9b%bdi")
	//var out tools.Outbound
	//tools.VlLinkToOut(&out, string(vless))
	////fmt.Printf("%+v\n", out)
	//os.Exit(0)


	//var nodes []tools.Node
	var ports []int
	var vmLinks []string
	vmLinks = tools.SubGetVms(subs)
	ports, err := tools.GetFreePorts(len(vmLinks))
	if err != nil {
		log.Fatal(err)
	}


	var goodPingNodes []*tools.Node
	var wgPing sync.WaitGroup
	pingJob := make(chan *tools.Node, 500)
	pingResult := make(chan *tools.Node, 500)

	for i, s := range vmLinks {
		var n tools.Node
		n.Init(strconv.Itoa(i), s, ports[i])
		n.CreateJson("temp/")

		pingJob <- &n
		wgPing.Add(1)
	}
	for i := 1; i <= threadPingCnt; i++ {
		go ping.XrayPing(&wgPing, pingJob, pingResult, pCount, pTimeout, pRealCount, pRealTimeout)
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
	for _, n := range goodPingNodes {
		fmt.Println("avgDelay:", (*n).AvgDelay)
	}

	var goodSpeedNodes []*tools.Node
	var wgSpeed sync.WaitGroup
	speedJob := make(chan *tools.Node, goodPingCnt)
	speedResult := make(chan *tools.Node, goodPingCnt)

	for _, n := range goodPingNodes {
		speedJob <- n
		wgSpeed.Add(1)
	}
	for i := 1; i <= threadSpeedCnt; i++ {
		go speedtest.XraySpeedTest(&wgSpeed, speedJob, speedResult, sTimeout, DSLine)
	}
	close(speedJob)
	wgSpeed.Wait()
	goodSpeeds := len(speedResult)
	for i := 1; i <= goodSpeeds; i++ {
		n := <-speedResult
		if (*n).DLSpeed > 5.0 {
			log.Println("Speed got one!")
			goodSpeedNodes = append(goodSpeedNodes, n)
		}
		log.Println("DownSpeed too slow, abandoned")
	}

	sort.Sort(tools.ByULSpeed(goodSpeedNodes))
	sort.Stable(tools.ByDelay(goodSpeedNodes))
	sort.Stable(tools.ByDLSpeed(goodSpeedNodes))

	var goodVmLinks []string
	for i, n := range goodSpeedNodes {
		fmt.Println((*n).AvgDelay, (*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
		//(*n).Id = strconv.Itoa(i)
		(*n).CreateFinalJson("jsons/", strconv.Itoa(i))
		str := []string{(*n).ShareLink, "\nDown: ", fmt.Sprintf("%.2f", (*n).DLSpeed), " Up: ", fmt.Sprintf("%.2f", (*n).ULSpeed), " Country: ", (*n).Country}
		vmOutStr := strings.Join(str, "")
		goodVmLinks = append(goodVmLinks, vmOutStr)
	}
	if len(goodVmLinks) != 0 {
		bytes := []byte(strings.Join(goodVmLinks[:], "\n"))
		err := os.WriteFile("vmOut.txt", bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("vmOut generated!")
		}
	}

	os.RemoveAll("temp/")
	os.MkdirAll("temp/", 0755)
}

