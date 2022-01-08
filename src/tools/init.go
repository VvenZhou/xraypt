package tools

import (
	"runtime"
	"time"
	"os"
)

var PreProxyPort int
var MainPort int
var RoutinePeriod int		// seconds

var XrayPath string
var SubsFilePath string
var PingOutPath string
var SpeedOutPath string

var GoodOutPath string
var BadOutPath string
var ErrorOutPath string

var LogPath string

var TempPath string
var JsonsPath string
var HalfJsonsPath string
var BackupPath string
var ConfigPath string
var OutPath string

var Mode int

var FlagVm, FlagVl, FlagSs, FlagSsr, FlagTrojan bool

var PThreadNum = 80
const SThreadNum = 10
const DSLine = 2.0

const PCnt = 3
const PingAllowFail = 2

const MaxTimeoutCnt = 25

const PRealCnt = 5
const PRealLeastNeeded = 3	// Must be >= 3 due to the Avg algorithm(src/ping/xrayping.go - getAvg())

const NodeTimeoutTolerance = 3

const subT = 5000
const pT = 2000 //ms
const pRealT = 1500 //ms
const sT = 15000 //ms


var RoutinePeriodDu = time.Duration(RoutinePeriod) * time.Second 

const PTimeout = time.Duration(pT * 2) * time.Millisecond
const PRealTimeout = time.Duration(pRealT * 2) * time.Millisecond
const STimeout = time.Duration(sT) * time.Millisecond
const SubTimeout = time.Duration(subT) * time.Millisecond

func PreCheck(protocols []string) {
	FlagVm, FlagVl, FlagSs, FlagSsr, FlagTrojan = checkProtocols(protocols)
	gVarInit()
	dirInit()
}

func isLinux() bool{
	os := runtime.GOOS
	if os == "linux" {
		return true
	}else{
		return false
	}
}

func gVarInit(){
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

		GoodOutPath = "out/goodOut.txt"
		BadOutPath = "out/badOut.txt"
		ErrorOutPath = "out/errorOut.txt"

		LogPath = "out/log.txt"

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
		ErrorOutPath = "out\\errorOut.txt"

		SubsFilePath = "config\\subs.txt"
	}
}

func dirInit(){
	if _, err := os.Stat(TempPath); os.IsNotExist(err) {
		os.MkdirAll(TempPath, 0755)
	}
	if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
		os.MkdirAll(ConfigPath, 0755)
	}
	if _, err := os.Stat(OutPath); os.IsNotExist(err) {
		os.MkdirAll(OutPath, 0755)
	}
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
