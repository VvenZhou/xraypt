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

type Links struct {
	vms []string
	sses []string
	ssrs []string
	trojans []string
	vlesses []string
}

func SubGet(protocols []string, subs []string) []string {
	var subLs Links
	//var strsOut []string

	flagVm, flagVl, flagSs, flagSsr, flagTrojan := checkProtocols(protocols)

	pStr, err := getAllFromFreefq()
	if err != nil {
		log.Println("Get from Freefq error:", err)
	}else{
		if flagVl {
			Re := regexp.MustCompile(`(vless://.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.vlesses = append(subLs.vlesses, list[1])
			}
		}
		if flagVm {
			Re := regexp.MustCompile(`(vmess://.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.vms = append(subLs.vms, list[1])
			}
		}
		if flagSs{
			Re := regexp.MustCompile(`(ss://.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.sses = append(subLs.sses, list[1])
			}
		}
		if flagTrojan {
			Re := regexp.MustCompile(`(trojan://.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.trojans = append(subLs.trojans, list[1])
			}
		}
	}
	for _, sub := range subs {
		pdata, err := getStrFromSublink(sub)
		if err != nil {
			log.Println("[ERROR]", "SubGet:", err)
			continue
		}
		strs := strings.Fields(*pdata)
		for _, s := range strs {
			if flagVm {
				if len(strings.Split(s, "vmess://")) == 2 {
					subLs.vms = append(subLs.vms, s)
				}
			}
			if flagSs {
				if len(strings.Split(s, "ss://")) == 2 {
					subLs.sses = append(subLs.sses, s)
				}
			}
			if flagTrojan {
				if len(strings.Split(s, "trojan://")) == 2 {
					subLs.trojans = append(subLs.trojans, s)
				}
			}
			if flagVl {
				if len(strings.Split(s, "vless://")) == 2 {
					subLs.vlesses = append(subLs.vlesses, s)
				}
			}
			if flagSsr {
				if len(strings.Split(s, "ssr://")) == 2 {
					subLs.ssrs = append(subLs.ssrs, s)
				}
			}
		}
	}
	if flagVm {
		yousVms := getVmFromYou()
		for _, vm := range yousVms {
			subLs.vms = append(subLs.vms, strings.TrimSpace(vm))
		}
		vmOutVms := getVmFromVmout()
		for _, vm := range vmOutVms {
			subLs.vms = append(subLs.vms, vm)
		}
	}

	subLs.vms = VmRemoveDulpicate(subLs.vms)

	if flagVm {
		log.Println("get vms:", len(subLs.vms))
	}
	if flagVl {
		log.Println("get vlesses:", len(subLs.vlesses))
	}
	if flagSs {
		log.Println("get sses:", len(subLs.sses))
	}
	if flagSsr {
		log.Println("get ssrs:", len(subLs.ssrs))
	}
	if flagTrojan{
		log.Println("get trojans:", len(subLs.trojans))
	}

	return subLs.vms
}

//func getVms(subs []string) []string {
//	var vms []string
//
//	vmOutVms := GetVmFromVmOut()
//	for _, vm := range vmOutVms {
//		vms = append(vms, vm)
//	}
//
//	yousVms := GetVmFromYousVms()
//	for _, vm := range yousVms {
//		vms = append(vms, strings.TrimSpace(vm))
//	}
//
//	vm2 := VmRemoveDulpicate(vms)
//	return vm2
//}

func getVmFromYou() []string {
	var t time.Duration = time.Duration(10000) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(PreProxyPort)}
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
	if len(psStr) != 0 {
		log.Println("getVmFromYou nonce get.")
		//log.Printf("nonce: %s post_id: %s\n", psStr[1], psStr[2])
	}else{
		log.Printf("getVmFromYou error: no nonce information!")
		return []string{}
	}

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

func getAllFromFreefq() (*string, error) {
	var t time.Duration = time.Duration(10000) * time.Millisecond
	subLink := "https://www.freefq.com/v2ray/"
	str := []string{"http://127.0.0.1", strconv.Itoa(PreProxyPort)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}
	resp, err := myClient.Get(subLink)
	if err != nil {
		//log.Println("fetch error!")
		//return []string{}
		return nil, err
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
		//log.Println("fetch h2 error!")
		//return []string{}
		return nil, err
	}
	defer resp2.Body.Close()
	contents, _ = ioutil.ReadAll(resp2.Body)
	doc, err = htmlquery.Parse(strings.NewReader(string(contents)))
	a = htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/table[2]/tbody/tr/td/div/fieldset/table/tbody/tr/td/a")
	h3 := htmlquery.SelectAttr(a, "href")
	log.Printf("%s\n", h3)

	resp3, err := myClient.Get(h3)
	if err != nil {
		//log.Println("fetch h2 error!")
		//return []string{}
		return nil, err
	}
	defer resp3.Body.Close()
	contents, _ = ioutil.ReadAll(resp3.Body)
	strContents := string(contents)

	return &strContents, nil
	//fmt.Printf("%s\n", string(contents))

	//var vms []string
	//vmRe := regexp.MustCompile(`(vmess://.*)<br>`)
	//strStr := vmRe.FindAllStringSubmatch(string(contents), -1)
	//for _, list := range strStr{
	//	vms = append(vms, list[1])
	//}
	//log.Println("Freefq get", len(vms), "vmesses.")

	//return vms
}

func getStrFromSublink(subLink string) (*string, error) {
	var t time.Duration = time.Duration(10000) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(PreProxyPort)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}
	//myClient := &http.Client{Timeout: t}

	resp, err := myClient.Get(subLink)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)
	log.Println("SubString got!")

	byteData, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		log.Println("error:", err)
	}
	strData := string(byteData)
	return &strData, nil
}

func getVmFromVmout() []string{
	content, err := ioutil.ReadFile("vmOut.txt")
	if err != nil {
		return []string{}
	}
	var vms []string
	strs := strings.Split(string(content), "\n")
	for _, s := range strs {
		if len(strings.Split(strings.TrimSpace(s), "vmess://")) == 2 {
			vms = append(vms, strings.TrimSpace(s))
		}
	}
	//log.Println("Get VmOut", len(vms), "vmesses.")
	return vms
}

func strInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func checkProtocols(protocols []string) (flagVm , flagVl, flagSs, flagSsr, flagTrojan bool) {
	if strInSlice("vless", protocols){
		flagVl = true
	}else{
		flagVl = false
	}
	if strInSlice("vmess", protocols){
		flagVm = true
	}else{
		flagVm = false
	}
	if strInSlice("ss", protocols){
		flagSs = true
	}else{
		flagSs = false
	}
	if strInSlice("trojan", protocols){
		flagTrojan = true
	}else{
		flagTrojan = false
	}
	if strInSlice("ssr", protocols){
		flagSsr = true
	}else{
		flagSsr = false
	}
	return
}
