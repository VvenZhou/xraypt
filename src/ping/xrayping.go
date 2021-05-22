package ping

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"time"
	"errors"

	"github.com/VvenZhou/xraypt/src/tools"
)


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


func XrayPing(jsonRead string) (int, error) {
	const timeout = 1000 //ms
	const count = 5 //count of ping
	var totalDelay, avgDelay int
	var fail int = 0

	port, _ := tools.GetFreePort()

	s := []string{"/tmp/tmp_", jsonRead}
	jsonWrite := strings.Join(s, "")
	err := tools.JsonChangePort(jsonRead, jsonWrite, port)
	if err != nil {
		log.Fatal(err)
	}

	cmd := tools.RunXray(jsonWrite)
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < count; i++ {
		delay, err := Ping(port, timeout)
		if err != nil {
			fail += 1
			fmt.Println(err)
		}else{
			fmt.Println(delay)
			totalDelay += delay
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
		//fmt.Printf("avgDelay: %d\n", avgDelay)
		return avgDelay, nil
	}
}
