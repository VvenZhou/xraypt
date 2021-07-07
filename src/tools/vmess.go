package tools

import (
	"strings"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"log"
)


type VmessShare struct {
	//V string `json:"v"`
	//Ps string `json:"ps"`
	Add string `json:"add"`
	Port string `json:"port"`
	Id string `json:"id"`
	Aid string `json:"aid"`
	//Scy string `json:"scy"`
	Net string `json:"net"`
	//Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	Tls string `json:"tls"`
	//Sni string `json:"sni"`
}

type VmessUser struct {
	Id string `json:"id"`
	AlterId int `json:"alterId, omitempty"`
	Security string `json:"security"`
	Level int `json:"level, omitempty"`
}

func VmLinkToVmOut(vmess *Outbound, vmShareLink string) {
	var vmShare VmessShare
	VmFillVmShare(&vmShare, vmShareLink)
	//headAndTail := strings.Split(vmShareLink, "vmess://")
	//data, _ := base64.StdEncoding.DecodeString(headAndTail[1])
	//err := json.Unmarshal(data, &vmShare)
	//if err != nil {
	//	//return
	//}

	//var vmess VmessOut

	port, _ := strconv.Atoi(vmShare.Port)
	aid, _ := strconv.Atoi(vmShare.Aid)
	user := VmessUser{Id: vmShare.Id, AlterId: aid,  Security: "auto", Level: 0}

	(*vmess).Protocol = "vmess"
	(*vmess).Settings.Vnext = []struct{Address string `json:"address"`; Port int `json:"port"`; Users []interface{} `json:"users"`}{{
		Address: vmShare.Add,
		Port: port,
		Users: []interface{}{ user}}}
	(*vmess).Tag = "proxy"
	(*vmess).StreamSettings.TcpSettings.Header.Type = "none"
	if vmShare.Net == "tcp" {
		(*vmess).StreamSettings.Network = "tcp"
		(*vmess).StreamSettings.TcpSettings.Header.Type = "none"
	}
	if vmShare.Net == "ws" {
		(*vmess).StreamSettings.Network = "ws"
		(*vmess).StreamSettings.WsSettings.Path = vmShare.Path
		(*vmess).StreamSettings.WsSettings.Headers.Host = vmShare.Host
	}
	if vmShare.Tls == "tls" {
		(*vmess).StreamSettings.Security = "tls"
		(*vmess).StreamSettings.TlsSettings.ServerName = vmShare.Host
		(*vmess).StreamSettings.TlsSettings.AllowInsecure = true
	}
	(*vmess).Mx.Enabled = true
	//return vmess
}

func VmFillVmShare(vmShareP *VmessShare, vmLink string) {
	var i interface{}
	headAndTail := strings.Split(vmLink, "vmess://")
	data, _ := base64.StdEncoding.DecodeString(headAndTail[1])
	err := json.Unmarshal(data, &i)
	if err != nil {
		log.Println(err)
		log.Println("error link:", vmLink)
		return
	}
	m := i.(map[string]interface{})

	vmShareP.Add = m["add"].(string)
	if i, ok := m["port"].(float64); ok {
		flt := strconv.FormatFloat(i, 'f', -1, 64)
		vmShareP.Port = flt
	}else{
		vmShareP.Port = m["port"].(string)
	}
	vmShareP.Id = m["id"].(string)
	if i, ok := m["aid"].(float64); ok {
		flt := strconv.FormatFloat(i, 'f', -1, 64)
		vmShareP.Aid = flt
	}else if s, ok := m["aid"].(string); ok {
		vmShareP.Aid = s
	}
	if s, ok := m["net"].(string); ok {
		vmShareP.Net = s
	}
	if s, ok := m["host"].(string); ok {
		vmShareP.Host = s
	}
	if s, ok := m["path"].(string); ok {
		vmShareP.Path = s
	}
	if s, ok := m["tls"].(string); ok {
		vmShareP.Tls = s
	}
}

func VmRemoveDulpicate(vmLinks []string) []string {
	var vmNoDup []string
	var vmShare []*VmessShare
	var flag bool
	for _, vmL := range vmLinks{
		flag = true
		var vmS VmessShare
		VmFillVmShare(&vmS, vmL)
		if len(vmNoDup) == 0 {
			vmNoDup = append(vmNoDup, vmL)
			vmShare = append(vmShare, &vmS)
			continue
		}
		for _, vm := range vmShare {
			if (*vm) == vmS {
				flag = false
				break
			}
		}
		if flag {
			vmNoDup = append(vmNoDup, vmL)
			vmShare = append(vmShare, &vmS)
		}
	}
	return vmNoDup
}
