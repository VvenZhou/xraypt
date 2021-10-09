package speedtest

import (
	"sync"
	"log"
	"math"
	"time"
	"net/http"
	"os"

	"github.com/VvenZhou/xraypt/src/tools"
)

//var Client http.Client
var M = make(map[int]http.Client)

func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, port int) {
	
	fixedPort := port
	Client := tools.HttpClientGet(fixedPort, tools.STimeout)
	M[os.Getpid()] = Client

	for node := range jobs {
		log.Println("Speed: start testing!")
		doTest(wg, node, result, fixedPort)
	}
}

func doTest(wg *sync.WaitGroup, node *tools.Node, result chan<- *tools.Node, port int){
	var x tools.Xray
	var fail int

	fixedPort := port
	node.Port = fixedPort
	node.CreateJson(tools.TempPath)

	x.Init(fixedPort, node.JsonPath)
	x.Run()
	defer wg.Done()
	defer x.Stop()

	START:
	user, err := FetchUserInfo()
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
	serverList, err := FetchServerList(user)
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
		s.PingTest()
		s.DownloadTest(true)
		if s.DLSpeed < tools.DSLine {
			log.Println("DownSpeed too slow, skipped.")
			break
		}
		s.UploadTest(true)

		(*node).Country = s.Country
		(*node).DLSpeed = math.Round(s.DLSpeed*100)/100
		(*node).ULSpeed = math.Round(s.ULSpeed*100)/100
		result <- node
		log.Println("Speed got one !")
	}
}
