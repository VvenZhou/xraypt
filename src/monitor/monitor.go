package monitor

import(
	"log"
	"errors"
	"fmt"
	"sort"
	"time"
	"context"
	"sync"
	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/ping"
//	"github.com/VvenZhou/xraypt/src/speedtest"
)

type Bench struct {
	GoodNodes []*tools.Node
	BadNodes []*tools.Node
	PreLength int
	MidDelay int
}

func (b *Bench) Refresh(ctx context.Context) {
	goodPingNodes, badPingNodes, _, _, err := ping.XrayPing(ctx, append((*b).GoodNodes, (*b).BadNodes...))
	if err != nil {
		return fmt.Errorf("Bench.Refresh:%w", err)
	}
	sort.Stable(tools.ByDelay(goodPingNodes))

	l := len(b.GoodNodes)
	if l > 0 {
		if l%2 == 0 {
			b.MidDelay = b.GoodNodes[l/2 -1].AvgDelay
		}else{
			b.MidDelay = b.GoodNodes[(l-1)/2].AvgDelay
		}
	}else{
		b.MidDelay = 9999
	}

	b.GoodNodes = goodPingNodes
	b.BadNodes = append(b.BadNodes, badPingNodes...)
}

func (b *Bench) Clean() {
	var nodes []*tools.Node
	for _, node := range b.BadNodes {
		if node.Timeout >= tools.NodeTimeoutTolerance {
			BadNodesBuffer = append(BadNodesBuffer, node)
		}else{
			nodes = append(nodes, node)
		}
	}

	b.GoodNodes = append(b.GoodNodes, nodes...)
	b.BadNodes = nil
}

const benchSize = 8

var cmdToDaemonCh = make(chan string)
var feedbackFromDaemonCh = make(chan string)

var CurrentNode *tools.Node
var CurrentNodePos int
var FirstBench *Bench

var BadNodesBuffer []*tools.Node
var curStatus int
var preStatus int

var daemonRunning bool

func AutoMonitor(ctx context.Context, cmdCh <-chan string, feedbackCh chan<- bool, dataCh <-chan string) {
	log.Println("AutoMonitor Start")

	cmdToRoutineCh := make(chan bool)
	feedbackFromRoutineCh := make(chan bool)
	var mu sync.Mutex
	var ticker *time.Ticker


	ticker = time.NewTicker(tools.RoutinePeriodDu)
	ticker.Stop()
	go func() {
		for {
			select {
			case <-cmdToRoutineCh :
				feedbackFromRoutineCh <- true
				return
			case <-ticker.C :
				ticker.Stop()
				mu.Lock()
				routine(ctx)
				mu.Unlock()
				ticker.Reset(tools.RoutinePeriodDu)
			}
		}
	}()

	curStatus = 0
	preStatus = 0

	err := firstIn(ctx)
	if err != nil {
		log.Println(err)

		cmdToRoutineCh <- true
		log.Println("Routine quit:", <- feedbackFromRoutineCh)

		log.Println("Auto monitor quit")
		feedbackCh <- true		//ready
		return
	}

	feedbackCh <- true

	for {
		select {
		case cmd := <-cmdCh :
			switch cmd {
			case "refresh" :
				mu.Lock()
				if curStatus == 1 {
					ticker.Stop()
				}
				log.Println("Cmd: Refresh")

				var data []string
				for i:=0; i<2; i++{
					data = append(data, <-dataCh)
				}
				refresh(ctx, data)

				if curStatus == 1 {
					ticker.Reset(tools.RoutinePeriodDu)
				}
				mu.Unlock()
				feedbackCh <- true
			case "fetch" :
				mu.Lock()
				if curStatus == 1 {
					ticker.Stop()
				}
				log.Println("Cmd: FetchNew")

				if daemonRunning == false {
					log.Println("XrayDaemon is not running, you can't fetch anything.")
				}else{
					fetchNew(ctx)
				}

				if curStatus == 1 {
					ticker.Reset(tools.RoutinePeriodDu)
				}
				mu.Unlock()
				feedbackCh <- true
			case "pause" :
				mu.Lock()
				log.Println("Cmd: Pause")

				if daemonRunning {
					if curStatus == 1 {
						ticker.Stop()
					}
					xrayDaemonStartStop("stop")
				}else{
					if curStatus == 1 {
						ticker.Reset(tools.RoutinePeriodDu)
					}
					xrayDaemonStartStop("start")
				}

				mu.Unlock()
				feedbackCh <- true
			case "next" :
				mu.Lock()
				log.Println("Cmd: Next")

				if CurrentNodePos < len(FirstBench.GoodNodes) - 1 {
					xrayDaemonStartStop("stop")
					CurrentNodePos += 1
					log.Println("Next Position:", CurrentNodePos)

					CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
					xrayDaemonStartStop("start")
				}else{
					xrayDaemonStartStop("stop")
					CurrentNodePos = 0
					log.Println("Next Position:", CurrentNodePos)

					CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
					xrayDaemonStartStop("start")
				}

				mu.Unlock()
				feedbackCh <- true
			case "previous" :
				mu.Lock()
				log.Println("Cmd: Previous")

				if CurrentNodePos > 0 {
					xrayDaemonStartStop("stop")
					CurrentNodePos -= 1
					log.Println("Pre Position:", CurrentNodePos)

					CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
					xrayDaemonStartStop("start")
				}else{
					xrayDaemonStartStop("stop")
					CurrentNodePos = len(FirstBench.GoodNodes) - 1
					log.Println("Pre Position:", CurrentNodePos)

					CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
					xrayDaemonStartStop("start")
				}

				mu.Unlock()
				feedbackCh <- true
//			case "print" :
//				log.Println("Cmd: Print")
			case "quit" :
				mu.Lock()
				log.Println("Cmd: Quit")

				if daemonRunning {
					xrayDaemonStartStop("stop")
				}


				var nodeStack []*tools.Node
				nodeStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
				if err != nil {
					log.Println(err)
				}
				nodeStackPop(&nodeStack, FirstBench.PreLength)	//remove old firstBench nodes

				nodeStack = append(FirstBench.GoodNodes, nodeStack...)
				tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)

				ticker.Stop()
				cmdToRoutineCh <- true
				log.Println("Routine quit:", <- feedbackFromRoutineCh)

				log.Println("Auto monitor quit")
				feedbackCh <- true		//ready
				return

			case "auto" :
				mu.Lock()
				log.Println("Cmd: Auto")
				curStatus = 1
				if preStatus != 1 {
					switch preStatus {
					case 0 :	// First in
					case 2 :
						//TODO: Stop Manual
						routine(ctx)
					}

					ticker.Reset(tools.RoutinePeriodDu)
					log.Println("Auto mode Started")
					preStatus = 1
				}else{
					log.Println("Already in Auto mode")
				}

				mu.Unlock()
				feedbackCh <- true
			case "manual" :
				mu.Lock()
				log.Println("Cmd: Manual")
				curStatus = 2
				if preStatus != 2{
					switch preStatus {
					case 0:
					case 1:
						//TODO: Stop Auto
						ticker.Stop()
					}
					log.Println("Manual mode Started")
					preStatus = 2
				}else{
					log.Println("Already in Manual mode")
				}

				mu.Unlock()
				feedbackCh <- true
			}
		}
	}
}

func firstIn(ctx context.Context) error {
	log.Println("AutoMonitor FirstIn")

	nodeStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	if err != nil {
		return fmt.Errorf("GetNodesFromFormatedFile(Good):%w", err)
	}

	var bench Bench
	err = updateOneBenchFromStackR(ctx, &bench, &nodeStack)
	if err != nil {
		log.Println("No good nodes in Good.txt, trying to find any in bad...")
		refresh(ctx, []string{"bad"})
		nodeStack, err = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
		if err != nil {
			return fmt.Errorf("GetNodesFromFormatedFile(Good):%w", err)
		}

		err = updateOneBenchFromStackR(ctx, &bench, &nodeStack)
		if err != nil {
			return errors.New("No Good Nodes")
//			return fmt.Errorf("updateOneBenchFromStackR:%w", err)
		}
	}

	CurrentNodePos = 0
	CurrentNode = bench.GoodNodes[CurrentNodePos]
	FirstBench = &bench

	(*FirstBench).Clean()
	FirstBench.PreLength = len(FirstBench.GoodNodes)
	nodeStackPush(&nodeStack, FirstBench.GoodNodes)
	err = tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)
	if err != nil {
		errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Good):%w", err))
	}

	xrayDaemonStartStop("start")

	log.Println("FirstIn done")
	return nil
}

func fetchNew(ctx context.Context) error {

	log.Println("start oneShot")
	goodPingNodes, badPingNodes, errorNodes := oneShot(ctx)
	log.Println("oneShot done")

	sort.Stable(tools.ByDelay(goodPingNodes))

	CurrentNodePos = 0
	l := len(goodPingNodes)
	if l >= benchSize {
		CurrentNode = goodPingNodes[CurrentNodePos]
		xrayDaemonStartStop("stop")
		xrayDaemonStartStop("start")
		FirstBench.GoodNodes = goodPingNodes[:benchSize]

		(*FirstBench).Clean()
		FirstBench.PreLength = benchSize
	}else if l > 0 {
		CurrentNode = goodPingNodes[CurrentNodePos]
		xrayDaemonStartStop("stop")
		xrayDaemonStartStop("start")
		FirstBench.GoodNodes = goodPingNodes

		(*FirstBench).Clean()
		FirstBench.PreLength = len(goodPingNodes)
	}else{
		log.Println("NO valid nodes")
		return
	}



	tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
	tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
	tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

}

func routine(ctx context.Context) error {
	var nodeStack []*tools.Node
	var again true
	getStack:
	nodeStack, _ = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	_, err := nodeStackPop(&nodeStack, FirstBench.PreLength)	//remove old firstBench nodes
	if err != nil {
		return fmt.Errorf("nodeStackPop:%w", err)
	}

	err = updateOneBenchFromStackR(ctx, FirstBench, &nodeStack)
	if err != nil {
//		return fmt.Errorf("nodeStackPop:%w", err)
		log.Println("updateOneBenchFromStackR:%w", err)
		if again {
			again = false
			refresh(ctx, []string{"bad"})
			goto getStack
		}else{
			(*FirstBench).Clean()
			return errors.New("No Good Nodes")
		}
	}

	l := len(FirstBench.GoodNodes)

	if l > benchSize {
		goodNodes := FirstBench.GoodNodes[benchSize:]
		nodeStackPush(&nodeStack, goodNodes)

		FirstBench.GoodNodes = FirstBench.GoodNodes[:benchSize]

		(*FirstBench).Clean()
	}else if l >= benchSize/2 && l <= benchSize {
		(*FirstBench).Clean()
	}else if l >= 0 && l < benchSize/2 {
		for l < benchSize/2 {
			err := updateOneBenchFromStackR(ctx, FirstBench, &nodeStack)
			if err != nil {
				log.Println(err)
				refresh(ctx, []string{"bad"})
				goto getStack
			}

			l = len(FirstBench.GoodNodes)
			(*FirstBench).Clean()
		}
	}

	log.Println("Ping CurrentNode...")
	goodPingNodes, _, _, _, _ := ping.XrayPing(ctx, []*tools.Node{CurrentNode})
	if goodPingNodes == nil {
		xrayDaemonStartStop("stop")
		CurrentNodePos = 0
		CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
		xrayDaemonStartStop("start")
	}else{
		if CurrentNode.AvgDelay - FirstBench.MidDelay > 50 {
			log.Println("Update CurrentNode")
			xrayDaemonStartStop("stop")
			CurrentNodePos = 0
			CurrentNode = FirstBench.GoodNodes[CurrentNodePos]
			xrayDaemonStartStop("start")
		}
	}

	FirstBench.PreLength = len(FirstBench.GoodNodes)
	nodeStackPush(&nodeStack, FirstBench.GoodNodes)

	sort.Stable(tools.ByDelay(nodeStack))
	tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)

	return nil
}

//func findValidNode() *tools.Node {
//	log.Println("test nodes from file")
//
//	nodeStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
//	if err != nil {
//		return nil
//	}
//
//	var bench Bench
//	updateOneBenchFromStackR(&bench, &nodeStack)
//	if len(bench.GoodNodes) == 0 {
//		return nil
//	}
//	sort.Stable(tools.ByDelay(bench.GoodNodes))
//	return bench.GoodNodes[0]
//}

//func testNodesFromFile(filePath string) ([]*tools.Node, []*tools.Node, []*tools.Node) {
//	var nodes []*tools.Node
//	nodes, _ = tools.GetNodesFromFormatedFile(filePath)
//	log.Println("start ping nodes")
//	goodPingNodes, badPingNodes, errorNodes, _, _ := ping.XrayPing(nodes)
//	log.Println("ping nodes done")
//
//	return goodPingNodes, badPingNodes, errorNodes
//}

func oneShot(ctx context.Context) ([]*tools.Node, []*tools.Node, []*tools.Node) {
	//Get subscription links
	var nodeLs tools.NodeLists
	tools.GetAllNodes(&nodeLs)
	

	var allNodes []*tools.Node
	allNodes = append(nodeLs.Vms, nodeLs.Sses...)

	log.Println("Subs get done!")

	//Ping Tests
	goodPingNodes, badPingNodes, errorNodes, _, _ := ping.XrayPing(ctx, allNodes)

	sort.Stable(tools.ByDelay(goodPingNodes))
	return goodPingNodes, badPingNodes, errorNodes
}


func nodeStackPush(stack *[]*tools.Node, nodes []*tools.Node) {
	*stack = append(nodes, *stack...)
}

func nodeStackPop(stack *[]*tools.Node, num int) ([]*tools.Node, error) {
	var nodes []*tools.Node
	ls := len(*stack)
	if ls >= num {
		nodes = (*stack)[:num]
		(*stack) = (*stack)[num:]
	}else if ls > 0 {
		nodes = (*stack)
		(*stack) = nil
	}else{
		return nil, errors.New("stack is empty")
	}
	return nodes, nil
}

func xrayDaemonStartStop(cmd string) {
	switch cmd {
	case "start" :
		go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
		log.Println("XrayDaemon :", <- feedbackFromDaemonCh)
		daemonRunning = true
	case "stop" :
		cmdToDaemonCh <- "TERM"
		log.Println("XrayDaemon quit:", <- feedbackFromDaemonCh)
		daemonRunning = false
	}
}

//not used yet
//func nodesToBenches (nodes []*tools.Node) []*Bench {
//
//	var benches []*Bench
//	var bench *Bench
//
//	for i, node := range nodes {
//		if i%benchSize == 0 {
//			bench = new(Bench)
//			(*bench).GoodNodes = append((*bench).GoodNodes, node)
//			benches = append(benches, bench)
//		}
//		if i%benchSize > 0 {
//			(*bench).GoodNodes = append((*bench).GoodNodes, node)
//		}
//	}
//
//	return benches
//}

func updateOneBenchFromStackR(ctx context.Context, bench *Bench, stack *[]*tools.Node) error{
	//for len(*stack) > 0 {
	for {
		nodes, err := nodeStackPop(stack, benchSize)
		if err != nil {
			return fmt.Errorf("nodeStackPop:", err)
		}
		(*bench).GoodNodes = append((*bench).GoodNodes, nodes...)
		bench.Refresh(ctx)
		if len((*bench).GoodNodes) > 0 {
			return nil
		}
	}
}

func refresh(ctx context.Context, options []string) error{
	var errs []error
	for _, op := range options {
		if op == "bench" {
			log.Println("Refresh FirstBench")
			routine(ctx)
			CurrentNodePos = -1
			log.Println("Refresh FirstBench done")
		}else if op == "good" {
			log.Println("Refresh goodOut.txt")

			var nodes []*tools.Node
			nodes, err = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Good):%w", err))
			}
			oldBadNodes, err := tools.GetNodesFromFormatedFile(tools.BadOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Bad):%w", err))
			}

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, err := ping.XrayPing(ctx, nodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("XrayPing:%w", err))
			}
			log.Println("ping nodes done")

			badNodes := append(badPingNodes, oldBadNodes...)

			sort.Stable(tools.ByDelay(goodPingNodes))

			err = tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Good):%w", err))
			}
			err = tools.WriteNodesToFormatedFile(tools.BadOutPath, badNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Bad):%w", err))
			}
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh goodOut.txt done")
		}else if op == "bad" {
			log.Println("Refresh badOut.txt")

			var nodes []*tools.Node
			nodes, err := tools.GetNodesFromFormatedFile(tools.BadOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Bad):%w", err))
			}
			oldGoodNodes, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Good):%w", err))
			}

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, err := ping.XrayPing(ctx, nodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("XrayPing:%w", err))
			}
			log.Println("ping nodes done")

			goodNodes := append(oldGoodNodes, goodPingNodes...)

			sort.Stable(tools.ByDelay(goodNodes))

			err = tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Good):%w", err))
			}
			err = tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Bad):%w", err))
			}
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh badOut.txt done")
		}else if op == "all" {
			log.Println("Refresh All")

			var nodes, nodes2 []*tools.Node
			nodes, err = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Good):%w", err))
			}
			nodes2, err = tools.GetNodesFromFormatedFile(tools.BadOutPath)
			if err != nil {
				errs = append(errs, fmt.Errorf("GetNodesFromFormatedFile(Bad):%w", err))
			}
			allNodes := append(nodes, nodes2...)

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, err := ping.XrayPing(ctx, allNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("XrayPing:%w", err))
			}
			log.Println("ping nodes done")

			sort.Stable(tools.ByDelay(goodPingNodes))

			err = tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Good):%w", err))
			}
			err = tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
			if err != nil {
				errs = append(errs, fmt.Errorf("WriteNodesToFormatedFile(Bad):%w", err))
			}
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh All done")
		}
	}
	if len(errs) == 0 {
		return nil
	}else{
		var err error
		for i, e := range(errs) {
			err = fmt.Errorf("%w|%w", err, e)
		}
		err = fmt.Errorf("(%w)", err)
		return err
	}
}
