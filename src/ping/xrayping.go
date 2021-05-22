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


func Ping(port int, timeout int) (int, error){
	var t time.Duration = time.Duration(timeout) * time.Millisecond
	str := []string{"http://127.0.0.1", strconv.Itoa(port)}
	proxyUrl, _ := url.Parse(strings.Join(str, ":"))
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}, Timeout: t}

	start := time.Now()
	resp, err := myClient.Get("http://www.google.com/gen_204")
	stop := time.Now()
	if err != nil {
		return 0, err
	}
	code := resp.StatusCode

	defer resp.Body.Close()
	if code != 204 {
		return 0, errors.New("Ping err: StatusCode is not 204")
	}

	elapsed := stop.Sub(start)
	delay := elapsed.Milliseconds()
	return int(delay), nil
}


func XrayPing(jsonPath string, count int, timeout int) (int, error) {
	var totalDelay int = 0
	var avgDelay int = 0
	var fail int = 0
	var max int = 0

	var x tools.Xray
	err := x.Init(jsonPath)
	if err != nil {
		log.Fatal(err)
	}
	err = x.Run()
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < count; i++ {
		delay, err := Ping(x.Port, timeout)
		if err != nil {
			fail += 1
			fmt.Println(err)
		}else{
			if max < delay {
				max = delay
			}
			fmt.Println(delay)
			totalDelay += delay
		}
	}

	err = x.Stop()
	if err != nil {
		log.Fatal(err)
	}

	if fail == 5 {
		//fmt.Println("None")
		return 0, errors.New("Ping not accessable")
	}else{
		avgDelay = (totalDelay-max)/(count-fail-1)
		//fmt.Printf("avgDelay: %d\n", avgDelay)
		return avgDelay, nil
	}
}
