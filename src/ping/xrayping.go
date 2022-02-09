package ping

import (
	"log"
	"fmt"
	"time"

	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/xray"
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

	for i := 1; i <= threadNum; i++ {
		go myPing(pingJob, pingResult)
//		time.Sleep(time.Millisecond * 20)
	}


	//Put nodesIn into pingJob
	for _, n := range nodesIn {
		pingJob <- n
	}

	close(pingJob)


	log.Println("length of all", allLen)
	log.Println("waiting for goroutine")

	start := time.Now()
	var goodPingNodes, badPingNodes, errorNodes []*tools.Node
	for i:=0; i< allLen; i++ {
		n := <- pingResult
		if n.AvgDelay == 9999 {
			n.Timeout += 1
			badPingNodes = append(badPingNodes, n)
		}else if n.AvgDelay == -1 {
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
	log.Println("length of good", len(goodPingNodes))
	
	return goodPingNodes, badPingNodes, errorNodes, timeOfPing, nil 
}

func myPing(jobs <-chan *tools.Node, result chan<- *tools.Node) {
	for n := range jobs {
		var good int = 0
		var fail int = 0
		var stat bool = false

		server, err := xray.StartXray(n.Type, n.ShareLink, false, true)
		if err != nil {
			n.AvgDelay = -1
			result <- n
			err = fmt.Errorf("xray.StartXray():", err)
			n.ErrorInfo = err
			continue
		}

		if err := server.Start(); err != nil {
			n.AvgDelay = -1
			result <- n
			err = fmt.Errorf("server.Start():", err)
			n.ErrorInfo = err
			continue
		}
//		time.Sleep(time.Millisecond * 100)
//		defer server.Close()

//		for i := 0; i < tools.PCnt; i++ {
////			_, code, _, err := doPing(pClient, "https://www.google.com/gen_204", nil, false)
//			_, err := xray.MeasureDelay(server, tools.PTimeout, "https://www.google.com/gen_204")
//			if err == nil {
//				good += 1
//				if good >= tools.PingLeastGood {
//					stat = true
//					break
//				}
//			}
//			time.Sleep(time.Millisecond * 50)
//		}
		for {
			_, err := xray.MeasureDelay(server, time.Millisecond * 5000, "https://duckduckgo.com")
			if err == nil {
				good += 1
				if good >= tools.PingLeastGood {
					stat = true
					break
				}
			}else{
				fail += 1
				if fail >= tools.PingLeastGood + 2 {
					break
				}
			}
//			time.Sleep(time.Millisecond * 20)
		}

		if stat {
			var pRealAvgDelay int
			var pRealDelayList []int

			for i := 0; i < tools.PRealCnt; i++{
//				delay, _, _, err := doPing(pRealClient, "https://duckduckgo.com", nil, true)
				delay, err := xray.MeasureDelay(server, tools.PRealTimeout, "https://duckduckgo.com")
				if err == nil {
//					log.Println("delay:", delay)
					pRealDelayList = append(pRealDelayList, delay)
				}
//				time.Sleep(time.Millisecond * 20)
			}
			if len(pRealDelayList) >= tools.PRealLeastGood {
				pRealAvgDelay = getAvg(pRealDelayList)
				n.AvgDelay = pRealAvgDelay
				log.Println("ping got one!", n.Type, "delay:", pRealAvgDelay)
			}else{
				n.AvgDelay = 9999
			}
		}else{
			n.AvgDelay = 9999
		}

		result <- n
		server.Close()
	}
}

//func doPing(url string, timeout time.Duration) (int, int, []*http.Cookie, error){
//	var coo []*http.Cookie
//
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		err = fmt.Errorf("http.NewRequest:", err)
//		return 0, 0, nil, err
//	}
////	if pReal {
////		for i := range cookies {
////			req.AddCookie(cookies[i])
////		}
//		req.Close = true
////		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36")
////	}
//
//	start := time.Now()
//	resp, err := myClient.Do(req) //send request
//	stop := time.Now()
//
//	if err != nil {
//		return 0, 0, nil, err
//	}
//	defer resp.Body.Close()
//	code := resp.StatusCode
//
//	if code >= 399 && code != 429{
//	//if code >= 399 {
//		return 0, code, nil, errors.New("Ping err: StatusCode error")
//	}
//
//	elapsed := stop.Sub(start)
//	delay := elapsed.Milliseconds() / 2
//	if pReal {
//		coo = resp.Cookies()
//		return int(delay), code, coo, nil
//	}else{
//		return int(delay), code, nil, nil
//	}
//}

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
