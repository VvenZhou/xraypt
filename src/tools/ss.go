package tools

import (
	"fmt"
	"strings"
	"errors"
	"regexp"
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
	ssSh, err := SsLinkToShare(ssShareLink)
	if err != nil {
		return fmt.Errorf("SsLinkToShare:%w", err)
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

func SsLinkToShare(ssShareLink string) (*SsShare, error) {
	var ssSh *SsShare
	o, err := linkToShareType0(ssShareLink)
	if err != nil {
		if !errors.Is(err, FormatErr) {
			return nil, err
		}
	}else{
		ssSh = o
		return ssSh, nil
	}

	o, err = linkToShareType1(ssShareLink)
	if err != nil {
		if !errors.Is(err, FormatErr) {
			return nil, err
		}
	}else{
		ssSh = o
		return ssSh, nil
	}

//	panic(ssShareLink)

	return nil, fmt.Errorf("%d: %w", ssShareLink, FormatErr)
}

func linkToShareType0(link string) (*SsShare, error) {		//Type: (base64)@(addr):(port)#comments
	t0 := regexp.MustCompile(`([a-zA-Z0-9=]*)@(.*?):([0-9]*)#.*`)
	result := t0.FindStringSubmatch(link)
	if len(result) != 4 {
		return nil, FormatErr
	}

	var base64DecodedBytes []byte
	base64DecodedBytes, err := base64.StdEncoding.DecodeString(result[1])
	if err != nil {
		base64DecodedBytes, err = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(result[1])
		if err != nil {
			err = fmt.Errorf("base64Decode:%d, %w", result[1], err)
			return nil, err
		}
	}

	methAndPwd:= strings.Split(string(base64DecodedBytes), ":")
	if len(methAndPwd) != 2 {
		return nil, FormatErr
	}

	port, err := strconv.Atoi(result[3])
	if err != nil {
		err = fmt.Errorf("strconv.Atoi:%w", err)
		return nil, err
	}

	var ssSh SsShare
	ssSh.Method = methAndPwd[0] 
	ssSh.Pwd = methAndPwd[1]
	ssSh.Addr = result[2]
	ssSh.Port = port

	return &ssSh, nil
}

func linkToShareType1(link string) (*SsShare, error) {		//Type: (base64)#(comments)
	t1 := regexp.MustCompile(`([a-zA-Z0-9=]*)#.*`)
	result := t1.FindStringSubmatch(link)
	if len(result) != 2 {
		return nil, FormatErr
	}

	var base64DecodedBytes []byte
	base64DecodedBytes, err := base64.StdEncoding.DecodeString(result[1])
	if err != nil {
		base64DecodedBytes, err = base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(result[1])
		if err != nil {
			err = fmt.Errorf("base64Decode:%d, %w", result[1], err)
			return nil, err
		}
	}

	//Format 0
	f0 := regexp.MustCompile(`(.*?):(.*?)@(.*?):([0-9]*)`)
	result = f0.FindStringSubmatch(string(base64DecodedBytes))
	if len(result)==5 {
		port, err := strconv.Atoi(result[4])
		if err != nil {
			err = fmt.Errorf("strconv.Atoi:%w", err)
			return nil, err
		}

		var ssSh SsShare
		ssSh.Method = result[1]
		ssSh.Pwd = result[2]
		ssSh.Addr = result[3]
		ssSh.Port = port

		return &ssSh, nil
	}

	return nil, FormatErr
}

func SsRemoveDuplicateNodes(nodes *[]*Node) {
	var ssS *SsShare
	var nodesNoDup []*Node
	var ssShareList []*SsShare
	var flag bool


	ssS, err := SsLinkToShare((*nodes)[0].ShareLink)
	if err != nil {
		log.Println("[Error]: SsRemoveDup: SsLinkToShare:", err)
	}

	nodesNoDup = append(nodesNoDup, (*nodes)[0])
	ssShareList = append(ssShareList, ssS)

	for _, node := range (*nodes){
		flag = true
		ssS2, err := SsLinkToShare(node.ShareLink)
		if err != nil {
			log.Println("[Error]: SsRemoveDup: SsLinkToShare:", err)
			continue
		}

		for _, ss := range ssShareList {
			//if ssShareCompare(ss, &ssS2) {
			if *ss == *ssS2 {
				flag = false
				break
			}
		}
		if flag {
			nodesNoDup = append(nodesNoDup, node)
			ssShareList = append(ssShareList, ssS2)
		}
	}

	(*nodes) = nodesNoDup

}
