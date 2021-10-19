package tools

import (
	//"log"
	"fmt"
	"strings"
	"strconv"
	"encoding/base64"
)

type ssShare struct{
	method string
	pwd string
	addr string
	port int
}

type server_ struct {
	Addr string `json:"address"`
	Port int `json:"port"`
	Method string `json:"method"`
	Pwd string `json:"password"`
	Level int `json:"level"`
}

type ssSettings struct {
	Servers []server_ `json:"servers"`
}

func SSLinkToSSout(ss *Outbound, ssShareLink string) error {
	var ssSh ssShare
	ssLinkToShare(&ssSh, ssShareLink)

	ser := server_ {
		Addr: ssSh.addr,
		Port: ssSh.port,
		Method: ssSh.method,
		Pwd: ssSh.pwd,
		Level: 0,
		}
	ssSet := ssSettings { Servers: []server_{ ser } }

	ss.Protocol = "shadowsocks"
	ss.Settings = ssSet
	ss.Tag = "proxy"
	ss.StreamSettings.TcpSettings.Header.Type = "none"
	ss.Mx.Enabled = false

	//log.Println(ss)

	return nil
}

func ssLinkToShare(ssShareP *ssShare, ssShareLink string) error{
	//log.Println(ssShareLink)
	strList0 := strings.Split(ssShareLink, "@") 
	strList1 := strings.Split(strList0[1], "#")
	strList2 := strings.Split(strList1[0], ":")

	methAndPwd, err := base64.StdEncoding.DecodeString(strList0[0])
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: base64Decode:", err)
		return err
	}

	strList3 := strings.Split(string(methAndPwd), ":")

	ssShareP.method = strList3[0]
	ssShareP.pwd = strList3[1]
	ssShareP.addr = strList2[0]

	port,err := strconv.Atoi(strList2[1])
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: base64Decode:", err)
		return err
	}
	ssShareP.port = port

	//log.Println(ssShareP)

	return nil
}

func SsRemoveDulpicate(ssLinks []string) []string {
	var ssS ssShare
	var ssNoDup []string
	var ssShareList []*ssShare
	var flag bool

	ssLinkToShare(&ssS, ssLinks[0])
	ssNoDup = append(ssNoDup, ssLinks[0])
	ssShareList = append(ssShareList, &ssS)

	for _, ssL := range ssLinks{
		var ssS2 ssShare
		flag = true
		ssLinkToShare(&ssS2, ssL)
		for _, ss := range ssShareList {
			//if ssShareCompare(ss, &ssS2) {
			if *ss == ssS2 {
				flag = false
				break
			}
		}
		if flag {
			ssNoDup = append(ssNoDup, ssL)
			ssShareList = append(ssShareList, &ssS2)
		}
	}
	return ssNoDup
}
