package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"os"
	"io"
	"bufio"
	"os/exec"
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

var status bool

var LogLineNum = 40

func main() {

	tools.PreCheck(tools.MainPort, protocols)

	logf := startLogSystem()
	defer logf.Close()
	defer fmt.Printf("\n")
//	defer stopCmd(logCmd)


	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	cmdCh := make(chan string)	
	feedbackCh := make(chan bool)		//0 for ready to go, 1 for busy, 2 for error
	dataCh := make(chan string, 3)

	scanner := bufio.NewScanner(os.Stdin)

	go monitor.AutoMonitor(cmdCh, feedbackCh, dataCh)

	status = <- feedbackCh

	for {
		var sList []string
		getInput: for {

			scanner.Scan()
			sList = strings.Fields(scanner.Text())
			l := len(sList)
			if l != 0 {
				status = false
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
					status = true
					log.Println("bad data")
					goto getInput
				}
			}

			dataCh <- datas[0]
			dataCh <- datas[1]

			cmdCh <- "refresh"

			status = <- feedbackCh

		case "fetch" :
			cmdCh <- "fetch"

			status = <- feedbackCh
		case "pause" :
			cmdCh <- "pause"
			status = <- feedbackCh
		case "print" :
			cmdCh <- "print"
		case "quit" :
			cmdCh <- "quit"

			status = <- feedbackCh

			os.RemoveAll(tools.TempPath)
			os.MkdirAll(tools.TempPath, 0755)

			time.Sleep(100 * time.Millisecond)
			return
		case "auto" :
			cmdCh <- "auto"
			status = <- feedbackCh
//		case "manual" :
//			cmdCh <- "manual"
//		case "help" :
//		case "clear", "clr" :
//			for i:=0; i<LogLineNum; i++ {
//				log.Printf("\n")
//			}
//
//			status = true
//
//			goto getInput
		default:
			status = true
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

func startLogSystem() *os.File {
	f, err := os.OpenFile(tools.LogPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	r, w := io.Pipe()
        multi := io.MultiWriter(f, w)
        log.SetOutput(multi)

        go func() {
		var content string
                var logC string
                logC = "(log system)"
		head := "\nLog:\n\n"
		prompt := "\nEnter command: "
		wait := "\nPlease wait..."


                scanner := bufio.NewScanner(r)
                for scanner.Scan() {
			s := scanner.Text()
                        logUpdate(&logC, s)
                        if status == true {
				content = head + logC + "\n" + prompt
			}else{
				content = head + logC + "\n" + wait
			}

                        cmd := exec.Command("clear")
                        cmd.Stdout = os.Stdout
                        cmd.Run()

                        fmt.Printf("%s", content)
                }
                if err := scanner.Err(); err != nil {
                        fmt.Fprintln(os.Stderr, "reading standard input:", err)
                }
        }()

	return f
}

func logUpdate(logC *string, newThings string) {
        lines := strings.Split(*logC, "\n")
        curLine := len(lines)

        if curLine >= LogLineNum {
                lines = append(lines, newThings)
                lines = lines[1:]
        }else{
                lines = append(lines, newThings)
        }

        *logC = strings.Join(lines, "\n")
}             

