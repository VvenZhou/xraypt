package tools

import (
	"runtime"
	"time"
	"strconv"
	"net/url"
	"net/http"
	"strings"
	"log"
	"os"
)

var JitPath string
var TempPath string
var BackupPath string
var PreProxyPort int
var JsonsPath string
var HalfJsonsPath string
var XrayPath string
var SubsFilePath string


const PThreadNum = 250
const SThreadNum = 10
const DSLine = 2.0

const PCnt = 5
const PingAllowFail = 3

const PRealCnt = 5
const PRealAllowFail = 2


const subT = 5000
const pT = 2000 //ms
const pRealT = 1000 //ms
const sT = 20000 //ms

const PTimeout = time.Duration(pT*2) * time.Millisecond
const PRealTimeout = time.Duration(pRealT*2) * time.Millisecond
const STimeout = time.Duration(sT) * time.Millisecond
const SubTimeout = time.Duration(subT) * time.Millisecond

func PreCheck(preProxyPort int) {
	GVarInit(preProxyPort)
	DirInit()
}

func isLinux() bool{
	os := runtime.GOOS
	log.Println("Platform:", os)
	SubsFilePath = "subs.txt"
	if os == "linux" {
		return true
	}else{
		return false
	}
}

func GVarInit(port int){
	PreProxyPort = port
	if isLinux() {
		JitPath = "jit/"
		TempPath = "temp/"
		BackupPath = "backup/"
		JsonsPath = "jsons/"
		HalfJsonsPath = "halfJsons/"
		XrayPath = "tools/xray"
	}else{
		JitPath = "jit\\"
		TempPath = "temp\\"
		BackupPath = "backup\\"
		JsonsPath = "jsons\\"
		HalfJsonsPath = "halfJsons\\"
		XrayPath = "tools\\xray.exe"
	}
}

func DirInit(){
	//if _, err := os.Stat(JitPath); os.IsNotExist(err) {
	//	// path/to/whatever does not exist
	//	os.MkdirAll(JitPath, 0755)
	//}
	if _, err := os.Stat(TempPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(TempPath, 0755)
	}
	if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(BackupPath, 0755)
	}
	if _, err := os.Stat(JsonsPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(JsonsPath, 0755)
	}
	if _, err := os.Stat(HalfJsonsPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(HalfJsonsPath, 0755)
	}
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
