package tools

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"fmt"
	"strings"
	"errors"

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
	Timeout	int
	ErrorInfo error
	Con       *Config
	ShareCon	interface{}
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

type ByTimeout []*Node

func (a ByTimeout) Len() int           { return len(a) }
func (a ByTimeout) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimeout) Less(i, j int) bool { return a[i].Timeout > a[j].Timeout }


func (n *Node) Init(ntype, shareLink string) {
	n.Type = ntype
	n.ShareLink = shareLink
}

func (n *Node) CreateJson(dirPath string, port int) error {
	err := (*n).genConfig()
	if err != nil {
		return fmt.Errorf("genConfig:%w", err)
	}

	name := uuid.NewString()
	s := []string{dirPath, name, ".json"}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(*(n.Con), "", "    ")
	if err != nil {
		return fmt.Errorf("MarshalIndent:%w", err)
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile:%w", err)
	}

	err = JsonChangePort(n.JsonPath, n.JsonPath, port)
	if err != nil {
		return fmt.Errorf("JsonChangePort:%w", err)
	}

	return nil
}

func (n *Node) CreateFinalJson(dirPath string, port int, name string) error {
	err := (*n).genConfig()
	if err != nil {
		err = fmt.Errorf("createConfig:", err)
		return err
	}
	GenFinalConfig(n.Con)

	s := []string{dirPath, name}
	n.JsonPath = strings.Join(s, "")

	byteValue, err := json.MarshalIndent(*(n.Con), "", "    ")
	if err != nil {
		log.Println(err)
		return err
	}

	err = ioutil.WriteFile(n.JsonPath, byteValue, 0644)
	if err != nil {
		log.Println(err)
		return err
	}

	err = JsonChangePort(n.JsonPath, n.JsonPath, port)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (n *Node) genConfig() error {
	var con Config
	switch n.Type{
		case "vmess": 
			var vmout Outbound
			var vmShare VmessShare
			err := VmlinkToVmshare(&vmShare, n.ShareLink)
			if err != nil {
				return fmt.Errorf("VmlinkToVmshare:%w", err)
			}
			n.ShareCon = vmShare
			err = VmLinkToVmOut(&vmout, n.ShareLink)
			if err != nil {
				return fmt.Errorf("VmLinkToVmOut:%w", err)
			}
			OutboundToTestConfig(&con, vmout)
			n.Con = &con
		case "ss":
			var ssout Outbound
			err := SSLinkToSSout(&ssout, n.ShareLink)
			if err != nil {
				return fmt.Errorf("SSLinkToSSout:%w", err)
			}
			OutboundToTestConfig(&con, ssout)
			n.Con = &con

		default :
			return errors.New("unknown node type")
	}

	return nil
}
