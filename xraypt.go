package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"os"

	"github.com/VvenZhou/xraypt/src/ping"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/monitor"
)


var protocols = []string{
	"vmess",
	"vless",
	"ss",
	"ssr",
	"trojan"}


func main() {

//	oldmain()
	tools.PreCheck(tools.MainPort, protocols)

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	cmdCh := make(chan string)	
	feedbackCh := make(chan int)	

	go monitor.AutoMonitor(cmdCh, feedbackCh)

	cmdCh <- "Auto"

	for {
		log.Println(<- feedbackCh)
	}


	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)
}

func oldmain() {
	tools.PreCheck(tools.MainPort, protocols)

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)



	//Get subscription links
	var subNLs tools.NodeLists
	tools.GetAllNodes(&subNLs)

	var allNodes []*tools.Node
	allNodes = append(subNLs.Vms, subNLs.Sses...)

	log.Println("Subs get done!")



	//Ping Tests
	goodPingNodes, badPingNodes, _, pingTime, _ := ping.XrayPing(allNodes)


	for i, n := range goodPingNodes {
		fmt.Println(i, n.AvgDelay)
	}
	fmt.Println("good:", len(goodPingNodes), "bad:", len(badPingNodes))

	os.Exit(0)


	//Generate halfGoodNodes
	generatePingOutFile(goodPingNodes)

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)


	//Speed Tests
	goodSpeedNodes, timeOfSpeed, _ := speedtest.XraySpeedTest(goodPingNodes)


	//Sort Nodes
	sort.Stable(tools.ByDelay(goodSpeedNodes))
	//sort.Sort(tools.ByULSpeed(goodSpeedNodes))
	sort.Stable(tools.ByDLSpeed(goodSpeedNodes))


	//Generate speedOut.txt
	generateSpeedOutFile(goodSpeedNodes)


	//Time Counting
	log.Println("-------------------")
	log.Println("Ping spent", pingTime, "s...")
	log.Println("Speedtest spent", timeOfSpeed, "s...")



	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)
}

func generatePingOutFile(nodes []*tools.Node) {
	var link []string
	for i, n := range nodes {
		n.CreateFinalJson(tools.HalfJsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", n.Type, "://", n.ShareLink, "\nDelay:", strconv.Itoa(n.AvgDelay)}
		vmOutStr := strings.Join(str, "")
		link = append(link, vmOutStr)
	}
	if len(link) != 0 {
		bytes := []byte(strings.Join(link[:], "\n"))
		err := os.WriteFile(tools.PingOutPath, bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("pingOut generated!")
		}
	}
}

func generateSpeedOutFile(nodes []*tools.Node) {
	var goodVmLinks []string
	for i, n := range nodes {
		fmt.Println(i, n.Type, (*n).AvgDelay, (*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
		//(*n).Id = strconv.Itoa(i)
		n.CreateFinalJson(tools.JsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", n.Type, "://", (*n).ShareLink,
				"\nDelay:", strconv.Itoa((*n).AvgDelay),
				" Down: ", fmt.Sprintf("%.2f", (*n).DLSpeed),
				" Up: ", fmt.Sprintf("%.2f", (*n).ULSpeed),
				" Country: ", (*n).Country, "\n"}
		vmOutStr := strings.Join(str, "")
		goodVmLinks = append(goodVmLinks, vmOutStr)
	}

	if len(goodVmLinks) != 0 {
		bytes := []byte(strings.Join(goodVmLinks[:], "\n"))
		err := os.WriteFile(tools.SpeedOutPath, bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("vmOut generated!")
		}
	}

}
