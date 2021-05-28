package tools
import (
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"io"
	"encoding/json"
	"encoding/base64"
	"log"
	"strings"
	"time"
	"regexp"
)

func SubGetVms(subs []string) []string {
	var vms []string
	yousVms := subGetYousVms()
	for _, vm := range yousVms {
		vms = append(vms, vm)
	}
	for _, sub := range subs {
		data, err := subGetStr(sub)
		if err != nil {
			log.Println("[ERROR]", "SubGet:", err)
			continue
		}
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

func subGetYousVms() []string {
	var t time.Duration = time.Duration(8000) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(8123)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

	rHtml, err := myClient.Get("https://www.youneed.win/free-v2ray")
	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer rHtml.Body.Close()

	body, err := io.ReadAll(rHtml.Body)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	ps_ajax := regexp.MustCompile(`var ps_ajax = \{.*,"nonce":"(.*?)".*,"post_id":"(\d+?)".*\};`)
	psStr := ps_ajax.FindStringSubmatch(string(body))
	log.Printf("nonce: %s post_id: %s\n", psStr[1], psStr[2])

	nonceStr := psStr[1]
	postId := psStr[2]
	data := url.Values{
		"action": {"validate_input"},
		"nonce": {nonceStr},
		"captcha": {"success"},
		"post_id": {postId},
		"type": {"captcha"},
		"protection": {""},
	}
	//fmt.Println(data.Encode())
	req, err := http.NewRequest(http.MethodPost, "https://www.youneed.win/wp-admin/admin-ajax.php", strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return []string{}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	respContent, err := myClient.Do(req)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer respContent.Body.Close()

	var vmes []string
	var res map[string]interface{}
	json.NewDecoder(respContent.Body).Decode(&res)
	vmRe := regexp.MustCompile(`<td align="center" class="v2ray"><a herf="#" data-raw="(.*)" style.*`)
	content, _ := res["content"].(string)
	vmStrStr := vmRe.FindAllStringSubmatch(content, -1)
	for _, vm := range vmStrStr {
		vmes = append(vmes, vm[1])
	}
	log.Println("YouNeedWind vmesses get!")
	return vmes
}

func subGetStr(subLink string) (string, error) {
	var t time.Duration = time.Duration(5000) * time.Millisecond
	port := 8123
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

	resp, err := myClient.Get(subLink)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	log.Println("SubString got!")
	//return string(contents)

	data, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		log.Println("error:", err)
	}
	return string(data), nil
}

