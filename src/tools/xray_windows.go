//go:build windows
package tools

import(
	"os/exec"
	"strconv"
	"time"
	"log"
	"io"
)

type Xray struct {
	Port     int
	JsonPath string
	cmd      *exec.Cmd
}


func XrayDaemon(node *Node, cmdCh <-chan string, feedbackCh chan<- string) (error) {
	var x Xray

	node.Port = MainPort

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

	feedbackCh <- "running confirmed"

	for {
		select {
		case cmd := <- cmdCh :
			switch cmd {
			case "TERM" :
				err = x.Stop()
				if err != nil {
					log.Fatal(err)
				}
				feedbackCh <- "TERM confirmed";	 //"confirmed"
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
//	x.cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}		// Windows specific
//	x.cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}		// Linux specific
	stdout, err := x.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)
	return stdout, nil
}

func (x *Xray) Stop() error {

//	err := x.cmd.Wait()
//	if err != nil {
//		panic("process wait")
//		return fmt.Errorf("Stop:%w",err)
//	}
//	return nil

	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(x.cmd.Process.Pid))
	kill.Stderr = nil
	kill.Stdout = nil
	return kill.Run()
}

