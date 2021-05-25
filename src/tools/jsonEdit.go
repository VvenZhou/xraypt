package tools

import (
	"encoding/json"
	"io/ioutil"
)

func JsonChangePort(jsonRead, jsonWrite string, port int) error {

	type Setting struct{
		Auth string `json:"auth"`
		IP string `json:"ip"`
		Udp bool `json:"udp"`
	}
	type Logs struct{
		Access string `json:"access"`
		Error string `json:"error"`
		Loglevel string `json:"loglevel"`
	}

	type Inbound struct{
		Tag string `json:"tag"`
		Listen string `json:"listen"`
		Protocol string `json:"protocol"`
		Port int `json:"port"`
		Settings Setting `json:"settings"`
	}

	type Config struct{
		Log Logs `json:"log"`
		Inbounds []Inbound `json:"inbounds"`
		Outbounds []interface{} `json:"outbounds"`
		Dns map[string]interface{} `json:"dns"`
		routing map[string]interface{} `json:"routing"`
	}

	byteValue, err := ioutil.ReadFile(jsonRead)
	if err != nil {
		return err
	}

	var con Config

	err = json.Unmarshal(byteValue, &con)
	if err != nil {
		return err
	}

	for i, in := range con.Inbounds {
		if in.Protocol == "http" {
			con.Inbounds[i].Port = port
		}
		con.Inbounds[0].Port = port
	}

	byteValue, err = json.MarshalIndent(con, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonWrite, byteValue, 0644)
	if err != nil {
		return err
	}
	return nil
}
