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


func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, timeout int, DSLine float64) {
	for node := range jobs {
		var x tools.Xray
		x.Init((*node).Port, (*node).JsonPath)
		x.Run()

		var t time.Duration = time.Duration(timeout) * time.Millisecond
		str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

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
			//s.PingTest(myClient)
			//if s.Country == "China" {
			//	log.Println("Speed Skipped for China.")
			//	break
			//}
			s.DownloadTest(false, myClient)
			//if s.DLSpeed < DSLine {
			//	log.Println("DownSpeed too slow, skipped.")
			//	break
			//}
			s.UploadTest(false, myClient)
			log.Println("Speed got one!")

			(*node).Country = s.Country
			(*node).DLSpeed = math.Round(s.DLSpeed*100)/100
			(*node).ULSpeed = math.Round(s.ULSpeed*100)/100
			result <- node
		}

		x.Stop()
		wg.Done()
		//return "", 0.0, 0.0
	}
}
