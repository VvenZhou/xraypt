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
	"os/signal"
	"time"
	"context"
	"syscall"

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
var userInt bool

func main() {

	flag.IntVar(&tools.MainPort, "mp", 8123, "main proxy out port num")
	flag.IntVar(&tools.PreProxyPort, "pp", 8123, "pre proxy in port num")
	flag.IntVar(&tools.RoutinePeriod, "rp", 300, "auto mode refresh routine period (unit: second)")
	flag.IntVar(&tools.PThreadNum, "tn", 160, "ping worker(thread) num")

	flag.Parse()

	tools.PreCheck(protocols)

	logf := startLogSystem()
	defer fmt.Printf("\n")
	defer logf.Close()
	defer log.Printf("Xraypt quit\n")

	os.RemoveAll(tools.TempPath)
	os.MkdirAll(tools.TempPath, 0755)

	cmdCh := make(chan string)	
	feedbackCh := make(chan bool)
	dataCh := make(chan string, 3)
	feedbackFromCli := make(chan bool)

	ctx := context.Background()
	ctx = setupCloseHandler(ctx)

	go monitor.AutoMonitor(ctx, cmdCh, feedbackCh, dataCh)

//	status = <- feedbackCh
//	log.Println()
//	cmdCh <- "auto"
//	status = <- feedbackCh
//	log.Println()


	go userCliWithContext(ctx, cmdCh, feedbackCh, dataCh, feedbackFromCli)

	select{
	case <-ctx.Done():
		log.Println("Main quit...")
		<-feedbackFromCli
	case <-feedbackFromCli:
		log.Println("Main quit...")
	}

	log.Println("...Main quit")
//	time.Sleep(100 * time.Millisecond)

	return
}

func startLogSystem() *os.File {
//	f, err := os.OpenFile(tools.LogPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	f, err := os.OpenFile(tools.LogPath, os.O_RDWR | os.O_CREATE ,0666)
	if err != nil {
		log.Fatal(err)
	}

	r, w := io.Pipe()
	multi := io.MultiWriter(f, w)
        log.SetOutput(multi)

        go func() {
		prompt := "Enter command: "
		wait := "Please wait..."

                scanner := bufio.NewScanner(r)
		stdin := bufio.NewReader(os.Stdin)
                for scanner.Scan() {
			s := scanner.Text()
			fmt.Printf("\r%s", s)
                        if status == true {
				stdin.Discard(stdin.Buffered())		//flush stdin buffer
				fmt.Printf("\n%s", prompt)
			}else{
				stdin.Discard(stdin.Buffered())		//flush stdin buffer
				fmt.Printf("\n%s", wait)
			}
                }
                if err := scanner.Err(); err != nil {
                        fmt.Fprintln(os.Stderr, "reading standard input:", err)
                }
        }()

	return f
}

func setupCloseHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	userInt = false
	go func(){
		for{
			<-c
			userInt = true
			log.Println("Ctrl+C pressed in Terminal")
			cancel()
		}
	}()

	return ctx
}

func userCliWithContext(ctx context.Context,
			cmdToMon chan<- string,
			feedbackFromMon <-chan bool,
			dataToMon chan<- string,
			feedbackToMain chan<- bool){
	status = <- feedbackFromMon
	log.Println()
	for{
		select{
		case <-ctx.Done():
			log.Println("UserCli get quit sig...")
			cmdToMon <- "quit"
			<- feedbackFromMon		//wait for monitor to quit

			os.RemoveAll(tools.TempPath)
			os.MkdirAll(tools.TempPath, 0755)

			time.Sleep(100 * time.Millisecond)
			feedbackToMain <- true
			log.Println("...UserCli quit")
			return
		default:
			quit := userCli(cmdToMon, feedbackFromMon, dataToMon)
			if quit {		//normal quit
				log.Println("UserCli get quit cmd...")
				feedbackToMain <- true
				log.Println("...UserCli quit")
				return
			}
		}
	}
}

func userCli(cmdCh chan<- string, feedbackCh <-chan bool, dataCh chan<- string) bool {
	scanner := bufio.NewScanner(os.Stdin)
//	for {
		var sList []string
//		getInput: for {
		for {
			scanner.Scan()
			if userInt {
				return false	//Must be false, for abnormal quit
			}
			sList = strings.Fields(scanner.Text())
			l := len(sList)
			if l != 0 {
				status = false
				break
			}else{
				log.Println()
			}
		}
		switch sList[0] {
		case "r", "refresh" :

			datas := []string{"bench", ""}

			for i, s := range sList[1:] {
				switch s {
				case "bench", "good", "bad", "all":
					datas[i] = s
				default:
					status = true
					log.Println("bad data")
					return false
				}
			}

			dataCh <- datas[0]
			dataCh <- datas[1]

			cmdCh <- "refresh"

			status = <- feedbackCh
			log.Println()

		case "f", "fetch" :
			cmdCh <- "fetch"

			status = <- feedbackCh
			log.Println()
		case "pau", "pause" :
			cmdCh <- "pause"
			status = <- feedbackCh
			log.Println()
		case "n", "next" :
			cmdCh <- "next"
			status = <- feedbackCh
			log.Println()
		case "p", "previous" :
			cmdCh <- "previous"
			status = <- feedbackCh
			log.Println()
		case "q", "quit" :
			cmdCh <- "quit"

			status = <- feedbackCh

			os.RemoveAll(tools.TempPath)
			os.MkdirAll(tools.TempPath, 0755)

			time.Sleep(100 * time.Millisecond)
			return true
		case "a", "auto" :
			cmdCh <- "auto"
			status = <- feedbackCh
			log.Println()
		case "m", "manual" :
			cmdCh <- "manual"
			status = <- feedbackCh
			log.Println()
//		case "help" :
//		case "print" :
//			cmdCh <- "print"
		case "clear", "clr" :

			switch tools.OSPlatform {
			case "linux":
				cmd := exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
			case "windows":
				cmd := exec.Command("cmd", "/c", "cls")
				cmd.Stdout = os.Stdout
				cmd.Run()
			}

			status = true
			log.Println()

//			goto getInput
		default:
			status = true
			log.Println("bad cmd:", sList[0])
//			goto getInput
		}

		return false
//	}
}
