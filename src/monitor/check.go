package monitor

import(
	"log"
	"net/http"
	"net/url"
	"strings"
	"strconv"
	"github.com/VvenZhou/xraypt/src/tools"
	"github.com/VvenZhou/xraypt/src/ping"
)

func main() {
	pingJob := make(chan *tools.Node, len(vmLinks))
	go ping.XrayPing(&wgPing, pingJob, pingResult, pCount, pTimeout, pRealCount, pRealTimeout)
}

func checkConnection(jobs <-chan *tools.Node, result chan<- *tools.Node){
	testJob := make(chan *tools.Node, len(vmLinks))
	for  := range jobs {
		go ping.XrayPing(&wgPing, pingJob, pingResult)
	}
}
