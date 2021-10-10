package tools

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
)

type Node struct {
	Id        string
	Type	string		// "vmess", "vless", etc.
	ShareLink string 	//without head("vmess://")
	JsonPath  string
	AvgDelay  int
	Country   string
	DLSpeed   float64
	ULSpeed   float64
	Port      int
	Con       *Config
}

type Xray struct {
	Port     int
	JsonPath string
	cmd      *exec.Cmd
}

type ByDLSpeed []*Node

func (a ByDLSpeed) Len() int           { return len(a) }
func (a ByDLSpeed) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDLSpeed) Less(i, j int) bool { return a[i].DLSpeed > a[j].DLSpeed } //actually this func should be More than

type ByULSpeed []*Node

func (a ByULSpeed) Len() int           { return len(a) }
func (a ByULSpeed) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByULSpeed) Less(i, j int) bool { return a[i].ULSpeed > a[j].ULSpeed } //actually this func should be More than

type ByDelay []*Node

func (a ByDelay) Len() int           { return len(a) }
func (a ByDelay) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDelay) Less(i, j int) bool { return a[i].AvgDelay < a[j].AvgDelay }

func (n *Node) Init(id, ntype, shareLink string) {
	n.Id = id
	n.Type = ntype
	n.ShareLink = shareLink
}

func (n *Node) CreateJson(dirPath string) {
	var con Config
	switch n.Type{
		case "vmess": 
			var vmout Outbound
			VmLinkToVmOut(&vmout, n.ShareLink)
			OutboundToTestConfig(&con, vmout)
			n.Con = &con

		default :
			log.Println("ERROR: unknown node type")
			return 
	}

	out := uuid.New().String()
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
	GenFinalConfig(n.Con)

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
	return nil
}

func (x *Xray) Run() error {
	x.cmd = exec.Command(XrayPath, "-c", x.JsonPath)

	err := x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(750 * time.Millisecond)
	return nil
}

func (x *Xray) Stop() error {
	err := x.cmd.Process.Signal(syscall.SIGTERM)

	_, err = x.cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

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

//func RemoveDuplicateStr(intSlice []string) []string {
//	keys := make(map[string]bool)
//	list := []string{}
//
//	for _, entry := range intSlice {
//		if _, value := keys[entry]; !value {
//			keys[entry] = true
//			list = append(list, entry)
//		}
//	}
//	return list
//}
