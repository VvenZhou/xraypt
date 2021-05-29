package tools

import (
	"strings"
	"encoding/base64"
	"encoding/json"
	"strconv"
)


type VmessShare struct {
	V string `json:"v"`
	Ps string `json:"ps"`
	Add string `json:"add"`
	Port string `json:"port"`
	Id string `json:"id"`
	Aid string `json:"aid"`
	Scy string `json:"scy"`
	Net string `json:"net"`
	Type string `json:"type"`
	Host string `json:"host"`
	Path string `json:"path"`
	Tls string `json:"tls"`
	Sni string `json:"sni"`
}


type VmessUser struct {
	Id string `json:"id"`
	AlterId int `json:"alterId, omitempty"`
	Security string `json:"security"`
	Level int `json:"level, omitempty"`
}

func VmLinkToVmOut(vmess *Outbound, vmShareLink string) {
	var vmShare VmessShare
	headAndTail := strings.Split(vmShareLink, "vmess://")
	data, _ := base64.StdEncoding.DecodeString(headAndTail[1])
	err := json.Unmarshal(data, &vmShare)
	if err != nil {
		return
	}

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
