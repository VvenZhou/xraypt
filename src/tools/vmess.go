package tools

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"fmt"
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

//type VmessUser struct {
//	Id string `json:"id"`
//	AlterId int `json:"alterId, omitempty"`
//	Security string `json:"security"`
//	Level int `json:"level, omitempty"`
//}

type User_ struct {
	Id string `json:"id"`
	AlterId int `json:"alterId, omitempty"`
	Security string `json:"security"`
	Level int `json:"level, omitempty"`
}

type Vnext_ struct {
	Address string `json:"address"`
	Port int `json:"port"`
	//Users []interface{} `json:"users"`
	Users []User_ `json:"users"`
}

type VmessSettings struct {
	Vnext []Vnext_ `json:"vnext"`
}

func VmLinkToVmOut(vmess *Outbound, vmShareLink string) error {
	var vmShare VmessShare

	err := vmlinkToVmshare(&vmShare, vmShareLink)
	if err != nil {
		return err
	}

	port, _ := strconv.Atoi(vmShare.Port)
	aid, _ := strconv.Atoi(vmShare.Aid)
	user := User_{Id: vmShare.Id, AlterId: aid,  Security: "auto", Level: 0}
	vnext := Vnext_ { Address: vmShare.Add, 
			  Port: port,
			  Users: []User_{ user },
			}
	vmSettings := VmessSettings { Vnext: []Vnext_{ vnext }}

	vmess.Protocol = "vmess"
	vmess.Settings = vmSettings
	//(*vmess).Settings.Vnext = []struct{Address string `json:"address"`; Port int `json:"port"`; Users []interface{} `json:"users"`}{{
	//	Address: vmShare.Add,
	//	Port: port,
	//	Users: []interface{}{ user}}}
	vmess.Tag = "proxy"
	vmess.StreamSettings.TcpSettings.Header.Type = "none"
	if vmShare.Net == "tcp" {
		vmess.StreamSettings.Network = "tcp"
		vmess.StreamSettings.TcpSettings.Header.Type = "none"
	}
	if vmShare.Net == "ws" {
		vmess.StreamSettings.Network = "ws"
		vmess.StreamSettings.WsSettings.Path = vmShare.Path
		vmess.StreamSettings.WsSettings.Headers.Host = vmShare.Host
	}
	if vmShare.Tls == "tls" {
		vmess.StreamSettings.Security = "tls"
		vmess.StreamSettings.TlsSettings.ServerName = vmShare.Host
		vmess.StreamSettings.TlsSettings.AllowInsecure = true
	}
	vmess.Mx.Enabled = false

	return nil
}

func vmlinkToVmshare(vmShareP *VmessShare, vmLink string) error {
	var i interface{}

	data, err := base64.StdEncoding.DecodeString(vmLink)
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: base64Decode:", err)
		return err
	}

	err = json.Unmarshal(data, &i)
	if err != nil {
		err = fmt.Errorf("ERROR: vmlinkToVmShare: jsonUnmarshal:", err)
		return err
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

	return nil
}

func VmRemoveDulpicate(vmLinks []string) []string {
	var vmS VmessShare
	var vmNoDup []string
	var vmShare []*VmessShare
	var flag bool

	vmlinkToVmshare(&vmS, vmLinks[0])
	vmNoDup = append(vmNoDup, vmLinks[0])
	vmShare = append(vmShare, &vmS)

	for _, vmL := range vmLinks{
		var vmS2 VmessShare
		flag = true
		vmlinkToVmshare(&vmS2, vmL)
		for _, vm := range vmShare {
			//if vmShareCompare(vm, &vmS2) {
			if *vm == vmS2 {
				flag = false
				break
			}
		}
		if flag {
			vmNoDup = append(vmNoDup, vmL)
			vmShare = append(vmShare, &vmS2)
		}
	}
	return vmNoDup
}

//Not used, yet.
func vmShareCompare(a, b *VmessShare) bool {

	if a.Add == b.Add && a.Port == b.Port && a.Id == b.Id && a.Aid == b.Aid && 
		a.Net == b.Net && a.Host == b.Host && a.Path == b.Path && a.Tls == b.Tls {
		return true
	}

	return false
}
