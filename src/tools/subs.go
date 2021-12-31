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
	Vms []string
	Sses []string
	Ssrs []string
	Trojans []string
	Vlesses []string
}

var protocols = []string{
	"vmess",
	"vless",
	"ss",
	"ssr",
	"trojan"}

func GetSubLinks(subLs *Links) {
	var subs []string
	byteData, err := ioutil.ReadFile(SubsFilePath)
	if err != nil {
		log.Println("SubFile read error:", err)
	}else{
		log.Println("SubFile get...")
		subs = strings.Fields(string(byteData))
	}

	subGet(subLs, protocols, subs)
}

func subGet(subLs *Links, protocols []string, subs []string) {
	var fail int
	var links []string

	flagVm, flagVl, flagSs, flagSsr, flagTrojan := checkProtocols(protocols)

	// Get links from Freefq
	START_FREEFQ:
	strLinks, err := getAllFromFreefq()
	if err != nil {
		log.Println("[ERROR]: Get from Freefq:", err)
		fail += 1
		if fail < 3 {
			goto START_FREEFQ
		}
	}else{
		if len(strLinks) != 0 {
			links = append(links, strLinks...)
		}
	}

	//for _, s := range links {
	//	fmt.Println(s, "\n")
	//}
	//os.Exit(0)


	//Sublink
	for _, sub := range subs {
		fail = 0
		START_SUBLINK:
		strLinks, err := getStrFromSublink(sub)
		if err != nil {
			log.Println("[ERROR]", "SubGet:", sub, err)
			fail += 1
			if fail < 3 {
				goto START_SUBLINK
			}
		}else{
			if len(strLinks) != 0 {
				links = append(links, strLinks...)
				log.Println("SubString got!")
			}
		}
	}

	// Links from speedOut
	strLinks, err = getLinksFromFile("speedOut.txt")
	if err != nil {
		log.Println("[ERROR]: getVmFrom speedOut.txt:", err)
	}else{
		if len(strLinks) != 0 {
			links = append(links, strLinks...)
			log.Println("speedOut got!")
		}
	}
	// Links from pingOut
	strLinks, err = getLinksFromFile("pingOut.txt")
	if err != nil {
		log.Println("[ERROR]: getVmFrom pingOut.txt:", err)
	}else{
		if len(strLinks) != 0 {
			links = append(links, strLinks...)
			log.Println("speedOut got!")
		}
	}

	// Vms from YouNeedWind
	if flagVm {
		fail = 0

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
				l := strings.Split(vm, "vmess://")
				subLs.Vms = append(subLs.Vms, l[1])
			}
		}

	}

	//Dispatch links
	for _, str := range links {
		var s []string
		if flagVm {
			if strings.HasPrefix(str, "vmess://") {
				s = strings.Split(str, "://")
				subLs.Vms = append(subLs.Vms, s[1])
			}
		}
		if flagVl {
			if strings.HasPrefix(str, "vless://") {
				s = strings.Split(str, "://")
				subLs.Vlesses = append(subLs.Vlesses, s[1])
			}
		}
		if flagSs {
			if strings.HasPrefix(str, "ss://") {
				s = strings.Split(str, "://")
				subLs.Sses = append(subLs.Sses, s[1])
			}
		}
		if flagSsr {
			if strings.HasPrefix(str, "ssr://") {
				s = strings.Split(str, "://")
				subLs.Ssrs = append(subLs.Ssrs, s[1])
			}
		}
		if flagTrojan{
			if strings.HasPrefix(str, "trojan://") {
				s = strings.Split(str, "://")
				subLs.Trojans = append(subLs.Trojans, s[1])
			}
		}
	}


	//Remove duplicates
	log.Println("start remove duplicates...")
	if flagVm && len(subLs.Vms)!=0 {
		log.Printf("vm befor: %d    ", len(subLs.Vms))
		subLs.Vms = VmRemoveDulpicate(subLs.Vms)
		log.Printf("after: %d\n", len(subLs.Vms))
	}
	if flagSs && len(subLs.Sses)!=0 {
		log.Printf("ss befor: %d    ", len(subLs.Sses))
		subLs.Sses = SsRemoveDulpicate(subLs.Sses)
		log.Printf("after: %d\n", len(subLs.Sses))
	}
	log.Println("remove duplicates done...")



	//Show total counts
	if flagVm {
		log.Println("get Vms:", len(subLs.Vms))
	}
	if flagVl {
		log.Println("get vlesses:", len(subLs.Vlesses))
	}
	if flagSs {
		log.Println("get sses:", len(subLs.Sses))
	}
	if flagSsr {
		log.Println("get ssrs:", len(subLs.Ssrs))
	}
	if flagTrojan{
		log.Println("get trojans:", len(subLs.Trojans))
	}

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

func getAllFromFreefq() ([]string, error) {
	//Get content from website
	log.Println("Freefq fetching start...")
	var cookie []*http.Cookie
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	subLink := "https://www.freefq.com/v2ray/"
	req := HttpNewRequest(subLink)

	resp, err := myClient.Do(req)
	if err != nil {
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
		return nil, err
	}
	defer resp3.Body.Close()
	contents, _ = ioutil.ReadAll(resp3.Body)
	strContents := string(contents)
	log.Println("Freefq fetching Done.")

	//Extract links from content
	var links []string
	for _, str := range strings.Fields(strContents) {
		s := extractAvailableLink(str)
		if s != "" {
			links = append(links, s)
		}
	}

	return links, nil
}

func getStrFromSublink(subLink string) ([]string, error) {
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


	var links []string
	strs := strings.Fields(strData)
	for _, str := range strs {
		s := extractAvailableLink(str)
		if s != "" {
			links = append(links, s)
		}
	}

	return links, nil
}

func getLinksFromFile(fileName string) ([]string, error){
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return []string{}, err
	}
	var links []string
	strs := strings.Fields(string(content))
	for _, str := range strs {
		s := extractAvailableLink(str)
		if s != "" {
			links = append(links, s)
		}
	}
	return links, nil
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

func extractAvailableLink(str string) string {
	Re := regexp.MustCompile(`[!a-z]*?vless://([^<]*)\s*<*`)
	strList := Re.FindAllStringSubmatch(str, -1)
	if len(strList) != 0 {
		return (strList[0][0])
	}
	Re = regexp.MustCompile(`[!a-z]*?vmess://([^<]*)\s*`)
	strList = Re.FindAllStringSubmatch(str, -1)
	if len(strList) != 0 {
		return (strList[0][0])
	}
	Re = regexp.MustCompile(`[!a-z]*?ss://([^<]*)\s*`)
	strList = Re.FindAllStringSubmatch(str, -1)
	if len(strList) != 0 {
		return (strList[0][0])
	}
	Re = regexp.MustCompile(`[!a-z]*?trojan://([^<]*)\s*`)
	strList = Re.FindAllStringSubmatch(str, -1)
	if len(strList) != 0 {
		return (strList[0][0])
	}
	Re = regexp.MustCompile(`[!a-z]*?ssr://([^<]*)\s*`)
	strList = Re.FindAllStringSubmatch(str, -1)
	if len(strList) != 0 {
		return (strList[0][0])
	}

	return ""
}

func strInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
