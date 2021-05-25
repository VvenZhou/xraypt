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
		Servers []string `json:"servers"`
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

//type Dns struct {
//	Servers []string `json:"servers"`
//}

//type Routing struct {
//	DomainStrategy string `json:"domainStrategy"`
//	Rules []struct {
//		Type string `json:"type"`
//		Domain []string `json:"domain"`
//		OutboundTag string `json:"outboundTag"`
//	} `json:"rules"`
//}
