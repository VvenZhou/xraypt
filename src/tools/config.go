package tools

type Config struct{
	Log struct {
		Access string `json:"access"`
		Error string `json:"error"`
		Loglevel string `json:"loglevel"`
	} `json:"log"`
	Inbounds []interface{} `json:"inbounds"`
	Outbounds []interface{} `json:"outbounds"`
	Dns struct {
		Servers []interface{} `json:"servers"`
	} `json:"dns"`
	Routing struct {
		DomainStrategy string `json:"domainStrategy"`
		Rules []struct {
			Type string `json:"type"`
			Domain []string `json:"domain"`
			OutboundTag string `json:"outboundTag"`
		} `json:"rules"`
	} `json:"routing"`
}

type Outbound struct {
	Protocol string `json:"protocol"`
	Settings interface{} `json:"settings"`
	Tag string `json:"tag"`
	StreamSettings struct {
		Network string `json:"network,omitempty"`
		Security string `json:"security,omitempty"`
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

	} `json:"streamSettings,omitempty"`
	Mx struct {
		Enabled bool `json:"enabled,omitempty"`
	} `json:"mux,omitempty"`
}

type SocksIn struct{
	Tag string `json:"tag"`
	Listen string `json:"listen"`
	Protocol string `json:"protocol"`
	Port int `json:"port"`
	Settings struct {
		Auth string `json:"auth"`
		Ip string `json:"ip"`
		Udp bool `json:"udp"`
	}`json:"settings"`
}

type HttpIn struct{
	Tag string `json:"tag"`
	Listen string `json:"listen"`
	Protocol string `json:"protocol"`
	Port int `json:"port"`
}


func GenFinalConfig(con *Config) {
	type Outbound struct{
		Protocol string `json:"protocol"`
		Tag string `json:"tag"`
	}

	type DnsServer struct {
		Address string `json:"address"`
		Domains []string `json:"domains"`
	}

	type RoutingRule struct{
		Type string `json:"type"`
		Domain []string `json:"domain"`
		OutboundTag string `json:"outboundTag"`
	}
	htIn := HttpIn{Tag: "http-in", Listen: "::", Port: -1, Protocol: "http"}
	soIn := SocksIn{Tag: "socks-in", Port: 1080, Listen: "::", Protocol: "socks", Settings: struct{Auth string `json:"auth"`; Ip string `json:"ip"`; Udp bool `json:"udp"`}{Auth: "noauth", Udp: true, Ip: "127.0.0.1"}}
	(*con).Inbounds = []interface{}{ htIn, soIn }

	s1 := DnsServer{ Address: "8.8.8.8", Domains: []string{"geosite:geolocation-!cn"}}
	s2 := DnsServer{ Address: "1.1.1.1", Domains: []string{"geosite:geolocation-!cn"}}
	s3 := DnsServer{ Address: "223.5.5.5", Domains: []string{"geosite:cn"}}

	r1 := RoutingRule{ Type: "field", Domain: []string{"geosite:category-ads-all"}, OutboundTag: "block"}
	r2 := RoutingRule{ Type: "field", Domain: []string{"domain:music.126.net", "domain:nature.com", "geosite:cn", "geosite:bing", "geosite:category-media-cn", "geosite:apple"}, OutboundTag: "direct"}
	r3 := RoutingRule{ Type: "field", Domain: []string{"geosite:google", "geosite:github", "domain:youneed.win", "domain:freefq.com", "geosite:telegram", "geosite:gfw", "geosite:geolocation-!cn"}, OutboundTag: "proxy"}

	o1 := Outbound{ Protocol: "freedom", Tag: "direct"}
	o2 := Outbound{ Protocol: "blackhole", Tag: "block"}

	(*con).Dns.Servers = []interface{}{ s1, s2, s3, "localhost"}
	(*con).Routing.Rules = []struct{Type string `json:"type"`; Domain []string `json:"domain"`; OutboundTag string `json:"outboundTag"`}{r1, r2, r3}

	preCon := (*con).Outbounds[0]
	(*con).Outbounds = []interface{}{ preCon, o1, o2}
}

func OutboundToTestConfig(con *Config, out Outbound ) {
	htIn := HttpIn{Tag: "http-in", Listen: "::", Port: -1, Protocol: "http"}
//	soIn := SocksIn{Tag: "socks-in", Port: 0, Listen: "::", Protocol: "socks", Settings: struct{Auth string `json:"auth"`; Ip string `json:"ip"`; Udp bool `json:"udp"`}{Auth: "noauth", Udp: true, Ip: "127.0.0.1"}}
	//var con Config
	(*con).Log.Loglevel = "none"
	con.Log.Access = "none"
	con.Log.Error = "none"
//	(*con).Inbounds = []interface{}{ htIn, soIn }
	(*con).Inbounds = []interface{}{ htIn }
	(*con).Outbounds = []interface{}{ out }
	(*con).Dns.Servers = []interface{}{ "8.8.8.8", "1.1.1.1" }

	(*con).Routing.DomainStrategy = "AsIs"
	(*con).Routing.Rules = []struct{Type string `json:"type"`; Domain []string `json:"domain"`; OutboundTag string `json:"outboundTag"`}{}
}
