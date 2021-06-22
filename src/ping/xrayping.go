package ping

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"time"
	"errors"
	"sync"

	"github.com/VvenZhou/xraypt/src/tools"
)

func XrayPing(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, pCount int, pTimeout time.Duration, pRealCount int, pRealTimeout time.Duration) {
	for n := range jobs {
		var fail int = 0

		var x tools.Xray
		err := x.Init((*n).Port, (*n).JsonPath)
		if err != nil {
			log.Fatal(err)
		}
		err = x.Run()
		if err != nil {
			log.Fatal(err)
		}

		str := []string{"http://127.0.0.1", strconv.Itoa(x.Port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		pClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: pTimeout}
		pRealClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: pRealTimeout}

		var cookies []*http.Cookie

		for i := 0; i < pCount; i++ {
			_, _, err := Ping(pClient, "https://www.google.com/gen_204", cookies)
			if err != nil {
				fail += 1
			}
		}

		if !(fail >= (pCount+1)/2) {
			var pRealTotalDelay int
			var pRealAvgDelay int
			var pRealCnt int
			var cookies []*http.Cookie

			for cnt := 1; cnt <= 3; cnt += 1{
				resp, err := pClient.Get("https://www.google.com")
				if err != nil{
					log.Println("get cookies error:", err)
					continue
				}
				cookies = resp.Cookies()
				break
			}
			//if len(cookies) == 0 {
			//	goto END
			//}

			for i := 0; i < pRealCount; i++{
				delay, code, err := Ping(pRealClient, "https://www.google.com", cookies)
				if err != nil && code != 429{
					log.Println("PingReal fail", i + 1, "times,", "error:", err)
					//time.Sleep(50*time.Millisecond)
				}else{
					//log.Println(delay)
					pRealTotalDelay += delay
					pRealCnt += 1
				}
			}
			if pRealTotalDelay == 0 {
				goto END
			}
			pRealAvgDelay = pRealTotalDelay / pRealCnt
			(*n).AvgDelay = pRealAvgDelay
			log.Println("ping got one!", "delay:", pRealAvgDelay)
			result <- n
		}
		END:
		  err = x.Stop()
		  if err != nil {
			log.Fatal(err)
		  }
		  wg.Done()
	}
}

func Ping(myClient *http.Client, url string, cookies []*http.Cookie) (int, int, error){
	req, _ := http.NewRequest("GET", url, nil)
	for i := range cookies {
		req.AddCookie(cookies[i])
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0i")
	start := time.Now()
	resp, err := myClient.Do(req) //send request
	stop := time.Now()
	if err != nil {
		return 0, 0, err
	}
	//start := time.Now()
	//resp, err := myClient.Get(url)
	//stop := time.Now()
	//if err != nil {
	//	return 0, 0, err
	//}
	code := resp.StatusCode

	defer resp.Body.Close()
	if code >= 399 && code != 429{
		//if code != 503 {
			//log.Println("code is", code, "instead of 204,")
		//}
		return 0, code,  errors.New("Ping err: StatusCode is not 204 or 429")
	}

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()/2
	return int(delay), code,  nil
}
