package tools

import(
	//"syscall"
	"os"
	"os/exec"
	"time"
	"fmt"
	"log"
	"io"
)
//build linux

type Xray struct {
	Port     int
	JsonPath string
	cmd      *exec.Cmd
}


func XrayDaemon(node *Node, cmdCh <-chan string, feedbackCh chan<- string) (error) {
	var x Xray

//	node.Port = MainPort

	node.CreateFinalJson(OutPath, MainPort, "cur.json")
	err := x.Init(MainPort, node.JsonPath)
	if err != nil {
		log.Fatal(err)
	}

//	stdout, err := x.Run()
	_, err = x.Run()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	feedbackCh <- "running"

	for {
		select {
		case cmd := <- cmdCh :
			switch cmd {
			case "TERM" :
				err = x.Stop()
				if err != nil {
					log.Fatal(err)
				}
				feedbackCh <- "TERM";	 //"confirmed"
				return nil
			}
//		default :
//			buff := make([]byte, 10)
//			var n int
//			n, err = stdout.Read(buff)
//			if n >0 {
//				fmt.Printf(string(buff[:n]))
//			}
			//feedback <- "running"
		}
	}	
}

func (x *Xray) Init(port int, jsonPath string) error {
	x.Port = port
	x.JsonPath = jsonPath
	return nil
}

func (x *Xray) Run() (io.ReadCloser, error) {
	x.cmd = exec.Command(XrayPath, "-c", x.JsonPath)
//	x.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}		// Linux specifical
	stdout, err := x.cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("StdoutPipe:%w",err)
	}

	err = x.cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("Start:%w",err)
	}

	time.Sleep(500 * time.Millisecond)
	return stdout, nil
}

func (x *Xray) Stop() error {
//	syscall.Kill(-x.cmd.Process.Pid, syscall.SIGTERM)
	err := x.cmd.Process.Signal(os.Interrupt)		//do not use syscall(os specifical)
	if err != nil {
		log.Println(err)
		x.cmd.Process.Kill()
	}

	err = x.cmd.Wait()
	if err != nil {
		panic("cmd wait")
		return fmt.Errorf("Stop:%w",err)
	}
	return nil
}

