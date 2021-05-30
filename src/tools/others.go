package tools

import(
	"log"
	"net"
	"os/exec"
	"time"
	"strings"
	"syscall"
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
	Port int
	Con *Config
}

type Xray struct{
	Port int
	JsonPath string
	cmd *exec.Cmd
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


func (n *Node) Init(id string, shareLink string, port int) {
	n.Id = id
	n.ShareLink = shareLink
	n.Port = port
}

func (n *Node) CreateJson(dirPath string) {
	var vmout Outbound
	var con Config
	VmLinkToVmOut(&vmout, n.ShareLink)
	OutToConfig(&con, vmout)

	n.Con = &con

	out, err := exec.Command("uuidgen").Output()
	if err != nil {
		log.Fatal(err)
	}
	s := []string{dirPath, strings.TrimSpace(string(out)), ".json"}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(con, "", "    ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		log.Println(err)
	}

	err = JsonChangePort(n.JsonPath, n.JsonPath, n.Port)
	if err != nil {
		log.Println(err)
	}
}

func (n *Node) CreateFinalJson(dirPath string, name string) {
	ConfigFinal(n.Con)

	//s := []string{dirPath, n.Id, ".json"}
	s := []string{dirPath, name, ".json"}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(*(n.Con), "", "    ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		log.Println(err)
	}

	err = JsonChangePort(n.JsonPath, n.JsonPath, 8123)
	if err != nil {
		log.Println(err)
	}
}

func (x *Xray) Init(port int, jsonPath string) error {
	x.Port = port
	x.JsonPath = jsonPath
	//x.randPort = false

	//s := []string{"temp/xrayRun_port_", strconv.Itoa(x.Port), ".json"}
	//x.JsonPath = strings.Join(s, "")
	//err := JsonChangePort(jsonPath, jsonPath, x.Port)
	//if err != nil {
	//	return err
	//}
	return nil
}

func (x *Xray) Run() error {
	//if randPort {
	//	//log.Println("randport")
	//	x.randPort = true
	//	x.Port, _ = GetFreePort()
	//	s := []string{"temp/", "xrayRun_port_", strconv.Itoa(x.Port), ".json"}

	//	path := strings.Join(s, "")
	//	err := JsonChangePort(x.JsonPath, path, x.Port)
	//	if err != nil {
	//		return err
	//	}
	//	x.JsonPath = path
	//}

	//log.Println("runnning Xray: ", x.JsonPath, " at ", x.Port)
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

	//if x.randPort {
	//	//os.Remove(x.JsonPath)
	//	//err = os.Remove(x.JsonPath)
	//	//if err != nil {
	//	//	log.Fatal(err)
	//	//}
	//}

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

func GetFreePorts(count int) ([]int, error) {
	var ports []int
	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
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
