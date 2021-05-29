package tools

import (
	//"strconv"
//	"strings"
	"fmt"
	"regexp"
)

type VlessShare struct{
	Id string
	Addr string
	Port int
	Scy string
}

type VlessUser struct {
	Id string `json:"id"`
	Encryption string `json:"encryption"`
	Level int `json:"level, omitempty"`
}

func VlLinkToOut(vless *Outbound, vlShareLink string) {
	//vl := strings.Split(vlShareLink, "&")
	//for _, s := range vl{
	//	fmt.Printf("%s\n", s)
	//}
	encryRe := regexp.MustCompile(`encryption=(.*?)[&]`)
	encry := encryRe.FindStringSubmatch(vlShareLink)
	fmt.Println(encry)


	//user := VlessUser{Id: vlList[1], Encryption: "none", Level: 0}
	//port, _ := strconv.Atoi(vlList[3])

	//(*vless).Protocol = "vless"
	//(*vless).Settings.Vnext = []struct{Address string `json:"address"`; Port int `json:"port"`; Users []interface{} `json:"users"`}{{
	//	Address: vlList[2],
	//	Port: port,
	//	Users: []interface{}{ user}}}
	//(*vless).Tag = "proxy"
	//(*vless).StreamSettings.TcpSettings.Header.Type = "none"
	//if vlList[5] == "tcp" {
	//	(*vless).StreamSettings.Network = "tcp"
	//	(*vless).StreamSettings.TcpSettings.Header.Type = "none"
	//}
	//if vlList[5] == "ws" {
	//	(*vless).StreamSettings.Network = "ws"
	//	(*vless).StreamSettings.WsSettings.Path = vlList[7]
	//	(*vless).StreamSettings.WsSettings.Headers.Host = vlList[6]
	//}
	//if vlList[4] == "tls" {
	//	(*vless).StreamSettings.Security = "tls"
	//	(*vless).StreamSettings.TlsSettings.ServerName = vlList[6]
	//	(*vless).StreamSettings.TlsSettings.AllowInsecure = true
	//}
	//(*vless).Mx.Enabled = true
}
