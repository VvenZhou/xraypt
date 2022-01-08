package main

import (
	"flag"
	"fmt"
	"log"
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

	flag.IntVar(&tools.MainPort, "mp", 8123, "main proxy out port num")
	flag.IntVar(&tools.PreProxyPort, "pp", 8123, "pre proxy in port num")
	flag.IntVar(&tools.RoutinePeriod, "rp", 300, "auto mode refresh routine period (unit: second)")

	flag.Parse()

	tools.PreCheck(protocols)

	logf := startLogSystem()
	defer logf.Close()
	defer fmt.Printf("\n")
	defer log.Printf("\n\n\n")


	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	cmdCh := make(chan string)	
	feedbackCh := make(chan bool)		//0 for ready to go, 1 for busy, 2 for error
	dataCh := make(chan string, 3)

	scanner := bufio.NewScanner(os.Stdin)

	go monitor.AutoMonitor(cmdCh, feedbackCh, dataCh)

	status = <- feedbackCh
	log.Println()

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
			log.Println()

		case "fetch" :
			cmdCh <- "fetch"

			status = <- feedbackCh
			log.Println()
		case "pause" :
			cmdCh <- "pause"
			status = <- feedbackCh
			log.Println()
		case "print" :
			cmdCh <- "print"
		case "quit" :
			cmdCh <- "quit"

			status = <- feedbackCh
			log.Println()

			os.RemoveAll(tools.TempPath)
			os.MkdirAll(tools.TempPath, 0755)

			time.Sleep(100 * time.Millisecond)
			return
		case "auto" :
			cmdCh <- "auto"
			status = <- feedbackCh
			log.Println()
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
		head := "\nLogs:\n\n"
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

