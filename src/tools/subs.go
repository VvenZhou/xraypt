package tools
import (
	"net/url"
	"sort"
	"net/http"
	"io/ioutil"
	"io"
	"errors"
	"os"
	"fmt"
	"encoding/json"
	"encoding/base64"
	"log"
	"strings"
	"strconv"
	"regexp"
	"context"
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


func GetAllNodes(ctx context.Context, nodeLs *NodeLists) error {
	var errs []error
	// Get nodes from GoodOut.txt
	err := GetNodeLsFromFormatedFile(nodeLs, GoodOutPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("Get NodeLs from FFile(GoodOut):%w", err))
	}

	var nodeLs2 NodeLists
	err = GetNodeLsFromFormatedFile(&nodeLs2, BadOutPath)
	if err != nil {
		errs = append(errs, fmt.Errorf("Get NodeLs from FFile(BadOut):%w", err))
	}else{
		for _, node := range nodeLs2.Vms {
			if node.Timeout < MaxTimeoutCnt {
				nodeLs.Vms = append(nodeLs.Vms, node)
			}
		}
		for _, node := range nodeLs2.Sses {
			if node.Timeout < MaxTimeoutCnt {
				nodeLs.Sses = append(nodeLs.Sses, node)
			}
		}
	}

	sort.Stable(ByTimeout(nodeLs.Vms))
	sort.Stable(ByTimeout(nodeLs.Sses))

	//Get nodes from Web
	var subs []string
	var subLs Links
	byteData, err := ioutil.ReadFile(SubsFilePath)
	if err != nil {
//		return fmt.Errorf("ReadFile(Subfile):%w", err)
		errs = append(errs, fmt.Errorf("ReadFile(SubFile):%w", err))
	}else{
		log.Println("SubFile get...")
		subs = strings.Fields(string(byteData))
		err = getLinks(ctx, &subLs, subs)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			errs = append(errs, fmt.Errorf("GetLinks:%w", err))
		}
		subLs.AddToNodeLists(nodeLs)
	}

	//Remove duplicates
	log.Println("Removing duplicates...")
	if FlagVm && len(nodeLs.Vms)!=0 {
		l := len(nodeLs.Vms)
		VmRemoveDuplicateNodes(&(nodeLs.Vms))
		log.Printf("vm %d -> %d", l,  len(nodeLs.Vms))
	}
	if FlagSs && len(nodeLs.Sses)!=0 {
		l := len(nodeLs.Sses)
		SsRemoveDuplicateNodes(&(nodeLs.Sses))
		log.Printf("ss %d -> %d\n", l, len(nodeLs.Sses))
	}
	log.Println("...Remove duplicates done")


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


	if len(errs) == 0 {
		return nil
	}else{
		var err error
		for _, e := range(errs) {
			err = fmt.Errorf("%w|%w", err, e)
		}
		err = fmt.Errorf("(%w)", err)
		return err
	}
}

func getLinks(ctx context.Context, subLs *Links, subs []string) error {
	var fail int
	var links []string

	// Get links from Freefq
	fail = 0
	for {
		strLinks, err := getAllFromFreefq(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Println("[Error] Freefq:", err)
			fail += 1
			if fail > 2 {
				break
			}
		}else{
			log.Println("Freefq get", len(strLinks), "links.")
			if len(strLinks) != 0 {
				links = append(links, strLinks...)
			}
			break
		}
	}

	// Get Vms from YouNeedWind
	if FlagVm {
		fail = 0
		for {
			youVms, err := getVmFromYou(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				log.Println("[Error] You:", err)
				fail += 1
				if fail > 2 {
					break
				}
			}else{
				log.Println("You get", len(youVms), "links.")
				for _, vm := range youVms {
					l := strings.Split(vm, "vmess://")
					subLs.Vms = append(subLs.Vms, l[1])
				}
				break
			}
		}
	}

	//Sublink
	for _, sub := range subs {
		fail = 0
		for {
			strLinks, err := getStrFromSublink(ctx, sub)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return err
				}
				log.Println("[Error] SubGet:", sub, err)
				fail += 1
				if fail > 2 {
					break
				}
			}else{
				log.Println("Subs get", len(strLinks), "links.")
				if len(strLinks) != 0 {
					links = append(links, strLinks...)
				}
				break
			}
		}
	}

	//Dispatch links
	DispatchLinks(subLs, links)

	return nil
}

func GetNodeLsFromFile(nodeLs *NodeLists, filePath string) {
	var subLs Links
	links, _ := getLinksFromFile(filePath)
	DispatchLinks(&subLs, links)
	subLs.AddToNodeLists(nodeLs)
}

func GetNodeLsFromFormatedFile(nodeLs *NodeLists, filePath string ) error {
	var nodes []*Node
	nodes, err := GetNodesFromFormatedFile(filePath)
	if err != nil {
		return fmt.Errorf("GetNodesFromFFile:%w", err)
	}

	DispatchNodes(nodeLs, nodes)

	return nil
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
		return nil, fmt.Errorf("ReadFile:%w", err)
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
		return fmt.Errorf("WriteFile:%w", err)
	}else{
//		log.Println(filePath, "generated!")
		return nil
	}
}

func getStrFromSublink(ctx context.Context, subLink string) ([]string, error) {
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	req := HttpNewRequest("GET", subLink)
	req = req.WithContext(ctx)

	resp, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do:%w", err)
	}
	defer resp.Body.Close()
	contents, _ := ioutil.ReadAll(resp.Body)

	byteData, err := base64.StdEncoding.DecodeString(string(contents))
	if err != nil {
		return nil, fmt.Errorf("Base64 decode:%w", err)
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

func getVmFromYou(ctx context.Context) ([]string, error) {
	log.Println("You start...")
//	var cookie []*http.Cookie
	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	req := HttpNewRequest("GET", "https://www.youneed.win/free-v2ray")
	req = req.WithContext(ctx)

	rHtml, err := myClient.Do(req)
	if err != nil {
		//log.Println(err)
		return nil, fmt.Errorf("Do(0):%w", err)
	}
//	defer rHtml.Body.Close()

	body, err := io.ReadAll(rHtml.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll:%w", err)
	}

	rHtml.Body.Close()

	ps_ajax := regexp.MustCompile(`var ps_ajax = \{.*,"nonce":"(.*?)".*,"post_id":"(\d+?)".*\};`)
	psStr := ps_ajax.FindStringSubmatch(string(body))
	if len(psStr) == 0 {
		return nil, errors.New("No nonce info")
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
		return nil, fmt.Errorf("http.NewRequest:%w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(ctx)

	respContent, err := myClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Do(1):%w", err)
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
//	log.Println("YouNeedWind get", len(vmes), "vmesses.")
	log.Println("...You finished")
	return vmes, nil
}

func getAllFromFreefq(ctx context.Context) ([]string, error) {
	log.Println("Freefq start...")
	//Get content from website
	var links []string

	myClient := HttpClientGet(PreProxyPort, SubTimeout)
	subLinks := []string{"https://www.freefq.com/v2ray/", 
			"https://www.freefq.com/free-ss/",
			"https://www.freefq.com/free-xray/"}
	for _, subLink := range(subLinks) {
		req := HttpNewRequest("GET", subLink)
		req = req.WithContext(ctx)

		resp, err := myClient.Do(req)
		if err != nil {
			return nil, err
		}
//		defer resp.Body.Close()

		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		resp.Body.Close()

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

		req = HttpNewRequest("GET", h2)
		req = req.WithContext(ctx)
		resp2, err := myClient.Do(req)
		if err != nil {
			return nil, err
		}
//		defer resp2.Body.Close()
		contents, _ = ioutil.ReadAll(resp2.Body)

		resp2.Body.Close()

		doc, err = htmlquery.Parse(strings.NewReader(string(contents)))
		if err != nil {
			return nil, err
		}
		a = htmlquery.FindOne(doc, "/html/body/table[4]/tbody/tr/td[1]/table[2]/tbody/tr/td/table[2]/tbody/tr/td/div/fieldset/table/tbody/tr/td/a")
		h3 := htmlquery.SelectAttr(a, "href")
		log.Printf("%s\n", h3)

		req = HttpNewRequest("GET", h3)
		req = req.WithContext(ctx)
		resp3, err := myClient.Do(req)
		if err != nil {
			return nil, err
		}
//		defer resp3.Body.Close()
		contents, _ = ioutil.ReadAll(resp3.Body)

		resp3.Body.Close()

		strContents := string(contents)

		//Extract links from content
		for _, str := range strings.Fields(strContents) {
			s := extractAvailableLink(str)
			if s != "" {
				links = append(links, s)
			}
		}
	}

	log.Println("...Freefq finished")
	return links, nil
}
