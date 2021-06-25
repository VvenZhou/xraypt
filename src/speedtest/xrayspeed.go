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
		var fail int
		x.Init((*node).Port, (*node).JsonPath)
		x.Run()

		str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: timeout}

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
			if s.DLSpeed < DSLine {
				log.Println("DownSpeed too slow, skipped.")
				break
			}
			s.UploadTest(true, myClient)
			//if s.DLSpeed >= 10 {
			//	s := []string{ s.Country, "_", strconv.Itoa(int(s.Latency.Milliseconds())), "_", strconv.FormatFloat(s.DLSpeed, 'f', 4, 64)}
			//	name := strings.Join(s, "")
			//	(*node).CreateFinalJson(tools.JitPath, name)
			//}

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
