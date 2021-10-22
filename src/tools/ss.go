package tools

import (
	"fmt"
	"strings"
	"strconv"
	"encoding/base64"
	"log"
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
	err := ssLinkToShare(&ssSh, ssShareLink)
	if err != nil {
		err = fmt.Errorf("ssLinkToShare:", err)
		return err
	}

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
	var method, pwd, addr string
	var port int
	var err error

	f := strings.Contains(ssShareLink, "@")
	if f {
		method, pwd, addr, port, err = linkToShareType0(ssShareLink)
		if err != nil {
			err = fmt.Errorf("linkToShareType0:", err)
			return err
		}
	}else{
		method, pwd, addr, port, err = linkToShareType1(ssShareLink)
		if err != nil {
			err = fmt.Errorf("linkToShareType1:", err)
			return err
		}
	}

	ssShareP.method = method
	ssShareP.pwd = pwd
	ssShareP.addr = addr
	ssShareP.port = port

	return nil
}

func linkToShareType0(link string) (string, string, string, int, error) {
	base64AndAddrPortEmail := strings.Split(link, "@") 
	addrPortAndEmail := strings.Split(base64AndAddrPortEmail[1], "#")
	addrAndPort := strings.Split(addrPortAndEmail[0], ":")

	methAndPwd, err := base64.StdEncoding.DecodeString(base64AndAddrPortEmail[0])
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: base64Decode:", err)
		return "", "", "", 0, err
	}

	strList3 := strings.Split(string(methAndPwd), ":")

	method := strList3[0]
	pwd := strList3[1]
	addr := addrAndPort[0]
	port, err := strconv.Atoi(addrAndPort[1])
	if err != nil {
		err = fmt.Errorf("strconv.Atoi:", err)
		return "", "", "", 0, err
	}

	return method, pwd, addr, port, nil
}

func linkToShareType1(link string) (string, string, string, int, error) {
	base64AndEmail := strings.Split(link, "#")
	base64DecodedStr, err := base64.StdEncoding.DecodeString(base64AndEmail[0])
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: base64Decode:", err)
		return "", "", "", 0, err
	}

	methPAndAddrP := strings.Split(string(base64DecodedStr), "@") 
	methodAndPwd := strings.Split(methPAndAddrP[0], ":")
	addrAndPort := strings.Split(methPAndAddrP[1], ":")

	method := methodAndPwd[0]
	pwd := methodAndPwd[1]
	addr := addrAndPort[0]
	port, err := strconv.Atoi(addrAndPort[1])
	if err != nil {
		err = fmt.Errorf("strconv.Atoi:", err)
		return "", "", "", 0, err
	}

	return method, pwd, addr, port, nil
}

func SsRemoveDulpicate(ssLinks []string) []string {
	var ssS ssShare
	var ssNoDup []string
	var ssShareList []*ssShare
	var flag bool

	err := ssLinkToShare(&ssS, ssLinks[0])
	if err != nil {
		log.Println("ERROR: SsRemoveDup: ssLinkToShare:", err)
	}
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
