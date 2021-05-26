package speedtest

import (
	"net/http"
	"strings"
	"net/url"
	"time"
	"strconv"
	"sync"
	"log"
	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, timeout int) {
	for node := range jobs {
		var x tools.Xray
		x.Init(8123, node.JsonPath)
		x.Run(true)

		var t time.Duration = time.Duration(timeout) * time.Millisecond
		str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

		user, _ := FetchUserInfo(myClient)

		serverList, _ := FetchServerList(user, myClient)
		targets, _ := serverList.FindServer([]int{})

		for _, s := range targets {
			//s.PingTest(myClient)
			log.Println("start Download testing...")
			s.DownloadTest(false, myClient)
			log.Println("Download testing finished")
			log.Println("start Upload testing...")
			s.UploadTest(false, myClient)
			log.Println("Upload testing finished")

			//x.Stop()
			//return s.Country, s.DLSpeed, s.ULSpeed

			(*node).Country = s.Country
			(*node).DLSpeed = s.DLSpeed
			(*node).ULSpeed = s.ULSpeed
			result <- node
		}

		x.Stop()
		wg.Done()
		//return "", 0.0, 0.0
	}
}
