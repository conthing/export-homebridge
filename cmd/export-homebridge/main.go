package main

import (
	"flag"
	"fmt"
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
	"github.com/conthing/utils/common"
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
	Port int `json:"port"`
	Statusport string `json:"statusport"`
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

var cfg  = Config{}

func boot(_ interface{}) (needRetry bool, err error) {

	err = httpsender.HttpPost(cfg.HTTP.Statusport)
	if err!=nil{
		return true,err
	}

	err = zmqinit.InitZmq(cfg.HTTP.Statusport)
	if err!=nil{
		return true,err
	}

	go zmqinit.ZmqInit()


	return false, nil
}


func main() {
	start := time.Now()

	var cfgfile string
	flag.StringVar(&cfgfile, "config", "configuration.toml", "Specify a profile other than default.")
	flag.StringVar(&cfgfile, "c", "configuration.toml", "Specify a profile other than default.")
	flag.Parse()

		common.InitLogger(&common.LoggerConfig{Level: "DEBUG", SkipCaller: true})

	err := common.LoadConfig(cfgfile, &cfg)
	if err != nil {
		common.Log.Errorf("failed to load config %v", err)
		return
	}

	if common.Bootstrap(boot, nil, 60000, 2000) != nil {
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
