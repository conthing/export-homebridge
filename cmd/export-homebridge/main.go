package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpsender "github.com/conthing/export-homebridge/pkg/http"
	zmqinit "github.com/conthing/export-homebridge/pkg/zmqinit"
	"github.com/gorilla/context"
	serial "github.com/tarm/goserial"
)

//Config is the data from config
type Config struct {
	Serial   serial.Config
	HTTP     HTTPConfig
	Commands []CommandConfig
}

//HTTPConfig is the port from config
type HTTPConfig struct {
	Port int
}

//CommandConfig is the data from config
type CommandConfig struct {
	Name string
	Data []int
}

//CommandData is the data from config
type CommandData struct {
	Data []byte
}

//Status is the data which is ready to send to js
type Status struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Service        string `json:"service"`
	Characteristic map[string]interface{}
}

func main() {
	start := time.Now()
	var profile string

	flag.StringVar(&profile, "profile", "config.json", "Specify a profile other than default.")
	flag.StringVar(&profile, "p", "config.json", "Specify a profile other than default.")
	flag.Parse()

	cfg := &Config{}

	httpsender.HttpPost()

	go zmqinit.ZmqInit()

	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(profile)
	if err != nil {
		log.Println(err)
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, cfg)
	if err != nil {
		log.Println(err)
		return
	}

	errs := make(chan error, 3)
	listenForInterrupt(errs)

	startHTTPServer(errs, cfg.HTTP.Port)

	// Time it took to start service
	log.Printf("HTTP server listening on port %d, started in: %s", cfg.HTTP.Port, time.Since(start).String())

	// recv error channel
	c := <-errs
	log.Println(fmt.Sprintf("terminating: %v", c))
	os.Exit(0)

}

func startHTTPServer(errChan chan error, port int) {
	go func() {
		r := zmqinit.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r))
	}()
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
