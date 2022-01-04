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
	BadNodesBuffer = append(BadNodesBuffer, b.BadNodes...)
	b.BadNodes = nil
}

const benchSize = 8
var curStatus int	// 0 for running normally, 1 for First in, 2 for Busy, 3 for error
var preStatus int

var cmdToDaemonCh = make(chan string)
var feedbackFromDaemonCh = make(chan string)

var CurrentNode *tools.Node
var FirstBench *Bench
var SecondBench *Bench

var BadNodesBuffer []*tools.Node

func AutoMonitor(cmdCh <-chan string, feedbackCh chan<- int) {
	log.Println("AutoMonitor Start")
	cmdToRoutineCh := make(chan bool)
	var ticker *time.Ticker

	preStatus = 0
	curStatus = 0
	for {
		select {
		case cmd := <-cmdCh :
			switch cmd {
			case "Auto" :
				log.Println("Cmd: Auto")
				curStatus = 1
				if preStatus != 1 {
					switch preStatus {
					case 0 :	// First in
						log.Println("First in")
						FirstIn()
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

			case "Manual" :
				log.Println("Cmd: Manual")
				curStatus = 2
				if preStatus != 2 {
					switch preStatus {
					case 0 :	// First in
						log.Println("First in")
						FirstIn()
					case 1 :
						ticker.Stop()
						cmdToRoutineCh <- true
						log.Println("Auto mode Stopped")
					}

					//TODO: Manual start


					preStatus = 2
				}
			case "Refresh" :
				log.Println("Cmd: Refresh")
				curStatus = 3
				//refresh()
			case "Pause/Resume" :
				log.Println("Cmd: Manual")
				curStatus = 4
				//stopDaemon()
			case "Stop" :
				log.Println("Cmd: Manual")
				curStatus = 5
				//stopCurrentAction()
			case "Quit" :
				log.Println("Cmd: Manual")
				curStatus = 6
				//quitProgram()

			}
		}
	}
}

func FirstIn() {
	goodNode := findValidNode()
	if goodNode == nil {
		log.Println("can't find valid node")
		return
	}

	log.Println("found valid node, start XrayDaemon")

	go tools.XrayDaemon(goodNode, cmdToDaemonCh, feedbackFromDaemonCh)

	log.Println("XrayDaemon is", <- feedbackFromDaemonCh)

	log.Println("start oneShot")
	goodPingNodes, badPingNodes, errorNodes := oneShot()
	log.Println("oneShot done")

	cmdToDaemonCh <- "TERM"
	log.Println("XrayDaemon :", <- feedbackFromDaemonCh)

	sort.Stable(tools.ByDelay(goodPingNodes))

	CurrentNode = goodPingNodes[0]
	FirstBench = new(Bench)
	SecondBench = new(Bench)
	(*FirstBench).GoodNodes = nodesStackPop(&goodPingNodes, benchSize)

	tools.WriteNodesToFormatedFile(tools.GoodOutPath, goodPingNodes)
	tools.WriteNodesToFormatedFile(tools.BadOutPath, badPingNodes)
	tools.WriteNodesToFormatedFile(tools.ErrorOutPath, errorNodes)


	go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
	log.Println("XrayDaemon :", <- feedbackFromDaemonCh)

}

func routine() {
	log.Println("In routine")
	(*FirstBench).Refresh()

	goodPingNodes, _, _, _, _ := ping.XrayPing([]*tools.Node{CurrentNode})
	if goodPingNodes == nil {
		//Stop XrayDaemon
		cmdToDaemonCh <- "TERM"
		log.Println("XrayDaemon :", <- feedbackFromDaemonCh)

		CurrentNode = FirstBench.GoodNodes[0]
		go tools.XrayDaemon(CurrentNode, cmdToDaemonCh, feedbackFromDaemonCh)
		log.Println("XrayDaemon :", <- feedbackFromDaemonCh)
	}


	count := benchSize - len(FirstBench.GoodNodes)		//bad nodes count
	log.Println("count =", count)
	if count >= benchSize/2 {
		var nodeLs tools.NodeLists
		tools.GetNodeLsFromFormatedFile(&nodeLs, tools.GoodOutPath)
		var stack []*tools.Node
		stack = append(nodeLs.Vms, nodeLs.Sses...)

		SecondBench.GoodNodes = nodesStackPop(&stack, count)
		(*SecondBench).Refresh()
		for count >= benchSize/2 {
			good2 := len(SecondBench.GoodNodes)
			if good2 >= count {	//SecondBench has enough GoodNodes to add to FirstBench
				FirstBench.GoodNodes = append(FirstBench.GoodNodes, SecondBench.GoodNodes[:count]...)
				(*FirstBench).Clean()

				nodesStackPush(&stack, SecondBench.GoodNodes[count:])
				SecondBench.GoodNodes = nil
				(*SecondBench).Clean()

				count = 0
			}else{		//SecondBench dosn't have enough GoodNodes that fit in FirstBench
				if len(SecondBench.GoodNodes) > 0 {
					FirstBench.GoodNodes = append(FirstBench.GoodNodes, SecondBench.GoodNodes...)
					count = count - good2

					SecondBench.GoodNodes = nodesStackPop(&stack, count)
					(*SecondBench).Refresh()
				}else{
					SecondBench.GoodNodes = nodesStackPop(&stack, count)
					(*SecondBench).Refresh()
				}
			}
		}
		//Write back GoodOutFile
		tools.WriteNodesToFormatedFile(tools.GoodOutPath, stack)
	}else{
		(*FirstBench).Clean()
	}

	//Write back BadOutFile
	var nodeLs tools.NodeLists
	tools.GetNodeLsFromFormatedFile(&nodeLs, tools.BadOutPath)
	var stack []*tools.Node
	stack = append(nodeLs.Vms, nodeLs.Sses...)
	stack = append(stack, BadNodesBuffer...)
	BadNodesBuffer = nil

	tools.WriteNodesToFormatedFile(tools.BadOutPath, stack)

	sort.Stable(tools.ByDelay(FirstBench.GoodNodes))
}


func findValidNode() *tools.Node {
	log.Println("test nodes from file")
	var goodNodes []*tools.Node

	nodesStack, err := tools.GetNodesFromFormatedFile(tools.GoodOutPath)
	if err != nil {
		return nil
	}

	for len(goodNodes) == 0 {
		if len(nodesStack) == 0 {
			return nil
		}
		nodes := nodesStackPop(&nodesStack, benchSize)
		goodNodes, _, _, _, _ = ping.XrayPing(nodes)
	}

	sort.Stable(tools.ByDelay(goodNodes))
	
	return goodNodes[0]
}

func testNodesFromFile(fileName string) ([]*tools.Node, []*tools.Node, []*tools.Node) {
	var nodeLs tools.NodeLists
	tools.GetNodeLsFromFile(&nodeLs, fileName)
	log.Println("get nodesls from file done")

	var allNodes []*tools.Node
	allNodes = append(nodeLs.Vms, nodeLs.Sses...)
	log.Println("start ping nodes")
	goodPingNodes, badPingNodes, errorNodes, _, _ := ping.XrayPing(allNodes)
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


func nodesStackPush(stack *[]*tools.Node, nodes []*tools.Node) {
	*stack = append(nodes, *stack...)
}

func nodesStackPop(stack *[]*tools.Node, num int) []*tools.Node {
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
