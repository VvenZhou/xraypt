package tools

import(
	"log"
	"net"
	"os/exec"
	//"os"
	"time"
	"strings"
	"syscall"
	"strconv"
	"encoding/json"
	"io/ioutil"
)

type Node struct {
	Id string
	ShareLink string
	JsonPath string
	AvgDelay int
	Country string
	DLSpeed float64
	ULSpeed float64

	Con *Config
}

type ByDLSpeed []*Node
func (a ByDLSpeed) Len() int { return len(a) }
func (a ByDLSpeed) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDLSpeed) Less(i, j int) bool { return a[i].DLSpeed > a[j].DLSpeed } //actually this func should be More than

type ByULSpeed []*Node
func (a ByULSpeed) Len() int { return len(a) }
func (a ByULSpeed) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByULSpeed) Less(i, j int) bool { return a[i].ULSpeed > a[j].ULSpeed } //actually this func should be More than

type ByDelay []*Node
func (a ByDelay) Len() int { return len(a) }
func (a ByDelay) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDelay) Less(i, j int) bool { return a[i].AvgDelay < a[j].AvgDelay }

type Xray struct{
	Port int
	JsonPath string
	randPort bool
	cmd *exec.Cmd
}

func (n *Node) Init(id string, shareLink string) {
	n.Id = id
	n.ShareLink = shareLink
}

func (n *Node) CreateJson(dirPath string) {
	var vmout VmessOut
	var con Config
	VmLinkToVmOut(&vmout, n.ShareLink)
	VmOutToConfig(&con, vmout)

	n.Con = &con

	s := []string{dirPath, n.Id, ".json"}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(con, "", "    ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		log.Println(err)
	}
}

func (n *Node) CreateFinalJson(dirPath string) {
	VmConfigFinal(n.Con)

	s := []string{dirPath, n.Id, ".json"}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(*(n.Con), "", "    ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		log.Println(err)
	}
}

func (x *Xray) Init(port int, jsonPath string) error {
	x.Port = port
	x.JsonPath = jsonPath
	x.randPort = false

	return nil

	//s := []string{"/tmp/tmp_", jsonPath}
	//x.JsonPath = strings.Join(s, "")
	//err := JsonChangePort(jsonPath, x.JsonPath, x.Port)
	//if err != nil {
	//	return err
	//}
	//return nil
}

func (x *Xray) Run(randPort bool) error {
	if randPort {
		//log.Println("randport")
		x.randPort = true
		x.Port, _ = GetFreePort()
		s := []string{"temp/", "xrayRun_port_", strconv.Itoa(x.Port), ".json"}

		path := strings.Join(s, "")
		err := JsonChangePort(x.JsonPath, path, x.Port)
		if err != nil {
			return err
		}
		x.JsonPath = path
	}
	x.cmd = exec.Command("tools/xray", "-c", x.JsonPath)
	x.cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	//stdout, _ := cmd.StdoutPipe()

	err := x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	//log.Printf("xray executing, using json: %s\n", x.JsonPath)
	time.Sleep(500 * time.Millisecond)
	//log.Println("xray started!")
	//go print(stdout)
	return nil
}

func (x *Xray) Stop() error {
	err := x.cmd.Process.Signal(syscall.SIGTERM)

	//err := x.cmd.Process.Kill()
	//if err != nil {
	//	log.Fatal(err)
	//}

	_, err = x.cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	//log.Println(ps.Success())

	if x.randPort {
		//os.Remove(x.JsonPath)
		//err = os.Remove(x.JsonPath)
		//if err != nil {
		//	log.Fatal(err)
		//}
	}

	//log.Println("xray stopped.")
	return nil
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

func RemoveDuplicateStr(intSlice []string) []string {
    keys := make(map[string]bool)
    list := []string{}

    for _, entry := range intSlice {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }
    return list
}
