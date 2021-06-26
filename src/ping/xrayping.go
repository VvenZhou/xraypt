package ping

import (
	"log"
	"time"
	"net/http"
	"errors"
	"sync"
	//"strconv"
	//"strings"
	//"net/url"

	"github.com/VvenZhou/xraypt/src/tools"
)

func XrayPing(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node) {
	for n := range jobs {
		port := n.Port

		var x tools.Xray
		err := x.Init(port, (*n).JsonPath)
		if err != nil {
			log.Fatal(err)
		}
		err = x.Run()
		if err != nil {
			log.Fatal(err)
		}

		//str := []string{"http://127.0.0.1", strconv.Itoa(port)}
		//proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		//pClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: tools.PTimeout}
		//pRealClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: tools.PRealTimeout}

		pClient := tools.HttpClientGet(x.Port, tools.PTimeout)
		pRealClient := tools.HttpClientGet(x.Port, tools.PRealTimeout)

		var cookie []*http.Cookie
		var fail int = 0

		for i := 0; i < tools.PCnt; i++ {
			_, code, coo, err := Ping(pClient, "https://www.google.com/gen_204", cookie, false)
			if err != nil {
				if code != 0 {
					log.Println("ping error: code:", code, err)
				}
				fail += 1
			}else{
				if len(coo) > 0 && len(cookie) == 0 {
					cookie = coo
				}
				//log.Println("ping gen succ", cookies)
			}
		}

		if fail <= 3 {
			var pRealTotalDelay int
			var pRealAvgDelay int
			var pRealCnt int

			fail = 0
			for i := 0; i < tools.PRealCnt; i++{
				//delay, code, coo, err := Ping(pRealClient, "https://www.google.com", cookie, true)
				delay, code, coo, err := Ping(pRealClient, "https://duckduckgo.com", cookie, true)
				//if err != nil && code != 429{
				if err != nil {
					fail += 1
					log.Println("PingReal error: code:", code, err)
					//time.Sleep(50*time.Millisecond)
				}else{
					if len(coo) != 0 && len(cookie) == 0 {
						cookie = coo
					}
					//log.Println(delay)
					pRealTotalDelay += delay
					pRealCnt += 1
					if pRealCnt >= 4 {
						break
					}
				}
			}
			//if pRealTotalDelay == 0 {
			//	goto END
			//}
			if fail >= 3 {
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

func Ping(myClient *http.Client, url string, cookies []*http.Cookie, pReal bool) (int, int, []*http.Cookie, error){
	var coo []*http.Cookie

	req, _ := http.NewRequest("GET", url, nil)
	req.Close = true
	if pReal {
		for i := range cookies {
			req.AddCookie(cookies[i])
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0i")
	}
	//req := tools.HttpNewRequest(url, cookies)

	start := time.Now()
	resp, err := myClient.Do(req) //send request
	stop := time.Now()
	if err != nil {
		return 0, 0, nil, err
	}
	defer resp.Body.Close()
	code := resp.StatusCode

	//if code >= 399 && code != 429{
	if code >= 399 {
		//if code != 503 {
			//log.Println("code is", code, "instead of 204,")
		//}
		return 0, code, nil, errors.New("Ping err: StatusCode error")
	}


	coo = resp.Cookies()

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()/2
	return int(delay), code, coo, nil
}
