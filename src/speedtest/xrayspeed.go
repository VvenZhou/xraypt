package speedtest

import (
	"sync"
	"log"
	"math"
	"time"

	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node) {
	for node := range jobs {
		log.Println("Speed: start testing!")
		doTest(wg, node, result)
	}
}

func doTest(wg *sync.WaitGroup, node *tools.Node, result chan<-*tools.Node){
	var x tools.Xray
	var fail int
	x.Init((*node).Port, (*node).JsonPath)
	x.Run()
	defer wg.Done()
	defer x.Stop()

	myClient := tools.HttpClientGet(x.Port, tools.STimeout)

	START:
	user, err := FetchUserInfo(myClient)
	if err != nil {
		log.Println("[ERROR]", "Fetch user info:", err)
		fail += 1
		if fail >= 3 {
			return
		}else{
			time.Sleep(3 * time.Second)
			goto START
		}
	}
	fail = 0

	START_1:
	serverList, err := FetchServerList(user, myClient)
	if err != nil {
		log.Println("[ERROR]", "Fetch server list:", err)
		fail += 1
		if fail >= 3 {
			//x.Stop()
			//wg.Done()
			return
		}else{
			time.Sleep(5 * time.Second)
			goto START_1
		}
	}
	fail = 0

	START_2:
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		log.Println("[ERROR]", "Find server:", err)
		fail += 1
		if fail >= 3 {
			return
		}else{
			time.Sleep(5 * time.Second)
			goto START_2
		}
	}
	fail = 0

	for _, s := range targets {
		//if s.Country == "China" || s.Country == "Hong Kong"{
		if s.Country == "China" {
			break
		}
		s.PingTest(myClient)
		s.DownloadTest(false, myClient)
		if s.DLSpeed < tools.DSLine {
			log.Println("DownSpeed too slow, skipped.")
			break
		}
		s.UploadTest(false, myClient)

		(*node).Country = s.Country
		(*node).DLSpeed = math.Round(s.DLSpeed*100)/100
		(*node).ULSpeed = math.Round(s.ULSpeed*100)/100
		result <- node
		log.Println("Speed got one !")
	}
}
