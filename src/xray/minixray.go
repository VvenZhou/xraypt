package xray

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/VvenZhou/xraypt/src/tools"

	applog "github.com/xtls/xray-core/app/log"
	commlog "github.com/xtls/xray-core/common/log"
	"github.com/xtls/xray-core/app/dispatcher"
	"github.com/xtls/xray-core/app/proxyman"
	v2net "github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/serial"
	core "github.com/xtls/xray-core/core"
	"github.com/xtls/xray-core/infra/conf"
	cfgcommon "github.com/xtls/xray-core/infra/conf"
)

//type VmessLink struct {
//	Ver      string      `json:"v"`
//	Add      string      `json:"add"`
////	Aid      interface{} `json:"aid"`
//	Host     string      `json:"host"`
//	ID       string      `json:"id"`
//	Net      string      `json:"net"`
//	Path     string      `json:"path"`
//	Port     interface{} `json:"port"`
//	Ps       string      `json:"ps"`
//	TLS      string      `json:"tls"`
//	Type     string      `json:"type"`
//	OrigLink string      `json:"-"`
//}

func StartXray(lType string, shareLink string, useMux, allowInsecure bool) (*core.Instance, error) {
	loglevel := commlog.Severity_Error
	var ob *core.OutboundHandlerConfig
	if lType == "vmess" {
		var lk *VmessLink

		lk, err := ParseVmess("vmess://"+shareLink)
		if err != nil {
			var vmShare tools.VmessShare
			lk = new(VmessLink)
			err := tools.VmlinkToVmshare(&vmShare, shareLink)
			if err != nil {
				return nil, err
			}

			lk.Ver = "2"
			lk.Add = vmShare.Add
			lk.Host = vmShare.Host
			lk.ID = vmShare.Id
			lk.Net = vmShare.Net
			lk.Path = vmShare.Path
			lk.Port = vmShare.Port
			lk.Ps = "none"
			lk.TLS = vmShare.Tls
			lk.Type = "none"
			lk.OrigLink = "vmess://" + shareLink
		}

		ob, err = Vmess2Outbound(lk, useMux, allowInsecure)
		if err != nil {
			return nil, err
		}
	}else if lType == "ss" {
		var ssShare tools.SsShare
		err := tools.SsLinkToShare(&ssShare, shareLink)
		if err != nil {
			return nil, err
		}
		ob, err = ss2Outbound(&ssShare, useMux, allowInsecure)
		if err != nil {
			return nil, err
		}
	}


	config := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&applog.Config{
				ErrorLogType:  applog.LogType_File,
				ErrorLogPath:  os.DevNull,
				ErrorLogLevel: loglevel,
			}),
			serial.ToTypedMessage(&dispatcher.Config{}),
//			serial.ToTypedMessage(&proxyman.InboundConfig{}),
			serial.ToTypedMessage(&proxyman.OutboundConfig{}),
		},
	}

//	commlog.RegisterHandler(commlog.NewLogger(fileWriterCreater))
	config.Outbound = []*core.OutboundHandlerConfig{ob}
	server, err := core.New(config)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func MeasureDelay(inst *core.Instance, timeout time.Duration, dest string) (int, error) {
	c, err := CoreHTTPClient(inst, timeout)
	if err != nil {
		return 0, err
	}

	req, _ := http.NewRequest("HEAD", dest, nil)
//	req.Close = true

	start := time.Now()
	resp, err := c.Do(req)
	stop := time.Now()
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code > 399 {
		return -1, fmt.Errorf("status incorrect (>= 400): %d", code)
	}
	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()

	return int(delay), nil
}

func Vmess2Outbound(v *VmessLink, useMux, allowInsecure bool) (*core.OutboundHandlerConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = "proxy"
	out.Protocol = "vmess"
	out.MuxSettings = &conf.MuxConfig{}
	if useMux {
		out.MuxSettings.Enabled = true
		out.MuxSettings.Concurrency = 8
	}

	p := conf.TransportProtocol(v.Net)
	s := &conf.StreamConfig{
		Network:  &p,
		Security: v.TLS,
	}

	switch v.Net {
	case "tcp":
		s.TCPSettings = &conf.TCPConfig{}
		if v.Type == "" || v.Type == "none" {
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(`{ "type": "none" }`))
		} else {
			pathb, _ := json.Marshal(strings.Split(v.Path, ","))
			hostb, _ := json.Marshal(strings.Split(v.Host, ","))
			s.TCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`
			{
				"type": "http",
				"request": {
					"path": %s,
					"headers": {
						"Host": %s
					}
				}
			}
			`, string(pathb), string(hostb))))
		}
	case "kcp":
		s.KCPSettings = &conf.KCPConfig{}
		s.KCPSettings.HeaderConfig = json.RawMessage([]byte(fmt.Sprintf(`{ "type": "%s" }`, v.Type)))
	case "ws":
		s.WSSettings = &conf.WebSocketConfig{}
		s.WSSettings.Path = v.Path
		s.WSSettings.Headers = map[string]string{
			"Host": v.Host,
		}
	case "h2", "http":
		s.HTTPSettings = &conf.HTTPConfig{
			Path: v.Path,
		}
		if v.Host != "" {
			h := cfgcommon.StringList(strings.Split(v.Host, ","))
			s.HTTPSettings.Host = &h
		}
	}

	if v.TLS == "tls" {
		s.TLSSettings = &conf.TLSConfig{
			Insecure: allowInsecure,
		}
		if v.Host != "" {
			s.TLSSettings.ServerName = v.Host
		}
	}

	out.StreamSetting = s
	oset := json.RawMessage([]byte(fmt.Sprintf(`{
  "vnext": [
    {
      "address": "%s",
      "port": %v,
      "users": [
        {
          "id": "%s",
          "security": "auto"
        }
      ]
    }
  ]
}`, v.Add, v.Port, v.ID)))
	out.Settings = &oset
	return out.Build()
}

func ss2Outbound(s *tools.SsShare, useMux, allowInsecure bool) (*core.OutboundHandlerConfig, error) {
	out := &conf.OutboundDetourConfig{}
	out.Tag = "proxy"
	out.Protocol = "shadowsocks"
	oset := json.RawMessage([]byte(fmt.Sprintf(`{
  "servers": [
    {
      "address": "%s",
      "port": %d,
      "method": "%s",
      "password": "%s"
    }
  ]
}`, s.Addr, s.Port, s.Method, s.Pwd)))
	out.Settings = &oset
	return out.Build()
}


func CoreHTTPClient(inst *core.Instance, timeout time.Duration) (*http.Client, error) {
	if inst == nil {
		return nil, errors.New("core instance nil")
	}

	tr := &http.Transport{
		DisableKeepAlives: true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dest, err := v2net.ParseDestination(fmt.Sprintf("%s:%s", network, addr))
			if err != nil {
				return nil, err
			}
			return core.Dial(ctx, inst, dest)
		},
	}

	c := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	return c, nil
}

func CoreVersion() string {
	return core.Version()
}

