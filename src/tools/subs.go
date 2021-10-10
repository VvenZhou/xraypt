package tools
import (
	"net/url"
	"net/http"
	"io/ioutil"
	"io"
	"encoding/json"
	"encoding/base64"
	"log"
	"strings"
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

	var fail int = 0
	START_FREEFQ:
	pStr, err := getAllFromFreefq()
	if err != nil {
		log.Println("[ERROR]: Get from Freefq:", err)
		fail += 1
		if fail < 3 {
			goto START_FREEFQ
		}
	}else{
		if flagVl {
			Re := regexp.MustCompile(`vless://(.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.vlesses = append(subLs.vlesses, list[1])
			}
		}
		if flagVm {
			Re := regexp.MustCompile(`vmess://(.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.vms = append(subLs.vms, list[1])
			}
		}
		if flagSs{
			Re := regexp.MustCompile(`ss://(.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.sses = append(subLs.sses, list[1])
			}
		}
		if flagTrojan {
			Re := regexp.MustCompile(`trojan://(.*)<br>`)
			strStr := Re.FindAllStringSubmatch(*pStr, -1)
			for _, list := range strStr{
				subLs.trojans = append(subLs.trojans, list[1])
			}
		}
	}
	for _, sub := range subs {
		fail = 0
		START_SUBLINK:
		pdata, err := getStrFromSublink(sub)
		if err != nil {
			log.Println("[ERROR]", "SubGet:", sub, err)
			fail += 1
			if fail < 3 {
				goto START_SUBLINK
			}
		}else{
			log.Println("SubString got!")
			strs := strings.Fields(*pdata)
			for _, s := range strs {
				if flagVm {
					if l := strings.Split(s, "vmess://"); len(l)== 2 {
						subLs.vms = append(subLs.vms, l[1])
					}
				}
				if flagSs {
					if l := strings.Split(s, "ss://"); len(l)== 2 {
						subLs.sses = append(subLs.sses, l[1])
					}
				}
				if flagTrojan {
					if l := strings.Split(s, "trojan://"); len(l)== 2 {
						subLs.trojans = append(subLs.trojans, l[1])
					}
				}
				if flagVl {
					if l := strings.Split(s, "vless://"); len(l)== 2 {
						subLs.vlesses = append(subLs.vlesses, l[1])
					}
				}
				if flagSsr {
					if l := strings.Split(s, "ssr://"); len(l)== 2 {
						subLs.ssrs = append(subLs.ssrs, l[1])
					}
				}
			}
		}
	}
	if flagVm {
		fail = 0

		// Vms from YouNeedWind
		START_YOU:
		yousVms, err := getVmFromYou()
		if err != nil {
			log.Println("[ERROR]: getVmFrom You:", err)
			fail += 1
			if fail < 3 {
				goto START_YOU
			}
		}else{
			for _, vm := range yousVms {
				l := strings.Split(strings.TrimSpace(vm), "vmess://")
				subLs.vms = append(subLs.vms, l[1])
			}
		}

		// Vms from Vmout
		vmOutVms, err := getVmFromVmout()
		if err != nil {
			log.Println("[ERROR]: getVmFrom Vmout:", err)
		}else{
			for _, vm := range vmOutVms {
				l := strings.Split(strings.TrimSpace(vm), "vmess://")
				subLs.vms = append(subLs.vms, l[1])
			}
		}
	}

	log.Println("start remove duplicates...")
	log.Println("befor:", len(subLs.vms))
	subLs.vms = VmRemoveDulpicate(subLs.vms)
	log.Println("after:", len(subLs.vms))
	log.Println("remove duplicates done...")

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

func getVmFromYou() ([]string, error) {
	log.Println("You fetching start...")
	var cookie []*http.Cookie
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	req := HttpNewRequest("https://www.youneed.win/free-v2ray")

	rHtml, err := myClient.Do(req)
	if err != nil {
		//log.Println(err)
		return []string{}, err
	}
	defer rHtml.Body.Close()

	coo := rHtml.Cookies()
	if len(coo) > 0 {
		cookie = coo
	}

	body, err := io.ReadAll(rHtml.Body)
	if err != nil {
		//log.Println(err)
		return []string{}, err
	}
	ps_ajax := regexp.MustCompile(`var ps_ajax = \{.*,"nonce":"(.*?)".*,"post_id":"(\d+?)".*\};`)
	psStr := ps_ajax.FindStringSubmatch(string(body))
	if len(psStr) != 0 {
		log.Println("getVmFromYou nonce get.")
		//log.Printf("nonce: %s post_id: %s\n", psStr[1], psStr[2])
	}else{
		log.Printf("getVmFromYou error: no nonce information!")
		return []string{}, err
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
	req, err = http.NewRequest(http.MethodPost, "https://www.youneed.win/wp-admin/admin-ajax.php", strings.NewReader(data.Encode()))
	if err != nil {
		//log.Println(err)
		return []string{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Close = true
	for i := range cookie {
		req.AddCookie(cookie[i])
	}
	respContent, err := myClient.Do(req)
	if err != nil {
		//log.Println(err)
		return []string{}, err
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
	return vmes, nil
}

func getAllFromFreefq() (*string, error) {
	log.Println("Freefq fetching start...")
	var cookie []*http.Cookie
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	subLink := "https://www.freefq.com/v2ray/"
	req := HttpNewRequest(subLink)

	resp, err := myClient.Do(req)
	if err != nil {
		//log.Println("fetch error!")
		//return []string{}
		return nil, err
	}
	defer resp.Body.Close()

	coo := resp.Cookies()
	if len(coo) > 0 {
		cookie = coo
	}
	contents, _ := ioutil.ReadAll(resp.Body)

	doc, err := htmlquery.Parse(strings.NewReader(string(contents)))
	if err != nil {
		return nil, err
	}
	a := htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/ul[1]/li[1]/a")
	h2Tail := htmlquery.SelectAttr(a, "href")
	log.Printf("%s\n", h2Tail)

	s := []string{"https://www.freefq.com", h2Tail}
	h2 := strings.Join(s, "")
	log.Printf("%s\n", h2)

	req = HttpNewRequest(h2, cookie)
	resp2, err := myClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()
	contents, _ = ioutil.ReadAll(resp2.Body)
	doc, err = htmlquery.Parse(strings.NewReader(string(contents)))
	if err != nil {
		return nil, err
	}
	a = htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/table[2]/tbody/tr/td/div/fieldset/table/tbody/tr/td/a")
	h3 := htmlquery.SelectAttr(a, "href")
	log.Printf("%s\n", h3)

	req = HttpNewRequest(h3, cookie)
	resp3, err := myClient.Do(req)
	if err != nil {
		//log.Println("fetch h2 error!")
		//return []string{}
		return nil, err
	}
	defer resp3.Body.Close()
	contents, _ = ioutil.ReadAll(resp3.Body)
	strContents := string(contents)
	log.Println("Freefq fetching Done.")
	return &strContents, nil
}

func getStrFromSublink(subLink string) (*string, error) {
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	//myClient := HttpClientGet(0, SubTimeout)
	req := HttpNewRequest(subLink)

	resp, err := myClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	byteData, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		return nil, err
	}
	strData := string(byteData)
	return &strData, nil
}

func getVmFromVmout() ([]string, error){
	content, err := ioutil.ReadFile("vmOut.txt")
	if err != nil {
		return []string{}, err
	}
	var vms []string
	strs := strings.Split(string(content), "\n")
	for _, s := range strs {
		if len(strings.Split(strings.TrimSpace(s), "vmess://")) == 2 {
			vms = append(vms, strings.TrimSpace(s))
		}
	}

	content, err = ioutil.ReadFile("vmHalfOut.txt")
	if err != nil {
		return []string{}, err
	}
	strs = strings.Split(string(content), "\n")
	for _, s := range strs {
		if len(strings.Split(strings.TrimSpace(s), "vmess://")) == 2 {
			vms = append(vms, strings.TrimSpace(s))
		}
	}
	return vms, nil
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

func strInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
