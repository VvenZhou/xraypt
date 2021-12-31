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
)



func main() {
	tools.PreCheck(8123)

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)



	//Get subscription links
	var subLs tools.Links

	tools.GetSubLinks(&subLs)
	vmLinks := subLs.Vms
	ssLinks := subLs.Sses

	log.Println("Subs get done!")



	//Ping Tests
	vmessGoodPingNodes, vmessPingTime, _ := ping.XrayPing("vmess", vmLinks) 
	ssGoodPingNodes, ssPingTime, _ := ping.XrayPing("ss", ssLinks) 

	timeOfPing := vmessPingTime + ssPingTime
	var allGoodPingNodes []*tools.Node
	allGoodPingNodes = append(allGoodPingNodes, vmessGoodPingNodes...)
	allGoodPingNodes = append(allGoodPingNodes, ssGoodPingNodes...)

	//goodPingCnt := len(allGoodPingNodes)
	for i, n := range allGoodPingNodes {
		fmt.Println(i, n.AvgDelay)
	}


	//Generate halfGoodNodes
	generatePingOutFile(allGoodPingNodes)

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)


	//Speed Tests
	allGoodSpeedNodes, timeOfSpeed, _ := speedtest.XraySpeedTest(allGoodPingNodes)


	//Sort Nodes
	sort.Stable(tools.ByDelay(allGoodSpeedNodes))
	//sort.Sort(tools.ByULSpeed(allGoodSpeedNodes))
	sort.Stable(tools.ByDLSpeed(allGoodSpeedNodes))


	//Generate speedOut.txt
	generateSpeedOutFile(allGoodSpeedNodes)


	//Time Counting
	log.Println("-------------------")
	log.Println("Ping spent", timeOfPing, "s...")
	log.Println("Speedtest spent", timeOfSpeed, "s...")



	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)
	//os.RemoveAll(tools.HalfJsonsPath)
	//os.MkdirAll(tools.HalfJsonsPath, 0755)
}

func generatePingOutFile(nodes []*tools.Node) {
	var halfGoodVmLinks []string
	for i, n := range nodes {
		n.CreateFinalJson(tools.HalfJsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", n.Type, "://", n.ShareLink, "\nDelay:", strconv.Itoa(n.AvgDelay)}
		vmOutStr := strings.Join(str, "")
		halfGoodVmLinks = append(halfGoodVmLinks, vmOutStr)
	}
	if len(halfGoodVmLinks) != 0 {
		bytes := []byte(strings.Join(halfGoodVmLinks[:], "\n"))
		err := os.WriteFile("pingOut.txt", bytes, 0644)
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
		err := os.WriteFile("speedOut.txt", bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("vmOut generated!")
		}
	}

}
