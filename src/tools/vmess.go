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

type VmessOut struct {
	Protocol string `json:"protocol"`
	Settings struct {
		Vnext []struct {
			Address string `json:"address"`
			Port int `json:"port"`
			Users []User `json:"users"`
			//Users []struct {
			//	Id string `json:"id"`
			//	AlterId int `json:"alterId, omitempty"`
			//	Security string `json:"security"`
			//	Level int `json:"level, omitempty"`
			//} `json:"users"`
		} `json:"vnext"`
	} `json:"settings"`
	Tag string `json:"tag"`
	StreamSettings struct {
		Network string `json:"network"`
		Security string `json:"security"`
		TlsSettings struct {
			ServerName string `json:"servername,omitempty"`
			AllowInsecure bool `json:"allowInsecure,omitempty"`
		} `json:"tlsSettings,omitempty"`
		WsSettings struct {
			Path string `json:"path,omitempty"`
			Headers struct { Host string `json:"Host,omitempty"` } `json:"headers,omitempty"`
		} `json:"wsSettings,omitempty"`
		TcpSettings struct {
			Header struct { Type string `json:"type,omitempty"`} `json:"header,omitempty"`
		} `json:"tcpSettings,omitempty"`

	} `json:"streamSettings"`
	Mx struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"mux,omitempty"`
}

type User struct {
	Id string `json:"id"`
	AlterId int `json:"alterId, omitempty"`
	Security string `json:"security"`
	Level int `json:"level, omitempty"`
}

func VmLinkToVmOut(vmess *VmessOut, vmShareLink string) {
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
	user := User{Id: vmShare.Id, AlterId: aid,  Security: "auto", Level: 0}

	(*vmess).Protocol = "vmess"
	(*vmess).Settings.Vnext = []struct{Address string `json:"address"`; Port int `json:"port"`; Users []User `json:"users"`}{{
		Address: vmShare.Add,
		Port: port,
		Users: []User{ user }}}
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

func VmOutToConfig(con *Config, vmOut VmessOut ) {
	htIn := HttpIn{Tag: "http-in", Listen: "::", Port: 8123, Protocol: "http"}
	soIn := SocksIn{Tag: "socks-in", Port: 1080, Listen: "::", Protocol: "socks", Settings: struct{Auth string `json:"auth"`; Ip string `json:"ip"`; Udp bool `json:"udp"`}{Auth: "noauth", Udp: true, Ip: "127.0.0.1"}}
	//var con Config
	(*con).Log.Loglevel = "error"
	(*con).Inbounds = []interface{}{ htIn, soIn }
	(*con).Outbounds = []interface{}{ vmOut }
	(*con).Dns.Servers = []interface{}{ "8.8.8.8", "1.1.1.1" }
	(*con).Routing.DomainStrategy = "AsIs"
	//(*con).Routing.Rules = []struct{Type string `json:"type"`; Domain []string `json:"domain"`; OutboundTag string `json:"outboundTag"`}{{Type: "field", Domain: []string{"geosite:google", "domain:speedtest.com"}, OutboundTag: "proxy"}}
	(*con).Routing.Rules = []struct{Type string `json:"type"`; Domain []string `json:"domain"`; OutboundTag string `json:"outboundTag"`}{}
}

func VmConfigFinal(con *Config) {

	type DnsServer struct {
		Address string `json:"address"`
		Domains []string `json:"domains"`
	}

	type RoutingRule struct{
		Type string `json:"type"`
		Domain []string `json:"domain"`
		OutboundTag string `json:"outboundTag"`
	}

	type Outbound struct{
		Protocol string `json:"protocol"`
		Tag string `json:"tag"`
	}

	s1 := DnsServer{ Address: "8.8.8.8", Domains: []string{"geosite:geolocation-!cn"}}
	s2 := DnsServer{ Address: "1.1.1.1", Domains: []string{"geosite:geolocation-!cn"}}
	s3 := DnsServer{ Address: "223.5.5.5", Domains: []string{"geosite:cn"}}

	r1 := RoutingRule{ Type: "field", Domain: []string{"geosite:category-ads-all"}, OutboundTag: "block"}
	r2 := RoutingRule{ Type: "field", Domain: []string{"geosite:cn", "geosite:bing", "geosite:category-media-cn", "geosite:apple"}, OutboundTag: "direct"}
	r3 := RoutingRule{ Type: "field", Domain: []string{"geosite:google", "geosite:github", "geosite:telegram", "geosite:gfw", "geosite:geolocation-!cn"}, OutboundTag: "proxy"}

	o1 := Outbound{ Protocol: "freedom", Tag: "direct"}
	o2 := Outbound{ Protocol: "blackhole", Tag: "block"}

	(*con).Dns.Servers = []interface{}{ s1, s2, s3, "localhost"}
	(*con).Routing.Rules = []struct{Type string `json:"type"`; Domain []string `json:"domain"`; OutboundTag string `json:"outboundTag"`}{r1, r2, r3}
	(*con).Outbounds = append((*con).Outbounds, o1, o2)
}
