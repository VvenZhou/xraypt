package speed

import (
	"fmt"
	"github.com/VvenZhou/xraypt"
)

func main() {
	const port = 8123
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	user, _ := speedtest.FetchUserInfo(myClient)

	serverList, _ := speedtest.FetchServerList(user, myClient)
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		s.PingTest(myClient)
		s.DownloadTest(false, myClient)
		s.UploadTest(false, myClient)

		fmt.Printf("Latency: %s, Download: %f, Upload: %f\n", s.Latency, s.DLSpeed, s.ULSpeed)
	}
}
