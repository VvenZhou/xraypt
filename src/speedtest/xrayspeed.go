package speedtest

import (
	"net/http"
	"strings"
	"net/url"
	"time"
	"strconv"
	"sync"
	"log"
	"math"

	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, timeout time.Duration, DSLine float64) {
	for node := range jobs {
		log.Println("Speed: start testing!")
		var x tools.Xray
		x.Init((*node).Port, (*node).JsonPath)
		x.Run()

		str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: timeout}

		user, err := FetchUserInfo(myClient)
		if err != nil {
			x.Stop()
			wg.Done()
			log.Println("[ERROR]", "Fetch user info:", err)
			continue
		}

		serverList, err := FetchServerList(user, myClient)
		if err != nil {
			x.Stop()
			wg.Done()
			log.Println("[ERROR]", "Fetch server list:", err)
			continue
		}
		targets, err := serverList.FindServer([]int{})
		if err != nil {
			x.Stop()
			wg.Done()
			log.Println("[ERROR]", "Find server:", err)
			continue
		}

		for _, s := range targets {
			if s.Country == "China"{
				break
			}
			s.PingTest(myClient)
			//log.Println(s.Latency.Milliseconds())
			//if s.Country == "China" {
			//	log.Println("Speed Skipped for China.")
			//	break
			//}
			s.DownloadTest(true, myClient)
			//if s.DLSpeed < DSLine {
			//	log.Println("DownSpeed too slow, skipped.")
			//	break
			//}
			s.UploadTest(true, myClient)
			if s.DLSpeed >= 10 {
				s := []string{ s.Country, "_", strconv.Itoa(int(s.Latency.Milliseconds())), "_", strconv.FormatFloat(s.DLSpeed, 'f', 4, 64)}
				name := strings.Join(s, "")
				(*node).CreateFinalJson(tools.JitPath, name)
			}

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
