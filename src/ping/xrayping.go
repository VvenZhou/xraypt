package ping

import (
	"log"
	"time"
	"net/http"
	"errors"
	"sync"

	"github.com/VvenZhou/xraypt/src/tools"
)

func XrayPing(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, port int) {
	fixedPort := port
	pClient := tools.HttpClientGet(fixedPort, tools.PTimeout)
	pRealClient := tools.HttpClientGet(fixedPort, tools.PRealTimeout)

	for n := range jobs {
		var x tools.Xray

		n.Port = fixedPort
		n.CreateJson(tools.TempPath)
		err := x.Init(fixedPort, (*n).JsonPath)
		if err != nil {
			log.Fatal(err)
		}
		err = x.Run()
		if err != nil {
			log.Fatal(err)
		}

		var cookie []*http.Cookie
		var fail int = 0

		for i := 0; i < tools.PCnt; i++ {
			_, code, _, err := Ping(pClient, "https://www.google.com/gen_204", nil, false)
			time.Sleep(time.Millisecond * 200)
			if err != nil {
				if code != 0 {
					//log.Println("ping error: code:", code, err)
				}
				fail += 1
			}
		}

		if fail <= tools.PingAllowFail {
			//log.Println("P good")
			//var pRealTotalDelay int
			var pRealAvgDelay int
			//var success int
			var pRealDelayList []int

			for i := 0; i < tools.PRealCnt; i++{
				//delay, code, coo, err := Ping(pRealClient, "https://www.google.com/ncr", cookie, true)
				delay, _, coo, err := Ping(pRealClient, "https://duckduckgo.com", cookie, true)
				time.Sleep(time.Millisecond * 200)
				//if err != nil && code != 429{
				if err != nil {
					//log.Println("PingReal error: code:", code, err)
					//time.Sleep(50*time.Millisecond)
				}else{
					//if len(coo) != 0 && len(cookie) == 0 {
					if len(coo) != 0 {
						cookie = coo
					}
					//log.Println(delay)
					//pRealTotalDelay += delay
					pRealDelayList = append(pRealDelayList, delay)
					//success += 1
					//if success >= 5 {
					//	break
					//}
				}
			}
			//if (tools.PRealCnt-success) <= tools.PRealAllowFail {
			//	pRealAvgDelay = pRealTotalDelay / success
			//	n.AvgDelay = pRealAvgDelay
			//	log.Println("ping got one!", "delay:", pRealAvgDelay)
			//	result <- n
			//}
			if len(pRealDelayList) >= tools.PRealLeastNeeded {
				pRealAvgDelay = getAvg(pRealDelayList)
				n.AvgDelay = pRealAvgDelay
				log.Println("ping got one!", "delay:", pRealAvgDelay)
				result <- n
			}
		}

		err = x.Stop()
		if err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}
}

func Ping(myClient http.Client, url string, cookies []*http.Cookie, pReal bool) (int, int, []*http.Cookie, error){
	var coo []*http.Cookie

	req, _ := http.NewRequest("GET", url, nil)
	if pReal {
		for i := range cookies {
			req.AddCookie(cookies[i])
		}
		req.Close = true
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
	}

	start := time.Now()
	resp, err := myClient.Do(req) //send request
	stop := time.Now()
	if err != nil {
		return 0, 0, nil, err
	}
	defer resp.Body.Close()
	code := resp.StatusCode

	if code >= 399 && code != 429{
	//if code >= 399 {
		return 0, code, nil, errors.New("Ping err: StatusCode error")
	}

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds() / 2
	if pReal {
		coo = resp.Cookies()
		return int(delay), code, coo, nil
	}else{
		return int(delay), code, nil, nil
	}
}

func getAvg(list []int) int {
	var min, max, total int
	max = list[0]
	min = list[0]

	for _, i := range list {
		if i > max {
			max = i
		}
		if i < min {
			min = i
		}
		total += i
	}

	return int((total - max - min) / (len(list) - 2 ))
}
