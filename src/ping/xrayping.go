package ping

import (
	"log"
	"time"
	"net/http"
	"errors"
	"sync"

	"github.com/VvenZhou/xraypt/src/tools"
)

func XrayPing(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node) {
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

		pClient := tools.HttpClientGet(x.Port, tools.PTimeout)
		pRealClient := tools.HttpClientGet(x.Port, tools.PRealTimeout)

		var cookies []*http.Cookie

		for i := 0; i < tools.PCnt; i++ {
			_, _, coo,  err := Ping(pClient, "https://www.google.com/gen_204", cookies)
			if err != nil {
				fail += 1
			}else{
				cookies = coo
				//log.Println("ping gen succ", cookies)
			}
		}

		if !(fail >= (tools.PCnt+1)/2) {
			var pRealTotalDelay int
			var pRealAvgDelay int
			var pRealCnt int

			for i := 0; i < tools.PRealCnt; i++{
				delay, code, _, err := Ping(pRealClient, "https://www.google.com", cookies)
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

func Ping(myClient *http.Client, url string, cookies []*http.Cookie) (int, int, []*http.Cookie, error){
	var coo []*http.Cookie

	req, _ := http.NewRequest("GET", url, nil)
	for i := range cookies {
		req.AddCookie(cookies[i])
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0i")
	start := time.Now()
	resp, err := myClient.Do(req) //send request
	stop := time.Now()
	if err != nil {
		return 0, 0, coo, err
	}
	code := resp.StatusCode

	defer resp.Body.Close()
	if code >= 399 && code != 429{
		//if code != 503 {
			//log.Println("code is", code, "instead of 204,")
		//}
		return 0, code, coo, errors.New("Ping err: StatusCode is not 204 or 429")
	}

	coo = resp.Cookies()

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()/2
	return int(delay), code, coo, nil
}
