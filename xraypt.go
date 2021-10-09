package main

import (
	"fmt"
	"log"
	"sync"
	"sort"
	"strconv"
	"strings"
	"os"
	"io/ioutil"

	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
)

var subs []string

var protocols = []string{
	"vmess",
	"vless",
	"ss",
	"ssr",
	"trojan"}

func main() {
	tools.PreCheck(8123)
	//os.Exit(0)

	//vless := []byte("vless://bdc07b5f-dd93-4c29-8fcf-25327ac2a55a@v2rayge1.free3333.xyz:443?encryption=none&security=tls&type=ws&host=v2rayge1.free3333.xyz&path=%2fray#https%3a%2f%2fgithub.com%2fAlvin9999%2fnew-pac%2fwiki%2bVLESS%e5%be%b7%e5%9b%bdi")
	//var out tools.Outbound
	//tools.VlLinkToOut(&out, string(vless))
	////fmt.Printf("%+v\n", out)
	//os.Exit(0)

	byteData, err := ioutil.ReadFile(tools.SubsFilePath)
	if err != nil {
		log.Println("SubFile read error:", err)
	}else{
		log.Println("SubFile get...")
		subs = strings.Fields(string(byteData))
	}

	var ports []int
	var vmLinks []string

	vmLinks = tools.SubGet(protocols, subs)
	log.Println("Subs get done!")

	var goodPingNodes []*tools.Node
	var wgPing sync.WaitGroup
	pingJob := make(chan *tools.Node, len(vmLinks))
	pingResult := make(chan *tools.Node, len(vmLinks))

	ports, err = tools.GetFreePorts(tools.PThreadNum)
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range vmLinks {
		var n tools.Node
		//n.Init(strconv.Itoa(i), s, ports[i])
		n.Init(strconv.Itoa(i), s)
		//n.CreateJson(tools.TempPath)

		pingJob <- &n
		wgPing.Add(1)
	}
	for i := 1; i <= tools.PThreadNum; i++ {
		go ping.XrayPing(&wgPing, pingJob, pingResult, ports[i-1])
	}
	close(pingJob)

	log.Println("waitting for ping to be done...")
	wgPing.Wait()
	log.Println("ping finished")
	goodPingCnt := len(pingResult)
	log.Printf("There are %d good pings\n", goodPingCnt)

	for i := 1; i <= goodPingCnt; i++  {
		n := <-pingResult
		goodPingNodes = append(goodPingNodes, n)
	}

	sort.Stable(tools.ByDelay(goodPingNodes))
	for i, n := range goodPingNodes {
		fmt.Println(i, n.AvgDelay)
	}

	var halfGoodVmLinks []string
	for i, n := range goodPingNodes {
		n.CreateFinalJson(tools.HalfJsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", n.ShareLink, "\nDelay:", strconv.Itoa(n.AvgDelay)}
		vmOutStr := strings.Join(str, "")
		halfGoodVmLinks = append(halfGoodVmLinks, vmOutStr)
	}
	if len(halfGoodVmLinks) != 0 {
		bytes := []byte(strings.Join(halfGoodVmLinks[:], "\n"))
		err := os.WriteFile("vmHalfOut.txt", bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("vmHalfOut generated!")
		}
	}

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	//os.Exit(0)

	var goodSpeedNodes []*tools.Node
	var wgSpeed sync.WaitGroup
	speedJob := make(chan *tools.Node, goodPingCnt)
	speedResult := make(chan *tools.Node, goodPingCnt)

	for _, n := range goodPingNodes {
		speedJob <- n
		wgSpeed.Add(1)
	}

	ports, err = tools.GetFreePorts(tools.SThreadNum)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= tools.SThreadNum; i++ {
		go speedtest.XraySpeedTest(&wgSpeed, speedJob, speedResult, ports[i-1])
	}
	close(speedJob)
	wgSpeed.Wait()
	goodSpeeds := len(speedResult)
	for i := 1; i <= goodSpeeds; i++ {
		n := <-speedResult
		goodSpeedNodes = append(goodSpeedNodes, n)
	}

	//var halfGoodVmLinks []string
	//for i, n := range goodPingNodes {
	//	n.CreateFinalJson(tools.HalfJsonsPath, strconv.Itoa(i))
	//	str := []string{strconv.Itoa(i), "\n", n.ShareLink, "\nDelay:", strconv.Itoa(n.AvgDelay)}
	//	vmOutStr := strings.Join(str, "")
	//	halfGoodVmLinks = append(halfGoodVmLinks, vmOutStr)
	//}
	//if len(halfGoodVmLinks) != 0 {
	//	bytes := []byte(strings.Join(halfGoodVmLinks[:], "\n"))
	//	err := os.WriteFile("vmHalfOut.txt", bytes, 0644)
	//	if err != nil {
	//		log.Println(err)
	//	}else{
	//		log.Println("vmHalfOut generated!")
	//	}
	//}

	sort.Sort(tools.ByULSpeed(goodSpeedNodes))
	sort.Stable(tools.ByDelay(goodSpeedNodes))
	sort.Stable(tools.ByDLSpeed(goodSpeedNodes))
	var goodVmLinks []string
	for i, n := range goodSpeedNodes {
		fmt.Println(i, (*n).AvgDelay, (*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
		//(*n).Id = strconv.Itoa(i)
		n.CreateFinalJson(tools.JsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", (*n).ShareLink, "\nDelay:", strconv.Itoa((*n).AvgDelay), " Down: ", fmt.Sprintf("%.2f", (*n).DLSpeed), " Up: ", fmt.Sprintf("%.2f", (*n).ULSpeed), " Country: ", (*n).Country, "\n"}
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

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)
	//os.RemoveAll(tools.JitPath)
	//os.MkdirAll(tools.JitPath, 0755)
	//os.RemoveAll(tools.HalfJsonsPath)
	//os.MkdirAll(tools.HalfJsonsPath, 0755)
}

