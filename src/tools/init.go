package tools

import (
	"runtime"
	"time"
	"log"
	"os"
)

var PreProxyPort int
const MainPort = 8123

var XrayPath string
var SubsFilePath string
var PingOutPath string
var SpeedOutPath string

var TempPath string
var JsonsPath string
var HalfJsonsPath string
var BackupPath string
var ConfigPath string
var OutPath string


var PThreadNum = 150
const SThreadNum = 10
const DSLine = 2.0

const PCnt = 5
const PingAllowFail = 3

const PRealCnt = 8
const PRealLeastNeeded = 5	// Must be >= 3 due to the Avg algorithm(src/ping/xrayping.go - getAvg())


const subT = 5000
const pT = 2000 //ms
const pRealT = 1000 //ms
const sT = 15000 //ms

const PTimeout = time.Duration(pT * 2) * time.Millisecond
const PRealTimeout = time.Duration(pRealT * 2) * time.Millisecond
const STimeout = time.Duration(sT) * time.Millisecond
const SubTimeout = time.Duration(subT) * time.Millisecond

func PreCheck(preProxyPort int) {
	gVarInit(preProxyPort)
	dirInit()
}

func isLinux() bool{
	os := runtime.GOOS
	log.Println("Platform:", os)
	if os == "linux" {
		return true
	}else{
		return false
	}
}

func gVarInit(port int){
	PreProxyPort = port
	if isLinux() {
		XrayPath = "tools/xray"

		TempPath = "temp/"
		BackupPath = "backup/"
		ConfigPath = "config/"
		OutPath = "out/"

		JsonsPath = "out/jsons/"
		HalfJsonsPath = "out/halfJsons/"

		PingOutPath = "out/pingOut.txt"
		SpeedOutPath = "out/speedOut.txt"
		SubsFilePath = "config/subs.txt"
	}else{
		XrayPath = "tools\\xray.exe"

		TempPath = "temp\\"
		BackupPath = "backup\\"
		ConfigPath = "config\\"
		OutPath = "out\\"

		JsonsPath = "out\\jsons\\"
		HalfJsonsPath = "out\\halfJsons\\"

		PingOutPath = "out\\pingOut.txt"
		SpeedOutPath = "out\\speedOut.txt"
		SubsFilePath = "config\\subs.txt"
	}
}

func dirInit(){
	if _, err := os.Stat(TempPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(TempPath, 0755)
	}
	if _, err := os.Stat(BackupPath); os.IsNotExist(err) {
		os.MkdirAll(BackupPath, 0755)
	}
	if _, err := os.Stat(JsonsPath); os.IsNotExist(err) {
		os.MkdirAll(JsonsPath, 0755)
	}
	if _, err := os.Stat(HalfJsonsPath); os.IsNotExist(err) {
		os.MkdirAll(HalfJsonsPath, 0755)
	}
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		os.MkdirAll(ConfigPath, 0755)
	}
	if _, err := os.Stat(OutPath); os.IsNotExist(err) {
		os.MkdirAll(OutPath, 0755)
	}
}

