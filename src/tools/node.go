package tools

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"

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

func (n *Node) Init(ntype, shareLink string) {
	//n.Id = id
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
		case "ss":
			var ssout Outbound
			err := SSLinkToSSout(&ssout, n.ShareLink)
			if err != nil {
				log.Println("ERROR: CreateJson: SSLinkToSSout:", err)
			}
			OutboundToTestConfig(&con, ssout)
			n.Con = &con
			//log.Println(con)

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

