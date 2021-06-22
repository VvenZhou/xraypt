package tools

import (
	"runtime"
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
