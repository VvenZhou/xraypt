package tools

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"time"
	"net/http"
	"net/url"
	"strings"
	"strconv"
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

func HttpClientGet(port int, timeout time.Duration) http.Client {
	if port == 0 {
		myClient := http.Client{}
		return myClient
	}else{
		str := []string{"http://127.0.0.1", strconv.Itoa(port)}
		proxyUrl, _ := url.Parse(strings.Join(str, ":"))
		myClient := http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: timeout}
		return myClient
	}
}

func HttpNewRequest(url string, cookies ...[]*http.Cookie) *http.Request {
	var coo []*http.Cookie

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:89.0) Gecko/20100101 Firefox/89.0i")
	req.Close = true
	if len(cookies) > 0 {
		coo = cookies[0]
		for i := range coo {
			req.AddCookie(coo[i])
		}
	}
	return req
}
