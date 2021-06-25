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
var XrayPath string
var SubsFilePath string


const PThreadNum = 200
const SThreadNum = 4
const DSLine = 5.0
const PCnt = 7
const PRealCnt = 3

const pT = 1500 //ms
const pRealT = 1500 //ms
const sT = 20000 //ms

const PTimeout = time.Duration(pT*2) * time.Millisecond
const PRealTimeout = time.Duration(pRealT*2) * time.Millisecond
const STimeout = time.Duration(sT) * time.Millisecond

func Init(preProxyPort int) {
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
		XrayPath = "tools/xray"
	}else{
		JitPath = "jit\\"
		TempPath = "temp\\"
		BackupPath = "backup\\"
		JsonsPath = "jsons\\"
		XrayPath = "tools\\xray.exe"
	}
}

func DirInit(){
	if _, err := os.Stat(JitPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(JitPath, 0755)
	}
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
}

func HttpClientGet(port int, timeout time.Duration) *http.Client {
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: timeout}
	return myClient
}
