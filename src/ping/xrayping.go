package ping

import (
	"log"
	"fmt"
	"time"
	"net/http"
	"errors"
	"context"

	"github.com/VvenZhou/xraypt/src/tools"
//	"github.com/VvenZhou/xraypt/src/xray"
)

type worker struct {
	id int
	err error
}

func (wk *worker)work(ctx context.Context,
			jobs <-chan *tools.Node,
			result chan<- *tools.Node,
			workerCh chan<- *worker,
			port int,
			url string) (err error){
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				wk.err = err
			}else{
				wk.err = fmt.Errorf("Panic happened with %v", r)
			}
		}else{
			wk.err = err
		}
		workerCh <- wk
	}()

	err = myPing(ctx, jobs, result, port, "https://duckduckgo.com")
	return err
}

func XrayPing(ctx context.Context, nodesIn []*tools.Node) ([]*tools.Node, []*tools.Node, []*tools.Node, float64, error){
	var threadNum int
	var done bool
	done = false

	allLen := len(nodesIn)
	if allLen == 0 {
		return nil, nil, nil, 0, nil
	}
	pingJob := make(chan *tools.Node, allLen)
	pingResult := make(chan *tools.Node)

	if allLen < tools.PThreadNum {
		threadNum = allLen
	}else{
		threadNum = tools.PThreadNum
	}


	//TODO: get free ports
	ports, err := tools.GetFreePorts(threadNum)
	if err != nil {
		return nil, nil, nil, 0, fmt.Errorf("GetFreePorts:%w", err)
	}

//	pingCtx, cancel := context.WithCancel(ctx)
//	defer cancel()
	workerCh := make(chan *worker, threadNum)
	for i, port := range(ports) {
//		go myPing(ctx, pingJob, pingResult, port, "https://duckduckgo.com")
		wk := &worker{id: i}
		go wk.work(ctx, pingJob, pingResult, workerCh, port, "https://duckduckgo.com")
//		time.Sleep(time.Millisecond * 20)
	}


	//Put nodesIn into pingJob
	for _, n := range nodesIn {
		pingJob <- n
	}

	close(pingJob)


	log.Println("length of all", allLen)
	log.Println("Ping start")

	start := time.Now()
	var goodPingNodes, badPingNodes, errorNodes []*tools.Node
	go func() {
		log.Println("Worker watcher start...")
		workerCnt := 0
		for wk := range(workerCh){
			workerCnt += 1
			if errors.Is(wk.err, context.Canceled) && workerCnt == threadNum {
				done = true
			}
			if workerCnt == threadNum {
				if done {
					log.Println("Context canceled")
				}
				close(pingResult)
				close(workerCh)
				log.Println("...Worker watcher quit")
				return
			}
		}
	}()
//	for i:=0; i< allLen; i++ {
//		n := <- pingResult:
	for n := range(pingResult) {
		if n.AvgDelay == 9999 {
			n.Timeout += 1
			badPingNodes = append(badPingNodes, n)
		}else if n.AvgDelay == -1 {
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
	log.Println("Ping finished")
	log.Println("length of good", len(goodPingNodes))
	
	if done {
		return goodPingNodes, badPingNodes, errorNodes, timeOfPing, tools.UsrIntErr
	}else{
		return goodPingNodes, badPingNodes, errorNodes, timeOfPing, nil 
	}
}

//func myPing(jobs <-chan *tools.Node, result chan<- *tools.Node) {
func myPing(ctx context.Context, jobs <-chan *tools.Node, result chan<- *tools.Node, port int, url string) error {
	pClient := tools.HttpClientGet(port, tools.PTimeout)
	pRClient := tools.HttpClientGet(port, tools.PRealTimeout)
	req := tools.HttpNewRequest("GET", url)
	req = req.WithContext(ctx)
	for n := range jobs {
		var good int = 0
//		var fail int = 0
		var stat bool = false
		var x tools.Xray

		err := n.CreateJson(tools.TempPath, port)
		if err != nil {
			n.AvgDelay = -1
			result <- n
			err = fmt.Errorf("CreateJson:%w", err)
			n.ErrorInfo = err
			continue
		}

		x.Init(port, n.JsonPath)

		_, err = x.Run()
		if err != nil {
			n.AvgDelay = -1
			result <- n
			err = fmt.Errorf("x.Run:%w", err)
			n.ErrorInfo = err
			continue
		}

//		server, err := xray.StartXray(n.Type, n.ShareLink, false, true)
//		if err != nil {
//			n.AvgDelay = -1
//			result <- n
//			err = fmt.Errorf("xray.StartXray():", err)
//			n.ErrorInfo = err
//			continue
//		}
//
//		if err := server.Start(); err != nil {
//			n.AvgDelay = -1
//			result <- n
//			err = fmt.Errorf("server.Start():", err)
//			n.ErrorInfo = err
//			continue
//		}

		for i:=0; i< tools.PCnt; i++ {
//			_, err := xray.MeasureDelay(server, time.Millisecond * 5000, "https://www.google.com/gen_204")
//			_, err := xray.MeasureDelay(server, time.Millisecond * 5000, "https://duckduckgo.com")
			_, err := doPing(&pClient, req)
			if err == nil {
				good += 1
				if good >= tools.PLeastGood {
					stat = true
					break
				}
			}else{
				if errors.Is(err, context.Canceled) {
					x.Stop()
					return err
				}
			}
			time.Sleep(time.Millisecond * 10)
		}

		if stat {
			var pRealAvgDelay int
			var pRealDelayList []int

			for i := 0; i < tools.PRealCnt; i++{
//				delay, err := xray.MeasureDelay(server, tools.PRealTimeout, "https://duckduckgo.com")
//				delay, err := xray.MeasureDelay(server, tools.PRealTimeout, "https://www.google.com/gen_204")
				delay, err := doPing(&pRClient, req)
				if err == nil {
//					log.Println("delay:", delay)
					pRealDelayList = append(pRealDelayList, delay)
				}else{
					if errors.Is(err, context.Canceled) {
						x.Stop()
						return err
					}
				}
				time.Sleep(time.Millisecond * 10)
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


//		server.Close()
		err = x.Stop()
		if err != nil {
			return err
		}
	}

	return nil
}

func doPing(myClient *http.Client, req *http.Request) (int, error){

	start := time.Now()
	resp, err := myClient.Do(req) //send request
	stop := time.Now()
	if err != nil {
		return 0, fmt.Errorf("Do:%w", err)
	}
	defer resp.Body.Close()

	code := resp.StatusCode

//	if code >= 399 && code != 429{
	if code >= 399 {
		return 0, errors.New("doPing: StatusCode error")
	}

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()

	return int(delay/2), nil
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
