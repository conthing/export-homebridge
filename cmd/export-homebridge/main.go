////todo 把现在sysmgmt的homebridge重启和清空数据文件两个功能搬过来
////todo 单元测试
//export-homebridge微服务功能介绍:1、向edgex注册export-homebridge这个微服务；2、从ha-project微服务获取灯光、窗帘等虚拟设备并将
// 虚拟设备传给homebridge；3、对事件event的处理，event包括:status、command等，status的处理是用zmq的pub和sub，command的处理是
// 用edgex的core-command这个微服务，最终homebridge上的虚拟设备会根据event做出相应的变化
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	httpsender "github.com/conthing/export-homebridge/pkg/http"
	"github.com/conthing/export-homebridge/pkg/logger"
	"github.com/conthing/export-homebridge/pkg/router"
	zmqinit "github.com/conthing/export-homebridge/pkg/zmqinit"

	"github.com/conthing/utils/common"
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
	Port       int    `json:"port"`
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

var cfg = Config{}

func boot(_ interface{}) (needRetry bool, err error) {
	err = httpsender.HttpPost(cfg.HTTP.Statusport)
	if err != nil {
		return true, err
	}
	err = zmqinit.InitZmq(cfg.HTTP.Statusport)
	if err != nil {
		return true, err
	}
	return false, nil
}
func main() {
	start := time.Now()
	var cfgfile string
	flag.StringVar(&cfgfile, "config", "configuration.toml", "Specify a profile other than default.") //如定义字符串就按定义的字符串来否则默认使用configuration.toml
	flag.StringVar(&cfgfile, "c", "configuration.toml", "Specify a profile other than default.")      //两种方式 同上
	flag.Parse()
	logger.InitLogger()
	err := common.LoadConfig(cfgfile, &cfg)
	if err != nil {
		logger.ERROR("failed to load config %v", err)
		return
	}
	if common.Bootstrap(boot, nil, 60000, 2000) != nil {
		return
	}
	errs := make(chan error, 3)
	listenForInterrupt(errs)
	startHTTPServer(errs, cfg.HTTP.Port)
	startZMQReceive(errs)
	// Time it took to start service
	logger.INFO("HTTP server listening on port %d, started in: %s", cfg.HTTP.Port, time.Since(start).String())
	// recv error channel
	c := <-errs
	logger.INFO("erminating: %v", c)
	os.Exit(0)
}

//有errs会使export-homebridge进程崩掉
func startHTTPServer(errChan chan error, port int) {
	go func() {
		r := router.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r))
	}()
}

//监听中断，如遇到ctrl+c等的可使export-homebridge直接结束掉，前提是export-homebridge得前台单独启动
func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

//添加原因是防止zmqinit.ZmqInit()函数因为数据错误造成export-homebridge启不来而健康检查又检查不到
func startZMQReceive(errChan chan error) {
	go func() {
		errChan <- zmqinit.ZmqInit()
	}()
}
