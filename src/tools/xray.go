package tools

import(
	"os/exec"
	"syscall"
//	"fmt"
	"time"
	"log"
	"io"
)

type Xray struct {
	Port     int
	JsonPath string
	cmd      *exec.Cmd
}


func XrayDaemon(node *Node, cmdCh <-chan string, feedback chan<- int) (error) {
	var x Xray
	node.Port = MainPort
	node.CreateJson(TempPath)
	err := x.Init(MainPort, (*node).JsonPath)
	if err != nil {
		log.Fatal(err)
	}
	//stdout, err := x.Run()
	_, err = x.Run()
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case cmd := <- cmdCh :
			switch cmd {
			case "TERM" :
				err = x.Stop()
				if err != nil {
					log.Fatal(err)
				}
				feedback <- 0;	 //"confirmed"
				return nil
			}
//		default :
//			buff := make([]byte, 10)
//			var n int
//			n, err = stdout.Read(buff)
//			if n >0 {
//				fmt.Printf(string(buff[:n]))
//			}
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
	stdout, err := x.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(750 * time.Millisecond)
	return stdout, nil
}

func (x *Xray) Stop() error {
	err := x.cmd.Process.Signal(syscall.SIGTERM)

	_, err = x.cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

