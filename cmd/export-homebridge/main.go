////todo 对这个微服务的功能介绍
////todo 把现在sysmgmt的homebridge重启和清空数据文件两个功能搬过来
////todo 单元测试
//export-homebridge微服务介绍:1、向edgex注册export-homebridge这个微服务；2、从ha-project微服务获取灯光、窗帘等虚拟设备并将
// 虚拟设备传给homebridge；3、对事件event的处理，event包括:status、command等，status的处理是用zmq的pub和sub，command的处理是
// 用edgex的core-command这个微服务，最终homebridge上的虚拟设备会根据event做出相应的变化

package main //约定:1、除了每个微服务的main.go外，package名可以不和距离main.go最近的文件夹名一样，其余的.go文件都要和距离它最
// 近的文件夹名保持一致，main.go的package名就为main；2、package和import之间要空一行；3、package名要小写；

import ( //导入的包包括3部分:1、goland自带的包；2、项目文件夹中的包；3、第三方包，例如:github上导入的包；约定:goland自
	// 带的包放在最上面，空一行放项目文件夹中的包，再空一行放第三方包
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

func boot(_ interface{}) (needRetry bool, err error) { //启动引导函数
	err = httpsender.HttpPost(cfg.HTTP.Statusport)
	if err != nil {
		return true, err
	}
	err = zmqinit.InitZmq(cfg.HTTP.Statusport)
	if err != nil {
		return true, err
	}
	//go zmqinit.ZmqInit()   //最下方有zmqinit.ZmqInit()函数的调用，这里是多余的go线程注释掉
	return false, nil
}
func main() {
	start := time.Now()
	var cfgfile string
	flag.StringVar(&cfgfile, "config", "configuration.toml", "Specify a profile other than default.") //如定义字符串就按定义的字符串来否则默认使用configuration.toml
	flag.StringVar(&cfgfile, "c", "configuration.toml", "Specify a profile other than default.")      //两种方式 同上
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
	errs := make(chan error, 3)          //定义一个数量为3个的切片通道
	listenForInterrupt(errs)             //调用监听函数
	startHTTPServer(errs, cfg.HTTP.Port) //调用HTTPServer函数
	startZMQReceive(errs)                //startZMQReceive函数的调用
	// Time it took to start service
	log.Printf("HTTP server listening on port %d, started in: %s", cfg.HTTP.Port, time.Since(start).String())
	// recv error channel
	c := <-errs
	log.Println(fmt.Sprintf("terminating: %v", c))
	os.Exit(0)
}

//有errs会使export-homebridge进程崩掉
func startHTTPServer(errChan chan error, port int) {
	go func() {
		r := zmqinit.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r)) //todo 需要研究?????
	}()
}

//监听中断，如遇到ctrl+c等的可使export-homebridge直接结束掉，前提是export-homebridge得前台单独启动
func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal) //todo  需要研究
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
