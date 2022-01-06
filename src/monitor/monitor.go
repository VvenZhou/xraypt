package monitor

import(
	"log"
	"sort"
	"time"
	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/ping"
//	"github.com/VvenZhou/xraypt/src/speedtest"
)

type Bench struct {
	GoodNodes []*tools.Node
	BadNodes []*tools.Node
	PreLength int
//	Good int
	Score int
}

func (b *Bench) Refresh() {
	goodPingNodes, badPingNodes, _, _, _ := ping.XrayPing(append((*b).GoodNodes, (*b).BadNodes...))
	sort.Stable(tools.ByDelay(goodPingNodes))

	b.GoodNodes = goodPingNodes
	b.BadNodes = append(b.BadNodes, badPingNodes...)
//	b.Good = len(goodPingNodes)
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

func AutoMonitor(cmdCh <-chan string, feedbackCh chan<- int, dataCh <-chan string) {
	log.Println("AutoMonitor Start")
	cmdToRoutineCh := make(chan bool)
	feedbackFromRoutineCh := make(chan bool)
	var ticker *time.Ticker

	curStatus = 0
	preStatus = 0

	firstIn()

	feedbackCh <- 0
	for {
		select {
		case cmd := <-cmdCh :
			switch cmd {
			case "refresh" :
				feedbackCh <- 1		//busy
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
				feedbackCh <- 0		//busy
			case "fetch" :
				feedbackCh <- 1
				log.Println("Cmd: FetchNew")

				if curStatus == 1 {
					ticker.Stop()
				}

				fetchNewNodesToFile()

				if curStatus == 1 {
					ticker.Reset(tools.RoutinePeriodDu)
				}
				feedbackCh <- 0
			case "pause" :
				log.Println("Cmd: Pause")
				//stopDaemon()
			case "print" :
				log.Println("Cmd: Print")
				//stopCurrentAction()
			case "quit" :
				feedbackCh <- 1		//busy
				log.Println("Cmd: Quit")

				if curStatus == 1 {
					ticker.Stop()
				}

				cmdToDaemonCh <- "TERM"
				log.Println("XrayDaemon quit:", <- feedbackFromDaemonCh)

				if curStatus == 1 {
					cmdToRoutineCh <- true 
					log.Println("Routine quit:", <- feedbackFromRoutineCh)
				}

				feedbackCh <- 0		//ready
				log.Println("Auto monitor quit")
				return

			case "auto" :
				feedbackCh <- 1		//busy
				log.Println("Cmd: Auto")
				curStatus = 1
				if preStatus != 1 {
					switch preStatus {
					case 0 :	// First in
						//log.Println("First in")
						//FirstIn()
					case 2 :
						//TODO: Stop Manual
					}


					if preStatus != 0 {
						routine()
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
								//log.Println("ticker stop")
								routine()
								ticker.Reset(tools.RoutinePeriodDu)
								//log.Println("ticker reset")
							}
						}
					}()
					log.Println("Auto mode Started")
					preStatus = 1
				}
				feedbackCh <- 0		//ready

			case "manual" :
				feedbackCh <- 1
				log.Println("Cmd: Manual")
				curStatus = 2
				if preStatus != 2 {
					switch preStatus {
					case 0 :	// First in
						log.Println("First in")
						//FirstIn()
					case 1 :
						ticker.Stop()
						cmdToRoutineCh <- true
						log.Println("Auto mode Stopped")
					}

					//TODO: Manual start


					log.Println("Manual mode Started")
					preStatus = 2
				}

				feedbackCh <- 0

			}
		}
	}
}

func firstIn() {
	log.Println("Init")

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

	go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
	log.Println("XrayDaemon :", <- feedbackFromDaemonCh)

	log.Println("Init done")
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

func routine() {
	var firstBenchGoodFlag bool
	log.Println("\n", "In routine")
	log.Println("FirstBench ping")
	(*FirstBench).Refresh()
	if len(FirstBench.GoodNodes) > 0 {
		firstBenchGoodFlag = true
	}else{
		firstBenchGoodFlag = false
	}

	log.Println("CurrentNode ping")
	goodPingNodes, _, _, _, _ := ping.XrayPing([]*tools.Node{CurrentNode})
	if goodPingNodes == nil {
		cmdToDaemonCh <- "TERM"
		log.Println("XrayDaemon :", <- feedbackFromDaemonCh)

		if firstBenchGoodFlag == true {
			CurrentNode = FirstBench.GoodNodes[0]
			go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
			log.Println("XrayDaemon :", <- feedbackFromDaemonCh)
		}
	}


	count := benchSize - len(FirstBench.GoodNodes)
	log.Println("count:", count)
	if count >= benchSize/2 {
		var nodeStack []*tools.Node
		nodeStack, _ = tools.GetNodesFromFormatedFile(tools.GoodOutPath)
		nodeStackPop(&nodeStack, FirstBench.PreLength)	//remove old firstBench nodes

//		if len(nodeStack) == benchSize {
//			refresh([]string{ "bad" })
//		}

		var bench Bench
		updateOneBenchFromStackR(&bench, &nodeStack)

		for count > 0 {
			if len(nodeStack) <= benchSize {
				refresh([]string{ "bad" })
			}
			good2 := len(bench.GoodNodes)
			if good2 >= count {
				log.Println("got enough nodes to append")
				FirstBench.GoodNodes = append(FirstBench.GoodNodes, bench.GoodNodes[:count]...)
				bench.GoodNodes = bench.GoodNodes[:count]
				count = 0
			}else{
				if good2 > 0 {
					log.Println("got nodes to append")
					FirstBench.GoodNodes = append(FirstBench.GoodNodes, bench.GoodNodes...)
					count = count - good2
					log.Println("count:", count)
					bench.Clean()

					updateOneBenchFromStackR(&bench, &nodeStack)
				}else{
					log.Println("no good Nodes in benchBuf")
					bench.Clean()
					updateOneBenchFromStackR(&bench, &nodeStack)
				}
			}
		}

		bench.Clean()
		nodeStackPush(&nodeStack, bench.GoodNodes)
		bench.GoodNodes = nil

		(*FirstBench).Clean()
		FirstBench.PreLength = len(FirstBench.GoodNodes)
		nodeStackPush(&nodeStack, FirstBench.GoodNodes)

		tools.WriteNodesToFormatedFile(tools.GoodOutPath, nodeStack)

	}else{
		(*FirstBench).Clean()
	}

	if firstBenchGoodFlag == false {
		CurrentNode = FirstBench.GoodNodes[0]
		go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
		log.Println("XrayDaemon :", <- feedbackFromDaemonCh)
	}

	//Write back BadOutFile
	if len(BadNodesBuffer) > 0 {
		var nodeStack []*tools.Node
		nodeStack, _ = tools.GetNodesFromFormatedFile(tools.BadOutPath)

		nodeStack = append(BadNodesBuffer, nodeStack...)
		BadNodesBuffer = nil
		tools.WriteNodesToFormatedFile(tools.BadOutPath, nodeStack)
	}

	sort.Stable(tools.ByDelay(FirstBench.GoodNodes))
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

func nodeStackPop(stack *[]*tools.Node, num int) []*tools.Node {
	var nodes []*tools.Node
	ls := len(*stack)
	if ls >= num {
		nodes = (*stack)[:num]
		(*stack) = (*stack)[num:]
	}else if ls > 0 {
		nodes = (*stack)
		(*stack) = nil
	}else{
		return nil
	}
	return nodes
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

func updateOneBenchFromStackR(bench *Bench, stack *[]*tools.Node) {
	for len(*stack) > 0 {
		(*bench).GoodNodes = append((*bench).GoodNodes, nodeStackPop(stack, benchSize)...)
		bench.Refresh()
		if len((*bench).GoodNodes) > 0 {
			return
		}
	}

	return
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
