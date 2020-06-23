//go:generate go run -tags generate gen.go

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"

	"github.com/zserge/lorca"
)

type counter struct {
	sync.Mutex
	lastNum float64
	lastOpt string
	count   float64
}

func (c *counter) Init() {
	c.lastNum = 0
	c.lastOpt = ""
	c.count = 0
}

func (c *counter) Calc(num string, opt string) string {
	c.Lock()
	defer c.Unlock()
	newNum, _ := strconv.ParseFloat(num, 32)

	switch c.lastOpt {
	case "":
		c.lastNum = newNum
	case "+":
		c.lastNum = c.lastNum + newNum
	case "-":
		c.lastNum = c.lastNum - newNum
	case "*":
		if num == "" {
			newNum = 1
		}
		c.lastNum = c.lastNum * newNum
	case "/":
		if num == "" {
			newNum = 1
		}
		c.lastNum = c.lastNum / newNum
	}

	c.count = c.lastNum

	if opt == "=" {
		c.lastNum = 0
		c.lastOpt = ""
	} else {
		c.lastOpt = opt
	}

	return strconv.FormatFloat(c.count, 'g', 5, 32)
}

func (c *counter) rstCalc() {
	c.Lock()
	defer c.Unlock()
	c.Init()
}

func main() {
	// log Setup
	f, err := os.OpenFile("logfile.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("file open error : %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetFlags(0)

	// Create UI with basic HTML passed via data URI
	ui, err := lorca.New("", "", 480, 320)
	if err != nil {
		log.Fatal(err)
	}
	defer ui.Close()

	ui.Bind("start", func() {
		log.Println("UI is ready")
	})

	c := &counter{}
	c.Init()
	ui.Bind("calculate", c.Calc)
	ui.Bind("reset", c.rstCalc)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))
	ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	log.Println("Server is ready")

	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt)
	select {
	case <-sigc:
	case <-ui.Done():
	}

	log.Println("exiting...")
}
