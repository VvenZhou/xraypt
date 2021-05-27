package tools
import (
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"encoding/base64"
	"log"
	"strings"

)

func SubGetVms(subs []string) []string {
	var vms []string
	for _, sub := range subs {
		data := SubGetStr(sub)
		strs := strings.Split(data, "\n")
		for _, s := range strs {
			if len(strings.Split(s, "vmess://")) == 2 {
				vms = append(vms, s)
			}
		}
	}
	log.Println("length of vms", len(vms))
	vmsNoDu := RemoveDuplicateStr(vms)
	log.Println("length of vmsNoDu", len(vms))
	return vmsNoDu
}

func SubGetStr(subLink string) string {
	port := 8123
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	resp, _ := myClient.Get(subLink)
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	log.Println("SubString got!")
	//return string(contents)

	data, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		log.Println("error:", err)
	}
	return string(data)
}
