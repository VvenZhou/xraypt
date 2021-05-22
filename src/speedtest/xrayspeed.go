package speedtest

import (
	"fmt"
	"http"
	"net/url"
	"strconv"
	"github.com/VvenZhou/xraypt/src/speedtest"
	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeed() {
	port, _ := tools.GetFreePort()

	cmd := tools.RunXray(jsonPath)

	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	user, _ := speedtest.FetchUserInfo(myClient)

	serverList, _ := speedtest.FetchServerList(user, myClient)
	targets, _ := speedtest.serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest(myClient)
		s.DownloadTest(false, myClient)
		s.UploadTest(false, myClient)

		fmt.Printf("Latency: %s, Download: %f, Upload: %f\n", s.Latency, s.DLSpeed, s.ULSpeed)
	}
}
