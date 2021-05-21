package ping

import (
	"fmt"
	"log"
	"os/exec"
//	"io"
//	"bufio"
	"net/http"
	"net"
	"net/url"
	"strings"
	"strconv"
	"time"
	"errors"

	"github.com/VvenZhou/xraypt/src/jsonEdit"
)

func Run(jsonPath string) (*Cmd, io.ReadCloser) {
	cmd := exec.Command("../tools/xray", "-c", jsonPath)
	stdout, _ := cmd.StdoutPipe()

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("xray executing, using json: %s\n", jsonPath)
	return cmd, stdout
}

//func print(stdout io.ReadCloser) {
//	for ;; {
//		r := bufio.NewReader(stdout)
//		line, _, err := r.ReadLine()
//		fmt.Printf("%s\n%s\n", string(line), err)
//	}
//}

func Ping(port int, interval int) (int, error){
	var timeout time.Duration = time.Duration(interval) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: timeout}

	start := time.Now()
	resp, err := myClient.Get("http://www.google.com/gen_204")
	elapsed := time.Since(start)

	if err != nil {
		return 0, err
	}

	code := resp.StatusCode
	defer resp.Body.Close()
	delay := elapsed.Milliseconds()
	if code != 204 {
		return 0, errors.New("Ping err: StatusCode is not 204")
	}
	return int(delay),nil
}

func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func XrayPing(jsonRead string) (int, error) {
	const timeout = 1000 //ms
	const count = 5 //count of ping
	var totalDelay, avgDelay int
	var fail int = 0

	jsonWrite := []string{"/tmp/tmp_", jsonPath}
	strings.Join(jsonWrite, "")
	err := jsonEdit.JsonChangePort(jsonRead, jsonWrite, port)
	if err != nil {
		log.Fatal(err)
	}

	cmd, _ := Run(jsonWrite)

	port, _ := GetFreePort()
	//fmt.Println(port)

	for i := 0; i < count; i++ {
		delay, err := Ping(port, timeout)
		if err != nil {
			fail += 1
			fmt.Print(err)
		}else{
			totalDelay += delay
			fmt.Println(delay)
		}
	}

	if err = cmd.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}

	if fail == 5 {
		//fmt.Println("None")
		return 0, errors.New("Ping not accessable")
	}else{
		avgDelay = int(totalDelay/(count-fail))
		fmt.Printf("avgDelay: %d\n", avgDelay)
		return avgDelay, nil
	}
}
