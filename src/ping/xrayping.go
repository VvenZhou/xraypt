package ping

import (
	"log"
	"fmt"
	"time"
	"net/http"
	"errors"

	"github.com/VvenZhou/xraypt/src/tools"
)

func XrayPing(nodesIn []*tools.Node) ([]*tools.Node, []*tools.Node, []*tools.Node, float64, error){
	var threadNum int

	allLen := len(nodesIn)
	if allLen == 0 {
		return nil, nil, nil, 0, nil
	}
	pingJob := make(chan *tools.Node, allLen)
	pingResult := make(chan *tools.Node, allLen)

	if allLen < tools.PThreadNum {
		threadNum = allLen
	}else{
		threadNum = tools.PThreadNum
	}

	//Put nodesIn into pingJob
	for _, n := range nodesIn {
		pingJob <- n
	}

	//var ports []int
	ports, err := tools.GetFreePorts(threadNum)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= threadNum; i++ {
		go myPing(pingJob, pingResult, ports[i-1])
	}
	close(pingJob)


	log.Println("waiting for goroutine")
	start := time.Now()

	log.Println("length of all", allLen)
	var goodPingNodes, badPingNodes, errorNodes []*tools.Node
	for i:=0; i< allLen; i++ {
		n := <- pingResult
		if n.AvgDelay == -1 {
			n.Timeout += 1
			badPingNodes = append(badPingNodes, n)
		}else if n.AvgDelay == -2 {
			//log.Println("-2 error:", n.ErrorInfo)
			n.Timeout = -1
			errorNodes = append(errorNodes, n)
		}else{
			n.Timeout = 0
			goodPingNodes = append(goodPingNodes, n)
		}
	}
	stop := time.Now()
	elapsed := stop.Sub(start)
	timeOfPing := elapsed.Seconds()
	log.Println("goroutine finished")
	
	return goodPingNodes, badPingNodes, errorNodes, timeOfPing, nil 
}

func myPing(jobs <-chan *tools.Node, result chan<- *tools.Node, port int) {
	fixedPort := port
	pClient := tools.HttpClientGet(fixedPort, tools.PTimeout)
	pRealClient := tools.HttpClientGet(fixedPort, tools.PRealTimeout)

	for n := range jobs {
		var x tools.Xray

		n.Port = fixedPort
		err := n.CreateJson(tools.TempPath)
		if err != nil {
			n.AvgDelay = -2
			result <- n
			err = fmt.Errorf("n.CreateJson:", err)
			n.ErrorInfo = err
			continue
		}
		err = x.Init(fixedPort, (*n).JsonPath)
		if err != nil {
			n.AvgDelay = -2
			result <- n
			err = fmt.Errorf("x.Init:", err)
			n.ErrorInfo = err
			continue
		}
		_, err = x.Run()
		if err != nil {
			n.AvgDelay = -2
			result <- n
			err = fmt.Errorf("x.Run:", err)
			n.ErrorInfo = err
			continue
		}

		var cookie []*http.Cookie
		var fail int = 0

		for i := 0; i < tools.PCnt; i++ {
			_, code, _, err := doPing(pClient, "https://www.google.com/gen_204", nil, false)
			if err != nil {
				if code != 0 {
					//log.Println("ping error: code:", code, err)
				}
				fail += 1
			}
			time.Sleep(time.Millisecond * 200)
		}

		if fail <= tools.PingAllowFail {
			var pRealAvgDelay int
			var pRealDelayList []int

			for i := 0; i < tools.PRealCnt; i++{
				//delay, code, coo, err := Ping(pRealClient, "https://www.google.com/ncr", cookie, true)
				delay, _, coo, err := doPing(pRealClient, "https://duckduckgo.com", cookie, true)
				if err != nil {
				}else{
					if len(coo) != 0 {
						cookie = coo
					}
					pRealDelayList = append(pRealDelayList, delay)
				}
				time.Sleep(time.Millisecond * 200)
			}
			if len(pRealDelayList) >= tools.PRealLeastNeeded {
				pRealAvgDelay = getAvg(pRealDelayList)
				n.AvgDelay = pRealAvgDelay
				log.Println("ping got one!", n.Type, "delay:", pRealAvgDelay)
				result <- n
			}else{
				n.AvgDelay = -1
				result <- n
			}
		}else{
			n.AvgDelay = -1
			result <- n
		}

		err = x.Stop()
		if err != nil {
			n.AvgDelay = -2
			result <- n
			err = fmt.Errorf("x.Stop:", err)
			n.ErrorInfo = err
			continue
			//log.Fatal(err)
		}
	}
}

func doPing(myClient http.Client, url string, cookies []*http.Cookie, pReal bool) (int, int, []*http.Cookie, error){
	var coo []*http.Cookie

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("http.NewRequest:", err)
		return 0, 0, nil, err
	}
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
