package tools

import (
	"encoding/json"
	"io/ioutil"
	"net"
)

func JsonChangePort(jsonRead, jsonWrite string, port int) error {

	byteValue, err := ioutil.ReadFile(jsonRead)
	if err != nil {
		return err
	}

	var con Config
	err = json.Unmarshal(byteValue, &con)
	if err != nil {
		return err
	}

	for _, in := range con.Inbounds {
		inMap := in.(map[string]interface{})
		if inMap["protocol"] == "http" {
			//con.Inbounds[i]["port"] = port
			inMap["port"] = port
		}
	}

	byteValue, err = json.MarshalIndent(con, "", "    ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(jsonWrite, byteValue, 0644)
	if err != nil {
		return err
	}

	//fmt.Printf("%v\n", con)
	return nil
}


func GetFreePorts(count int) ([]int, error) {
	var ports []int
	for i := 0; i < count; i++ {
		addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
		if err != nil {
			return nil, err
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			return nil, err
		}
		defer l.Close()
		ports = append(ports, l.Addr().(*net.TCPAddr).Port)
	}
	return ports, nil
}