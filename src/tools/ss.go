package tools

import (
	"fmt"
	"strings"
	"strconv"
	"encoding/base64"
	"log"
)

type SsShare struct{
	Method string
	Pwd string
	Addr string
	Port int
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
	var ssSh SsShare
	err := SsLinkToShare(&ssSh, ssShareLink)
	if err != nil {
		err = fmt.Errorf("ssLinkToShare:", err)
		return err
	}

	ser := server_ {
		Addr: ssSh.Addr,
		Port: ssSh.Port,
		Method: ssSh.Method,
		Pwd: ssSh.Pwd,
		Level: 0,
		}
	ssSet := ssSettings { Servers: []server_{ ser } }

	ss.Protocol = "shadowsocks"
	ss.Settings = ssSet
	ss.Tag = "proxy"
	ss.StreamSettings.TcpSettings.Header.Type = "none"
	ss.Mx.Enabled = false

	return nil
}

func SsLinkToShare(ssShareP *SsShare, ssShareLink string) error{
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

	ssShareP.Method = method
	ssShareP.Pwd = pwd
	ssShareP.Addr = addr
	ssShareP.Port = port

	return nil
}

func linkToShareType0(link string) (string, string, string, int, error) {
	base64AndAddrPortEmail := strings.Split(link, "@") 
	addrPortAndEmail := strings.Split(base64AndAddrPortEmail[1], "#")
	addrAndPort := strings.Split(addrPortAndEmail[0], ":")

	methAndPwd, err := base64.StdEncoding.DecodeString(base64AndAddrPortEmail[0])
	if err != nil {
		err = fmt.Errorf("ERROR: linkToVmShare: base64Decode:", err)
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
		err = fmt.Errorf("ERROR: linkToVmShare: base64Decode:", err)
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

func SsRemoveDuplicateNodes(nodes *[]*Node) {
	var ssS SsShare
	var nodesNoDup []*Node
	var ssShareList []*SsShare
	var flag bool


	err := SsLinkToShare(&ssS, (*nodes)[0].ShareLink)
	if err != nil {
		log.Println("ERROR: SsRemoveDup: ssLinkToShare:", err)
	}

	nodesNoDup = append(nodesNoDup, (*nodes)[0])
	ssShareList = append(ssShareList, &ssS)

	for _, node := range (*nodes){
		var ssS2 SsShare
		flag = true
		SsLinkToShare(&ssS2, node.ShareLink)
		for _, ss := range ssShareList {
			//if ssShareCompare(ss, &ssS2) {
			if *ss == ssS2 {
				flag = false
				break
			}
		}
		if flag {
			nodesNoDup = append(nodesNoDup, node)
			ssShareList = append(ssShareList, &ssS2)
		}
	}

	(*nodes) = nodesNoDup

}
