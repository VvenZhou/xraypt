package tools

import (
	"runtime"
	"log"
	"os"
)

var JitPath string
var TempPath string
var PreProxyPort int
var JsonsPath string
var XrayPath string

func Init(preProxyPort int) {
	GVarInit(preProxyPort)
	DirInit()
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

func GVarInit(port int){
	PreProxyPort = port
	if isLinux() {
		JitPath = "jit/"
		TempPath = "temp/"
		JsonsPath = "jsons/"
		XrayPath = "tools/xray"
	}else{
		JitPath = "jit\\"
		TempPath = "temp\\"
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
	if _, err := os.Stat(JsonsPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		os.MkdirAll(JsonsPath, 0755)
	}
}