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
	"github.com/antchfx/htmlquery"
)

func SubGetVms(subs []string) []string {
	var vms []string

	vmOutVms := subGetVmFromVmOut()
	for _, vm := range vmOutVms {
		vms = append(vms, vm)
	}

	yousVms := subGetYousVms()
	for _, vm := range yousVms {
		vms = append(vms, strings.TrimSpace(vm))
	}

	freefqVms := subGetFreefqVms()
	for _, vm := range freefqVms {
		vms = append(vms, strings.TrimSpace(vm))
	}

	for _, sub := range subs {
		data, err := subLinkGetStr(sub)
		if err != nil {
			log.Println("[ERROR]", "SubGet:", err)
			continue
		}
		strs := strings.Fields(data)
		for _, s := range strs {
			if len(strings.Split(s, "vmess://")) == 2 {
				vms = append(vms, s)
			}
		}
	}

	log.Println("length of vms", len(vms))
	vmsNoDu := RemoveDuplicateStr(vms)
	log.Println("length of vmsNoDu", len(vmsNoDu))
	return vmsNoDu
}

func subGetYousVms() []string {
	var t time.Duration = time.Duration(10000) * time.Millisecond
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
	log.Println("YouNeedWind get", len(vmes), "vmesses.")
	return vmes
}

func subGetFreefqVms() []string{
	var t time.Duration = time.Duration(10000) * time.Millisecond
	subLink := "https://www.freefq.com/v2ray/"
	port := 8123
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}
	resp, err := myClient.Get(subLink)
	if err != nil {
		log.Println("fetch error!")
		return []string{}
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	//doc, err := htmlquery.LoadURL("http://example.com/")
	doc, err := htmlquery.Parse(strings.NewReader(string(contents)))
	a := htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/ul[1]/li[1]/a")
	h2Tail := htmlquery.SelectAttr(a, "href")
	log.Printf("%s\n", h2Tail)

	s := []string{"https://www.freefq.com", h2Tail}
	h2 := strings.Join(s, "")
	log.Printf("%s\n", h2)
	resp2, err := myClient.Get(h2)
	if err != nil {
		log.Println("fetch h2 error!")
		return []string{}
	}
	defer resp2.Body.Close()
	contents, _ = ioutil.ReadAll(resp2.Body)
	doc, err = htmlquery.Parse(strings.NewReader(string(contents)))
	a = htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/table[2]/tbody/tr/td/div/fieldset/table/tbody/tr/td/a")
	h3 := htmlquery.SelectAttr(a, "href")
	log.Printf("%s\n", h3)

	resp3, err := myClient.Get(h3)
	if err != nil {
		log.Println("fetch h2 error!")
		return []string{}
	}
	defer resp3.Body.Close()
	contents, _ = ioutil.ReadAll(resp3.Body)
	//fmt.Printf("%s\n", string(contents))

	var vms []string
	vmRe := regexp.MustCompile(`(vmess://.*)<br>`)
	strStr := vmRe.FindAllStringSubmatch(string(contents), -1)
	for _, list := range strStr{
		vms = append(vms, list[1])
	}
	log.Println("Freefq get", len(vms), "vmesses.")
	return vms
}

func subLinkGetStr(subLink string) (string, error) {
	var t time.Duration = time.Duration(10000) * time.Millisecond
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

	data, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		log.Println("error:", err)
	}
	return string(data), nil
}

func subGetVmFromVmOut () []string{
	content, err := ioutil.ReadFile("vmOut.txt")
	if err != nil {
		log.Fatal(err)
	}
	var vms []string
	strs := strings.Split(string(content), "\n")
	for _, s := range strs {
		if len(strings.Split(strings.TrimSpace(s), "vmess://")) == 2 {
			vms = append(vms, strings.TrimSpace(s))
		}
	}
	log.Println("Get VmOut", len(vms), "vmesses.")
	return vms
}
