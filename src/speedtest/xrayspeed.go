package speedtest

import (
	"sync"
	"log"
	"math"
	"time"
	"context"

	"github.com/VvenZhou/xraypt/src/tools"
)


func XraySpeedTest(nodesIn []*tools.Node) ([]*tools.Node, float64, error) {
	//Do speedtest
	var nodesOut []*tools.Node
	var wgSpeed sync.WaitGroup

	allLen := len(nodesIn)
	speedJob := make(chan *tools.Node, allLen)
	speedResult := make(chan *tools.Node, allLen)

	var threadNum int
	if allLen < tools.SThreadNum {
		threadNum = allLen
	}else{
		threadNum = tools.SThreadNum
	}


	for _, n := range nodesIn {
		speedJob <- n
		wgSpeed.Add(1)
	}

	ports, err := tools.GetFreePorts(threadNum)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= threadNum; i++ {
		go mySpeedTest(&wgSpeed, speedJob, speedResult, ports[i-1])
		time.Sleep(time.Second * 3)
	}
	close(speedJob)

	start := time.Now()
	wgSpeed.Wait()
	stop := time.Now()
	elapsed := stop.Sub(start)
	timeOfSpeedTest := elapsed.Seconds()

	goodSpeeds := len(speedResult)
	for i := 1; i <= goodSpeeds; i++ {
		n := <-speedResult
		nodesOut = append(nodesOut, n)
	}

	return nodesOut, timeOfSpeedTest, nil
}

func mySpeedTest(wg *sync.WaitGroup, jobs <-chan *tools.Node, result chan<- *tools.Node, port int) {
	
	fixedPort := port
	client := tools.HttpClientGet(fixedPort, tools.STimeout)
	//M[os.Getpid()] = Client

	ctx := context.WithValue(context.Background(), "client", &client)
	//client := tools.HttpClientGet(fixedPort, tools.STimeout)

	for node := range jobs {
		log.Println("Speed: start testing!")

		//ctxWithCancel, cancel := context.WithCancel(ctx)
		doTest(wg, node, result, fixedPort, ctx)
		//cancel()
	}
}

func doTest(wg *sync.WaitGroup, node *tools.Node, result chan<- *tools.Node, port int, ctx context.Context){
	var x tools.Xray
	var fail int

	fixedPort := port
	node.Port = fixedPort
	node.CreateJson(tools.TempPath)

	x.Init(fixedPort, node.JsonPath)
	x.Run()
	defer wg.Done()
	defer x.Stop()

	START:
	user, err := FetchUserInfo(ctx)
	if err != nil {
		log.Println("[ERROR]", "Fetch user info:", err)
		fail += 1
		if fail >= 3 {
			return
		}else{
			time.Sleep(1 * time.Second)
			goto START
		}
	}
	fail = 0

	START_1:
	serverList, err := FetchServerList(user, ctx)
	if err != nil {
		log.Println("[ERROR]", "Fetch server list:", err)
		fail += 1
		if fail >= 3 {
			return
		}else{
			time.Sleep(1 * time.Second)
			goto START_1
		}
	}
	fail = 0

	START_2:
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		log.Println("[ERROR]", "Find server:", err)
		fail += 1
		if fail >= 3 {
			return
		}else{
			time.Sleep(1 * time.Second)
			goto START_2
		}
	}
	fail = 0

	for _, s := range targets {
		//if s.Country == "China" || s.Country == "Hong Kong"{
		//if s.Country == "China" {
		//	break
		//}
		err = s.PingTest(ctx)
		if err != nil {
			//log.Println(err)
		}
		//log.Println("s.Latency:", s.Latency)

		err = s.DownloadTestContext(ctx, true)
		if err != nil {
			//log.Println(err)
		}
		//if s.DLSpeed < tools.DSLine {
		//	if s.DLSpeed != 0 {
		//		log.Println("DownSpeed too slow, skipped:", s.DLSpeed)
		//	}
		//	break
		//}

		//s.UploadTestContext(ctx, false)
		//s.UploadTest(true, ctx)

		(*node).Country = s.Country
		(*node).DLSpeed = math.Round(s.DLSpeed*100)/100
		//(*node).ULSpeed = math.Round(s.ULSpeed*100)/100
		result <- node
		log.Println("Speed got one !")
	}
}
