package speedtest

import (
	"net/http"
	"strings"
	"net/url"
	"time"
	"strconv"
	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(jsonPath string, timeout int) (string, float64, float64) {
	var x tools.Xray
	x.Init(jsonPath)
	x.Run()

	var t time.Duration = time.Duration(timeout) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

	user, _ := FetchUserInfo(myClient)

	serverList, _ := FetchServerList(user, myClient)
	targets, _ := serverList.FindServer([]int{})

	for _, s := range targets {
		//s.PingTest(myClient)
		s.DownloadTest(false, myClient)
		s.UploadTest(false, myClient)

		x.Stop()

		return s.Country, s.DLSpeed, s.ULSpeed
	}
	return "", 0.0, 0.0
}
