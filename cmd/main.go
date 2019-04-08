package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	device "github.com/conthing/careyes-driver/pkg/device"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	zmq "github.com/pebbe/zmq4"
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

//CommandZmq is the command from zmq
type CommandZmq struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	ProxyID string `json:"proxy_id"`
	Command struct {
		Name   string `json:"name"`
		Params interface{}
	} `json:"command"`
}

//Status is the data which is ready to send to js
type Status struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Service        string `json:"service"`
	Characteristic map[string]interface{}
}

//Event means events from coredata
type Event struct {
	Device   string
	Readings []Reading
}

//Reading means readings
type Reading struct {
	Name  string
	Value string
}

var statusport string
var QRcode string

func commandHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("HTTP ", r.Method, " ", r.URL)

	defer r.Body.Close()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小

	var bd string
	bd = string(buf[:n])
	log.Print("222", bd)

	eventHanler(bd)
	//4.对收到的event进行处理，然后发给js   status
}

func eventHanler(bd string) {
	context1, _ := zmq.NewContext()
	statusReq, _ := context1.NewSocket(zmq.REQ)
	var event Event
	var status map[string]interface{}
	var statuses []map[string]interface{}
	status = make(map[string]interface{})
	bytestr := []byte(bd)
	err := json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		log.Fatal(err)
	}
	devicename := event.Device
	for i := range device.Accessaries {
		defaultname := device.Accessaries[i].Name
		defaultid := device.Accessaries[i].ProxyID
		defaulttype := device.Accessaries[i].Service
		fmt.Println(defaultid, defaulttype, defaultname)
		if defaultname == devicename {
			var list map[string]interface{}
			list = make(map[string]interface{})
			list["id"] = defaultid
			list["name"] = defaultname
			list["service"] = defaulttype
			//zu bao
			var reading map[string]interface{}
			reading = make(map[string]interface{})
			for j := range event.Readings {
				readingname := event.Readings[j].Name
				readingvalue := event.Readings[j].Value
				reading[readingname] = readingvalue

			}
			list["characteristic"] = reading
			statuses = append(statuses, list)
		}

		status["status"] = statuses
	}
	data, _ := json.MarshalIndent(status, "", " ")
	if statusport != "" {
		statusReq.Connect(statusport)
		statusReq.Send(string(data), 0)
		defer statusReq.Close()
	}
	fmt.Println(string(data))
}

func main() {
	start := time.Now()
	var profile string

	flag.StringVar(&profile, "profile", "config.json", "Specify a profile other than default.")
	flag.StringVar(&profile, "p", "config.json", "Specify a profile other than default.")
	flag.Parse()

	cfg := &Config{}

	httpPost()

	go zmqInit()

	//ReadFile函数会读取文件的全部内容，并将结果以[]byte类型返回
	data, err := ioutil.ReadFile(profile)
	if err != nil {
		log.Fatal(err)
		return
	}

	//读取的数据为json格式，需要进行解码
	err = json.Unmarshal(data, cfg)
	if err != nil {
		log.Fatal(err)
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
		r := LoadRestRoutes()
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

//LoadRestRoutes 定义REST资源
func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/homebridge", qrcodeHandler).Methods(http.MethodGet)

	return r
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	pincode := device.Pinbuffer.String()
	var data map[string]string
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = QRcode
	datasend = append(datasend, data)
	datajson, _ := json.MarshalIndent(datasend, "", " ")
	w.Write([]byte(datajson)) //多个homebridge的数据再组
}

func httpPost() {

	str := "{\"origin\":1471806386919,\"name\":\"RESTXMLClient\",\"addressable\":{\"origin\":1471806386919,\"name\":\"EdgeXTestRESTXML\",\"protocol\":\"HTTP\",\"method\": \"POST\",\"address\":\"localhost\",\"port\":8111,\"path\":\"/rest\"},\"format\":\"JSON\",\"enable\":true,\"destination\":\"REST_ENDPOINT\"}"

	var jsonstr = []byte(str)
	resp, err := http.Post("http://192.168.56.108:48071/api/v1/registration",
		"application/json",
		bytes.NewBuffer(jsonstr))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	fmt.Println(string(body))

	devicelisturl := "http://192.168.56.108:48081/api/v1/device"
	var devicelist = getMessage(devicelisturl)
	device.Decode([]byte(devicelist))
	fmt.Println(devicelist)
}

func getMessage(msg string) string {
	resp, err := http.Get(msg)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println(err)
	}

	result := string(body)

	fmt.Println(string(body))

	return result
}

func zmqInit() {
	context, _ := zmq.NewContext()
	commandRep, _ := context.NewSocket(zmq.REP)
	commandRep.Bind("tcp://127.0.0.1:9998")
	var commandzmq CommandZmq

	for {
		msg, _ := commandRep.Recv(0) //recieve message by commandrep

		msgbyte := []byte(msg)
		err := json.Unmarshal([]byte(msgbyte), &commandzmq)
		if err != nil {
			log.Fatal(err)
		}
		println("Got", string(msg))
		commandRep.Send(msg, 0)

		if commandzmq.Command.Name == "init" {
			statusport = commandzmq.Command.Params.(map[string]interface{})["statusport"].(string)
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string)

			fmt.Println(QRcode)

			//qrcode := commandzmq.Command.Params.QRcode
			for i := range device.Accessarysenders {
				//	var id = device.Accessaries[i].ProxyID
				//var name = device.Accessarysenders[i].Name
				var deviceid = device.Accessarysenders[i].ID
				for n := range device.Accessarysenders[i].Commands {
					switch device.Accessarysenders[i].Commands[n].Name {
					case "Light":
						var commandid = device.Accessarysenders[i].Commands[n].ID
						statuscommand := commandform(commandid, deviceid)
						result := getMessage(statuscommand)
						fmt.Println("123", result)
						eventHanler(result)
					case "Brightness":
						var commandid = device.Accessarysenders[i].Commands[n].ID
						statuscommand := commandform(commandid, deviceid)
						result := getMessage(statuscommand)
						fmt.Println(result)
						eventHanler(result)
					case "Percent":

					default:
					}

				}

			}
		} else {
			params := commandzmq.Command.Params.(string)
			proxyid := commandzmq.ProxyID
			fmt.Println(params)
			sendcommand(proxyid, params)
			//			switch commandzmq.Service {
			//			case "LightBulb":
			//				sendcommand(proxyid, params)
			//				fmt.Println("sevice is Light")
			//			case "WindowCovering":
			//				fmt.Println("sevice is WindowCovering")
			//			case "Thermostat":
			//				fmt.Println("sevice is LiThermostatght")
			//			case "Window":
			//				fmt.Println("sevice is Window")
			//			case "Door":
			//				fmt.Println("sevice is Door")
			//			case "Lock":
			//				fmt.Println("sevice is Lock")
			//			case "Switch":
			//				fmt.Println("sevice is Switch")
			//			case "Fan":
			//				fmt.Println("sevice is Fan")
			//			case "Fanv2":
			//				fmt.Println("sevice is Fanv2")
			//			default:
			//				fmt.Println("default")
			//			} //根据不同类型进行分类 switch case

			//name := commandzmq.Name
			//switch commandzmq.Name {
			//case "Light":

			//case "Curtain":

			//}

		}

		//发送具体的命令

	}

}

func sendcommand(proxyid string, params string) {
	for j := range device.Accessarysenders {
		deviceid := device.Accessarysenders[j].ID
		if deviceid == proxyid {
			for k := range device.Accessarysenders[j].Commands {
				switch device.Accessarysenders[j].Commands[k].Name {
				case "Light":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					put(controlcommand, params)
				case "Brightness":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					put(controlcommand, params)
				case "Percent":

				default:
				}

			}
		}
	}
}

func commandform(commandid string, deviceid string) string {
	controlstring := "http://localhost:48082/api/v1/device/"
	controlcommand := controlstring + deviceid + "/command/" + commandid
	return controlcommand
}

func put(commandstring string, params string) {

	payload := strings.NewReader(params)

	req, _ := http.NewRequest("PUT", commandstring, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
}
