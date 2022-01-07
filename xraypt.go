package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"os"
	"bufio"
	"os/exec"
	"syscall"
	"time"

	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/monitor"
)


var protocols = []string{
	"vmess",
	"vless",
	"ss",
	"ssr",
	"trojan"}


func main() {

	tools.PreCheck(tools.MainPort, protocols)

	logf, logCmd := startLogSystem()
	defer logf.Close()
	defer stopCmd(logCmd)


	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	cmdCh := make(chan string)	
	feedbackCh := make(chan int)		//0 for ready to go, 1 for busy, 2 for error
	dataCh := make(chan string, 3)

	scanner := bufio.NewScanner(os.Stdin)

	go monitor.AutoMonitor(cmdCh, feedbackCh, dataCh)

	<- feedbackCh

	for {

		var sList []string
		getInput: for {
			//TODO: print input Prompt
			fmt.Printf("Enter command: ")

			scanner.Scan()
			sList = strings.Fields(scanner.Text())
			l := len(sList)
			if l != 0 {
				break
			}
		}
		switch sList[0] {
		case "refresh" :

			datas := []string{"bench", ""}

			for i, s := range sList[1:] {
				switch s {
				case "bench", "good", "bad", "all":
					datas[i] = s
				default:
					log.Println("bad data")
					goto getInput
				}
			}

			dataCh <- datas[0]
			dataCh <- datas[1]

			cmdCh <- "refresh"

			<- feedbackCh

		case "fetch" :
			cmdCh <- "fetch"

			<- feedbackCh
		case "pause" :
			cmdCh <- "pause"
			<- feedbackCh
		case "print" :
			cmdCh <- "print"
		case "quit" :
			cmdCh <- "quit"

			<- feedbackCh

			os.RemoveAll(tools.TempPath)
			os.MkdirAll(tools.TempPath, 0755)

			time.Sleep(100 * time.Millisecond)
			return
		case "auto" :
			cmdCh <- "auto"
			<- feedbackCh
		case "manual" :
			cmdCh <- "manual"
		case "help" :
		case "clear", "clr" :
//			for i, s := range sList[1:] {
//				switch s {
//				case "log" :
//				default:
//					log.Println("bad data")
//					goto getInput
//				}
//			}

			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
			goto getInput
		default:
			log.Println("bad cmd")
			goto getInput
		}
	}
}

//Not used
func generateSpeedOutFile(nodes []*tools.Node) {
	var goodVmLinks []string
	for i, n := range nodes {
		fmt.Println(i, n.Type, (*n).AvgDelay, (*n).Country, " ", (*n).DLSpeed, " ", (*n).ULSpeed)
		//(*n).Id = strconv.Itoa(i)
		n.CreateFinalJson(tools.JsonsPath, strconv.Itoa(i))
		str := []string{strconv.Itoa(i), "\n", n.Type, "://", (*n).ShareLink,
				"\nDelay:", strconv.Itoa((*n).AvgDelay),
				" Down: ", fmt.Sprintf("%.2f", (*n).DLSpeed),
				" Up: ", fmt.Sprintf("%.2f", (*n).ULSpeed),
				" Country: ", (*n).Country, "\n"}
		vmOutStr := strings.Join(str, "")
		goodVmLinks = append(goodVmLinks, vmOutStr)
	}

	if len(goodVmLinks) != 0 {
		bytes := []byte(strings.Join(goodVmLinks[:], "\n"))
		err := os.WriteFile(tools.SpeedOutPath, bytes, 0644)
		if err != nil {
			log.Println(err)
		}else{
			log.Println("vmOut generated!")
		}
	}

}

func startLogSystem() (*os.File, *exec.Cmd) {
	f, err := os.OpenFile(tools.LogPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(f)

	dir, _ := os.Getwd()
	fullPath := fmt.Sprintf("%s/%s", dir, tools.LogPath)

	cmd := exec.Command("gnome-terminal", "--", "tools/tail.sh", fullPath)
	err = cmd.Start()
	if err != nil {
		f.Close()
		log.Fatal(err)
	}

	return f, cmd
}

func stopCmd(cmd *exec.Cmd) {
	err := cmd.Process.Signal(syscall.SIGTERM)

	_, err = cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("LogSystem quit", "\n")
}
