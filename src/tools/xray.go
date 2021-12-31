package tools

import(
	"os/exec"
	"syscall"
	"time"
	"log"
)

type Xray struct {
	Port     int
	JsonPath string
	cmd      *exec.Cmd
}



func (x *Xray) Init(port int, jsonPath string) error {
	x.Port = port
	x.JsonPath = jsonPath
	return nil
}

func (x *Xray) Run() error {
	x.cmd = exec.Command(XrayPath, "-c", x.JsonPath)

	err := x.cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(750 * time.Millisecond)
	return nil
}

func (x *Xray) Stop() error {
	err := x.cmd.Process.Signal(syscall.SIGTERM)

	_, err = x.cmd.Process.Wait()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

