package tools
import (
	"net/url"
	"sort"
	"net/http"
	"io/ioutil"
	"io"
	"errors"
	"os"
	"encoding/json"
	"encoding/base64"
	"log"
	"strings"
	"strconv"
	"regexp"
	"github.com/antchfx/htmlquery"
)


type NodeLists struct {
	Vms	[]*Node
	Sses	[]*Node
	Ssrs	[]*Node
	Trojans	[]*Node
	Vlesses	[]*Node
}

type Links struct {
	Vms []string
	Sses []string
	Ssrs []string
	Trojans []string
	Vlesses []string
}


func (l *Links) AddToNodeLists(nodeLs *NodeLists) {
	for _, str := range l.Vms {
		var n Node
		n.Init("vmess", str)
		nodeLs.Vms = append(nodeLs.Vms, &n)
	}
	for _, str := range l.Sses {
		var n Node
		n.Init("ss", str)
		nodeLs.Sses = append(nodeLs.Sses, &n)
	}
	for _, str := range l.Ssrs {
		var n Node
		n.Init("ssr", str)
		nodeLs.Ssrs = append(nodeLs.Ssrs, &n)
	}
	for _, str := range l.Trojans {
		var n Node
		n.Init("trojan", str)
		nodeLs.Trojans = append(nodeLs.Trojans, &n)
	}
	for _, str := range l.Vlesses {
		var n Node
		n.Init("vless", str)
		nodeLs.Vlesses = append(nodeLs.Vlesses, &n)
	}
}


func GetAllNodes(nodeLs *NodeLists) {
	// Get nodes from GoodOut.txt && 
	GetNodeLsFromFormatedFile(nodeLs, GoodOutPath)

	var nodeLs2 NodeLists
	GetNodeLsFromFormatedFile(&nodeLs2, BadOutPath)
	for _, node := range nodeLs2.Vms {
		if node.Timeout < MaxTimeoutCnt {
			nodeLs.Vms = append(nodeLs.Vms, node)
		}
	}
	sort.Stable(ByTimeout(nodeLs.Vms))
	for _, node := range nodeLs2.Sses {
		if node.Timeout < MaxTimeoutCnt {
			nodeLs.Sses = append(nodeLs.Sses, node)
		}
	}
	sort.Stable(ByTimeout(nodeLs.Sses))

	var subs []string
	var subLs Links
	byteData, err := ioutil.ReadFile(SubsFilePath)
	if err != nil {
		log.Println("SubFile read error:", err)
	}else{
		log.Println("SubFile get...")
		subs = strings.Fields(string(byteData))
	}
	getLinks(&subLs, subs)
	subLs.AddToNodeLists(nodeLs)

	//Remove duplicates
	log.Println("start remove duplicates...")
	if FlagVm && len(nodeLs.Vms)!=0 {
		log.Printf("vm befor: %d    ", len(nodeLs.Vms))
		VmRemoveDuplicateNodes(&(nodeLs.Vms))
		//nodeLs.Vms = VmRemoveDulpicate(nodeLs.Vms)
		log.Printf("after: %d\n", len(nodeLs.Vms))
	}
	if FlagSs && len(nodeLs.Sses)!=0 {
		log.Printf("ss befor: %d    ", len(nodeLs.Sses))
		SsRemoveDuplicateNodes(&(nodeLs.Sses))
		//nodeLs.Sses = SsRemoveDulpicate(nodeLs.Sses)
		log.Printf("after: %d\n", len(nodeLs.Sses))
	}
	log.Println("remove duplicates done...")


	//Show total counts
	if FlagVm {
		log.Println("get Vms:", len(nodeLs.Vms))
	}
	if FlagVl {
		log.Println("get vlesses:", len(nodeLs.Vlesses))
	}
	if FlagSs {
		log.Println("get sses:", len(nodeLs.Sses))
	}
	if FlagSsr {
		log.Println("get ssrs:", len(nodeLs.Ssrs))
	}
	if FlagTrojan{
		log.Println("get trojans:", len(nodeLs.Trojans))
	}
}

func getLinks(subLs *Links, subs []string) {
	var fail int
	var links []string

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

	// Get Vms from YouNeedWind
	if FlagVm {
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

	//Dispatch links
	DispatchLinks(subLs, links)


}

func GetNodeLsFromFile(nodeLs *NodeLists, filePath string) {
	var subLs Links
	links, _ := getLinksFromFile(filePath)
	DispatchLinks(&subLs, links)
	subLs.AddToNodeLists(nodeLs)
}

func GetNodeLsFromFormatedFile(nodeLs *NodeLists, filePath string ) {
	var nodes []*Node
	nodes, err := GetNodesFromFormatedFile(filePath)
	if err != nil {
		return
	}

	DispatchNodes(nodeLs, nodes)
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

func DispatchLinks(subLs *Links, links []string) {
	for _, str := range links {
		var s []string
		if FlagVm {
			if strings.HasPrefix(str, "vmess://") {
				s = strings.Split(str, "://")
				subLs.Vms = append(subLs.Vms, s[1])
			}
		}
		if FlagVl {
			if strings.HasPrefix(str, "vless://") {
				s = strings.Split(str, "://")
				subLs.Vlesses = append(subLs.Vlesses, s[1])
			}
		}
		if FlagSs {
			if strings.HasPrefix(str, "ss://") {
				s = strings.Split(str, "://")
				subLs.Sses = append(subLs.Sses, s[1])
			}
		}
		if FlagSsr {
			if strings.HasPrefix(str, "ssr://") {
				s = strings.Split(str, "://")
				subLs.Ssrs = append(subLs.Ssrs, s[1])
			}
		}
		if FlagTrojan{
			if strings.HasPrefix(str, "trojan://") {
				s = strings.Split(str, "://")
				subLs.Trojans = append(subLs.Trojans, s[1])
			}
		}
	}

}

func DispatchNodes(nodeLs *NodeLists, nodes []*Node) {
	for _, node := range nodes {
		if FlagVm {
			if node.Type == "vmess" {
				nodeLs.Vms = append(nodeLs.Vms, node)
			}
		}
		if FlagSs {
			if node.Type == "ss" {
				nodeLs.Sses = append(nodeLs.Sses, node)
			}
		}
		if FlagVl {
			if node.Type == "vless" {
				nodeLs.Vlesses = append(nodeLs.Vlesses, node)
			}
		}
		if FlagSsr {
			if node.Type == "ssr" {
				nodeLs.Ssrs = append(nodeLs.Ssrs, node)
			}
		}
		if FlagTrojan {
			if node.Type == "trojan" {
				nodeLs.Trojans = append(nodeLs.Trojans, node)
			}
		}
	}
}

func GetNodesFromFormatedFile(filePath string) ([]*Node, error) {
	var nodes []*Node
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	strs := strings.Fields(string(content))
	for _, str := range strs {
		var node Node
		s := strings.Split(str, ",")
		if len(s) != 4 {
			return nil, errors.New("string format error")
		}
		i, _ := strconv.Atoi(s[0])
		t, _ := strconv.Atoi(s[1])
		node.AvgDelay = i 
		node.Timeout = t
		//node.Country = s[2]
		node.Type = s[2]
		node.ShareLink = s[3]
		nodes = append(nodes, &node)
	}
	return nodes, nil
}

func WriteNodesToFormatedFile(filePath string, nodes []*Node) error {
	var rows []string

	if len(nodes) == 0  {
		return nil
	}

	for _, n := range nodes {
		str := []string{strconv.Itoa(n.AvgDelay), strconv.Itoa(n.Timeout), n.Type, n.ShareLink}
		row := strings.Join(str, ",")
		rows = append(rows, row)
	}

	bytes := []byte(strings.Join(rows[:], "\n"))
	err := os.WriteFile(filePath, bytes, 0644)
	if err != nil {
		log.Println(err)
		return err
	}else{
		log.Println(filePath, "generated!")
		return nil
	}
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
	var links []string

	log.Println("Freefq fetching start...")
	var cookie []*http.Cookie
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	subLinks := []string{"https://www.freefq.com/v2ray/", "https://www.freefq.com/free-ss/"}
	for _, subLink := range(subLinks) {
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
		for _, str := range strings.Fields(strContents) {
			s := extractAvailableLink(str)
			if s != "" {
				links = append(links, s)
			}
		}
	}

	log.Println("Freefq get", len(links), "links.")
	return links, nil
}
