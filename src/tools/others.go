package tools

import(
	"log"
	"net"
	"os/exec"
)

func RunXray(jsonPath string) (*exec.Cmd) {
	cmd := exec.Command("tools/xray", "-c", jsonPath)
	//stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("xray executing, using json: %s\n", jsonPath)
	//go print(stdout)
	return cmd
}

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
