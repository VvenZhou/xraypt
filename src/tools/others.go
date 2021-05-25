package tools

import(
	"log"
	"net"
	"os/exec"
	"os"
	"time"
	"strings"
	"syscall"
)

type Xray struct{
	Port int
	JsonPath string

	cmd *exec.Cmd
}

func (x *Xray) Init(jsonPath string) error {
	x.Port, _ = GetFreePort()

	s := []string{"/tmp/tmp_", jsonPath}
	x.JsonPath = strings.Join(s, "")
	err := JsonChangePort(jsonPath, x.JsonPath, x.Port)
	if err != nil {
		return err
	}
	return nil
}

func (x *Xray) Run() error {
	x.cmd = exec.Command("tools/xray", "-c", x.JsonPath)
	x.cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	//stdout, _ := cmd.StdoutPipe()

	err := x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("xray executing, using json: %s\n", x.JsonPath)
	time.Sleep(500 * time.Millisecond)
	log.Println("xray started!")
	//go print(stdout)
	return nil

}

func (x *Xray) Stop() error {
	err := x.cmd.Process.Signal(syscall.SIGTERM)

	//err := x.cmd.Process.Kill()
	//if err != nil {
	//	log.Fatal(err)
	//}

	ps, err := x.cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ps.Success())

	err = os.Remove(x.JsonPath)
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println("xray stopped.")
	return nil
}

//func RunXray(jsonPath string) (*exec.Cmd) {
//	cmd := exec.Command("tools/xray", "-c", jsonPath)
//	//stdout, _ := cmd.StdoutPipe()
//
//	err := cmd.Start()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Printf("xray executing, using json: %s\n", jsonPath)
//	time.Sleep(500 * time.Millisecond)
//	log.Println("xray started!")
//	//go print(stdout)
//	return cmd
//}

//func print(stdout io.ReadCloser) {
//	for {
//		r := bufio.NewReader(stdout)
//		line, _, _ := r.ReadLine()
//		fmt.Println(string(line))
//	}
//}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}
