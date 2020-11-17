////todo 把现在sysmgmt的homebridge重启和清空数据文件两个功能搬过来
////todo 单元测试
//export-homebridge微服务功能介绍:1、向edgex注册export-homebridge这个微服务；2、从ha-project微服务获取灯光、窗帘等虚拟设备并将
// 虚拟设备传给homebridge；3、对事件event的处理，event包括:status、command等，status的处理是用zmq的pub和sub，command的处理是
// 用edgex的core-command这个微服务，最终homebridge上的虚拟设备会根据event做出相应的变化
package main

import (
	"flag"
	"fmt"
	"github.com/conthing/export-homebridge/getedgexparams"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	zmqinit "github.com/conthing/export-homebridge/zmqreceivesendhandler"

	"github.com/conthing/utils/common"
	"github.com/gorilla/context"
	"github.com/tarm/goserial"
)

//Config is the data from config
type Config struct {
	Serial   serial.Config
	HTTP     HTTPConfig
	Commands []CommandConfig
	Log      common.LoggerConfig
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

//Status is the data which is ready to send to js
type Status struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Service        string `json:"service"`
	Characteristic map[string]interface{}
}

//初始化结构体Conifg，因不知道上方Config结构体里面的变量具体是什么，所以不用写。注:此句不可以改为cfg := Config{}因为若修改则需要放在函数内部，但由于多个函数在调用该变量所以不行
var cfg = Config{}

/*boot函数:1、因不知道edgex传入的params是什么，所以输入参数暂时不填写，输入参数的类型是interface{};2、判断export-homebridge有
没有注册成功，没有则返回err并继续尝试；3、判断zmq有没有初始化成功，没有则返回err并继续尝试，有则返回nil并不再尝试*/
func boot(_ interface{}) (needRetry bool, err error) {
	err = getedgexparams.HttpPost(cfg.HTTP.Statusport)
	if err != nil {
		common.Log.Errorf("boot(_ interface{}) getedgexparams.HttpPost(cfg.HTTP.Statusport) failed: %v", err)
		return true, err
	}
	err = zmqinit.InitZmq(cfg.HTTP.Statusport)
	if err != nil {
		common.Log.Errorf("boot(_ interface{}) zmqinit.InitZmq(cfg.HTTP.Statusport) failed: %v", err)
		return true, err
	}
	return false, nil
}

/* main函数:1、配置文件起名并加载配置文件里面的内容；2、初始化日志；3、定义可以使main.go崩溃或export-homebridge没法编译和正常
运行的3个错误的通道，分别是:listenForInterrupt(监听中断比如前台跑有没有按Ctrl+C)、startHTTPServer(监听HTTPServer有没有起来)、
startZMQReceive(ZMQ的PUB和SUB正常不正常即能不能正常的接收数据)*/
func main() {
	start := time.Now() //记录下当前本地的时间
	var cfgfile string
	flag.StringVar(&cfgfile, "config", "configuration.toml", "Specify a profile other than default.") //如定义字符串就按定义的字符串来否则默认使用configuration.toml
	flag.StringVar(&cfgfile, "c", "configuration.toml", "Specify a profile other than default.")      //两种方式 同上
	flag.Parse()
	err := common.LoadConfig(cfgfile, &cfg)
	common.InitLogger(&cfg.Log)
	if err != nil {
		common.Log.Errorf("main() common.LoadConfig(cfgfile, &cfg) failed %v", err)
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
	common.Log.Infof("HTTP server listening on port %d, started in: %s", cfg.HTTP.Port, time.Since(start).String())
	// recv error channel
	c := <-errs
	common.Log.Infof("erminating: %v", c)
	os.Exit(0)
}

//有errs会使export-homebridge进程崩掉
func startHTTPServer(errChan chan error, port int) {
	go func() {
		r := LoadRestRoutes()
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
