package speedtest

import (
	"sync"
	"log"
	"math"

	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node) {
	for node := range jobs {
		log.Println("Speed: start testing!")
		var x tools.Xray
		var fail int
		x.Init((*node).Port, (*node).JsonPath)
		x.Run()

		myClient := tools.HttpClientGet(x.Port, tools.STimeout)

		START:
		user, err := FetchUserInfo(myClient)
		if err != nil {
			log.Println("[ERROR]", "Fetch user info:", err)
			fail += 1
			if fail >= 3 {
				x.Stop()
				wg.Done()
				continue
			}else{
				goto START
			}
		}

		serverList, err := FetchServerList(user, myClient)
		if err != nil {
			log.Println("[ERROR]", "Fetch server list:", err)
			fail += 1
			if fail >= 3 {
				x.Stop()
				wg.Done()
				continue
			}else{
				goto START
			}
		}
		targets, err := serverList.FindServer([]int{})
		if err != nil {
			log.Println("[ERROR]", "Find server:", err)
			fail += 1
			if fail >= 3 {
				x.Stop()
				wg.Done()
				continue
			}else{
				goto START
			}
		}

		for _, s := range targets {
			if s.Country == "China" || s.Country == "Hong Kong"{
				break
			}
			s.PingTest(myClient)
			s.DownloadTest(true, myClient)
			if s.DLSpeed < tools.DSLine {
				log.Println("DownSpeed too slow, skipped.")
				break
			}
			s.UploadTest(true, myClient)

			(*node).Country = s.Country
			(*node).DLSpeed = math.Round(s.DLSpeed*100)/100
			(*node).ULSpeed = math.Round(s.ULSpeed*100)/100
			result <- node
			log.Println("Speed got one !")
		}

		x.Stop()
		wg.Done()
	}
}
