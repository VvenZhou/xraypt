package monitor

import(
	"log"
	"errors"
	"fmt"
	"sort"
	"time"
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

func (b *Bench) Refresh() {
	goodPingNodes, badPingNodes, _, _, _ := ping.XrayPing(append((*b).GoodNodes, (*b).BadNodes...))
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
var FirstBench *Bench

var BadNodesBuffer []*tools.Node
var curStatus int
var preStatus int

var daemonStatus int		//0 for stopped, 1 for started

func AutoMonitor(cmdCh <-chan string, feedbackCh chan<- int, dataCh <-chan string) {
	log.Println("AutoMonitor Start")
	var ticker *time.Ticker

	cmdToRoutineCh := make(chan bool)
	feedbackFromRoutineCh := make(chan bool)

	var mu sync.Mutex

	curStatus = 0
	preStatus = 0

	firstIn()

	feedbackCh <- 0
	for {
		select {
		case cmd := <-cmdCh :
			switch cmd {
			case "refresh" :
				mu.Lock()

				log.Println("Cmd: Refresh")

				if curStatus == 1 {
					ticker.Stop()
				}

				var data []string
				for i:=0; i<2; i++{
					data = append(data, <-dataCh)
				}
				refresh(data)

				if curStatus == 1 {
					ticker.Reset(tools.RoutinePeriodDu)
				}

				mu.Unlock()
				feedbackCh <- 0
			case "fetch" :
				mu.Lock()

				log.Println("Cmd: FetchNew")

				if curStatus == 1 {
					ticker.Stop()
				}

				if daemonStatus == 0 {
					log.Println("XrayDaemon is not running, you can't fetch anything.")
					break
				}

				fetchNewNodesToFile()

				if curStatus == 1 {
					ticker.Reset(tools.RoutinePeriodDu)
				}

				mu.Unlock()
				feedbackCh <- 0
			case "pause" :
				mu.Lock()

				log.Println("Cmd: Pause")

				switch daemonStatus {
				case 0 :
					if curStatus == 1 {
						ticker.Reset(tools.RoutinePeriodDu)
					}
					xrayDaemonStartStop("start")
				case 1 :
					if curStatus == 1 {
						ticker.Stop()
					}
					xrayDaemonStartStop("stop")
				}

				mu.Unlock()
				feedbackCh <- 0
			case "print" :
				log.Println("Cmd: Print")
				//stopCurrentAction()
			case "quit" :
				mu.Lock()
				log.Println("Cmd: Quit")

				if curStatus == 1 {
					ticker.Stop()
				}

				xrayDaemonStartStop("stop")

				if curStatus == 1 {
					cmdToRoutineCh <- true 
					log.Println("Routine quit:", <- feedbackFromRoutineCh)
				}

				log.Println("Auto monitor quit")
				feedbackCh <- 0		//ready
				return

			case "auto" :
				mu.Lock()
				log.Println("Cmd: Auto")
				curStatus = 1
				if preStatus != 1 {
					switch preStatus {
					case 0 :	// First in
					case 2 :
						routine()
						//TODO: Stop Manual
					}

					ticker = time.NewTicker(tools.RoutinePeriodDu)
					go func() {
						for {
							select {
							case <-cmdToRoutineCh :
								feedbackFromRoutineCh <- true
								return
							case <-ticker.C :
								ticker.Stop()
								mu.Lock()
								routine()
								mu.Unlock()
								ticker.Reset(tools.RoutinePeriodDu)
							}
						}
					}()
					log.Println("Auto mode Started")
					preStatus = 1
				}

				mu.Unlock()
				feedbackCh <- 0
			case "manual" :
//				log.Println("Cmd: Manual")
//				curStatus = 2
//				if preStatus != 2 {
//					switch preStatus {
//					case 0 :	// First in
//						log.Println("First in")
//						//FirstIn()
//					case 1 :
//						ticker.Stop()
//						cmdToRoutineCh <- true
//						log.Println("Auto mode Stopped")
//					}
//
//					//TODO: Manual start
//
//
//					log.Println("Manual mode Started")
//					preStatus = 2
//				}


			}
		}
	}
}

func firstIn() {
	log.Println("AutoMonitor FirstIn")

	nodeStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	if err != nil {
		return
	}

	var bench Bench
	updateOneBenchFromStackR(&bench, &nodeStack)
	if len(bench.GoodNodes) == 0 {
		return
	}

	CurrentNode = bench.GoodNodes[0]
	FirstBench = &bench

	(*FirstBench).Clean()
	FirstBench.PreLength = 0
	nodeStackPush(&nodeStack, FirstBench.GoodNodes)
	tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)

	xrayDaemonStartStop("start")

	log.Println("FirstIn done")
}

func fetchNewNodesToFile() {

	log.Println("start oneShot")
	goodPingNodes, badPingNodes, errorNodes := oneShot()
	log.Println("oneShot done")

	sort.Stable(tools.ByDelay(goodPingNodes))

	tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
	tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
	tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

}

func routine() error {
	var nodeStack []*tools.Node
	getStack:
	nodeStack, _ = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	_, err := nodeStackPop(&nodeStack, FirstBench.PreLength)	//remove old firstBench nodes
	if err != nil {
		err = fmt.Errorf("nodeStackPop:", err)
		return err
	}

	err = updateOneBenchFromStackR(FirstBench, &nodeStack)
	if err != nil {
		log.Println(err)
		refresh([]string{"bad"})
		goto getStack
	}

	l := len(FirstBench.GoodNodes)

	if l > benchSize {
		(*FirstBench).Clean()
		goodNodes := FirstBench.GoodNodes[benchSize:]
		nodeStackPush(&nodeStack, goodNodes)

		FirstBench.GoodNodes = FirstBench.GoodNodes[:benchSize]

	}else if l >= benchSize/2 && l <= benchSize {
		(*FirstBench).Clean()

	}else if l >= 0 && l < benchSize/2 {
		for l < benchSize/2 {
			err := updateOneBenchFromStackR(FirstBench, &nodeStack)
			if err != nil {
				log.Println(err)
				refresh([]string{"bad"})
				goto getStack
			}

			l = len(FirstBench.GoodNodes)
			(*FirstBench).Clean()
		}
	}

	log.Println("Ping CurrentNode...")
	goodPingNodes, _, _, _, _ := ping.XrayPing([]*tools.Node{CurrentNode})
	if goodPingNodes == nil {
		xrayDaemonStartStop("stop")
		CurrentNode = FirstBench.GoodNodes[0]
		xrayDaemonStartStop("start")
	}else{
		if CurrentNode.AvgDelay > FirstBench.MidDelay {
			xrayDaemonStartStop("stop")
			CurrentNode = FirstBench.GoodNodes[0]
			xrayDaemonStartStop("start")
		}
	}

	FirstBench.PreLength = len(FirstBench.GoodNodes)
	nodeStackPush(&nodeStack, FirstBench.GoodNodes)

	sort.Stable(tools.ByDelay(nodeStack))
	tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)

	return nil
}

func findValidNode() *tools.Node {
	log.Println("test nodes from file")

	nodeStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	if err != nil {
		return nil
	}

	var bench Bench
	updateOneBenchFromStackR(&bench, &nodeStack)
	if len(bench.GoodNodes) == 0 {
		return nil
	}

	sort.Stable(tools.ByDelay(bench.GoodNodes))
	
	return bench.GoodNodes[0]
}

func testNodesFromFile(filePath string) ([]*tools.Node, []*tools.Node, []*tools.Node) {
	var nodes []*tools.Node
	nodes, _ = tools.GetNodesFromFormatedFile(filePath)
	log.Println("start ping nodes")
	goodPingNodes, badPingNodes, errorNodes, _, _ := ping.XrayPing(nodes)
	log.Println("ping nodes done")

	return goodPingNodes, badPingNodes, errorNodes
}

func oneShot() ([]*tools.Node, []*tools.Node, []*tools.Node) {
	//Get subscription links
	var nodeLs tools.NodeLists
	tools.GetAllNodes(&nodeLs)

	var allNodes []*tools.Node
	allNodes = append(nodeLs.Vms, nodeLs.Sses...)

	log.Println("Subs get done!")

	//Ping Tests
	goodPingNodes, badPingNodes, errorNodes, _, _ := ping.XrayPing(allNodes)

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
		daemonStatus = 1
	case "stop" :
		cmdToDaemonCh <- "TERM"
		log.Println("XrayDaemon quit:", <- feedbackFromDaemonCh)
		daemonStatus = 0
	}
}

//not used yet
func nodesToBenches (nodes []*tools.Node) []*Bench {

	var benches []*Bench
	var bench *Bench

	for i, node := range nodes {
		if i%benchSize == 0 {
			bench = new(Bench)
			(*bench).GoodNodes = append((*bench).GoodNodes, node)
			benches = append(benches, bench)
		}
		if i%benchSize > 0 {
			(*bench).GoodNodes = append((*bench).GoodNodes, node)
		}
	}

	return benches
}

func updateOneBenchFromStackR(bench *Bench, stack *[]*tools.Node) error{
	//for len(*stack) > 0 {
	for {
		nodes, err := nodeStackPop(stack, benchSize)
		if err != nil {
			err = fmt.Errorf("nodeStackPop:", err)
			return err
		}
		(*bench).GoodNodes = append((*bench).GoodNodes, nodes...)
		bench.Refresh()
		if len((*bench).GoodNodes) > 0 {
			return nil
		}
	}
}

func refresh(options []string) {
	for _, op := range options {
		if op == "bench" {
			log.Println("Refresh FirstBench")
			routine()
			log.Println("Refresh FirstBench done")
		}else if op == "good" {
			log.Println("Refresh goodOut.txt")

			var nodes []*tools.Node
			nodes, _ = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
			oldBadNodes, _ := tools.GetNodesFromFormatedFile(tools.BadOutPath)

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, _ := ping.XrayPing(nodes)
			log.Println("ping nodes done")

			badNodes := append(badPingNodes, oldBadNodes...)

			sort.Stable(tools.ByDelay(goodPingNodes))

			tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
			tools.WriteNodesToFormatedFile(tools.BadOutPath, badNodes)
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh goodOut.txt done")
		}else if op == "bad" {
			log.Println("Refresh badOut.txt")

			var nodes []*tools.Node
			nodes, _ = tools.GetNodesFromFormatedFile(tools.BadOutPath)
			oldGoodNodes, _ := tools.GetNodesFromFormatedFile(tools.GoodOutPath)

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, _ := ping.XrayPing(nodes)
			log.Println("ping nodes done")

			goodNodes := append(oldGoodNodes, goodPingNodes...)

			sort.Stable(tools.ByDelay(goodNodes))

			tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodNodes)
			tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh badOut.txt done")
		}else if op == "all" {
			log.Println("Refresh All")

			var nodes, nodes2 []*tools.Node
			nodes, _ = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
			nodes2, _ = tools.GetNodesFromFormatedFile(tools.BadOutPath)
			allNodes := append(nodes, nodes2...)

			log.Println("start ping nodes")
			goodPingNodes, badPingNodes, _, _, _ := ping.XrayPing(allNodes)
			log.Println("ping nodes done")

			sort.Stable(tools.ByDelay(goodPingNodes))

			tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
			tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
			//tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)

			log.Println("Refresh All done")
		}
	}
}
