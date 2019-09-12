package zmqinit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/conthing/export-homebridge/pkg/device"
	httpsender "github.com/conthing/export-homebridge/pkg/http"

	"github.com/conthing/utils/common"
	"github.com/gorilla/mux"
	zmq "github.com/pebbe/zmq4"
)

//CommandZmq is the command from zmq
type CommandZmq struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	ID      string `json:"id"`
	Command struct {
		Name   string `json:"name"` //结构体里面嵌套结构体，可以存在两个Name
		Params interface{}
	} `json:"command"` //这个对应的是Command，也是正确的语法
}
type Event struct {
	Device   string
	Readings []Reading
}

//Reading means readings
type Reading struct {
	Name  string
	Value string
}
type DimmerableLightStatus struct {
	Id             string                      `json:"id"`
	Name           string                      `json:"name"`
	Service        string                      `json:"service"`
	Characteristic StDimmerLightCharacteristic `json:"characteristic"`
}
type StDimmerLightCharacteristic struct {
	Brightness int  `json:"brightness"`
	On         bool `json:"on"`
}
type LightStatus struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Service        string                `json:"service"`
	Characteristic StLightCharacteristic `json:"characteristic"`
}
type StLightCharacteristic struct {
	On bool `json:"on"`
}
type CurtainStatus struct {
	Id             string                  `json:"id"`
	Name           string                  `json:"name"`
	Service        string                  `json:"service"`
	Characteristic StCurtainCharacteristic `json:"characteristic"`
}
type StCurtainCharacteristic struct {
	Percent int `json:"percent"`
}

var QRcode string
var newPublisher *zmq.Socket
var statuspubport string //statuspubport是publish的port，在这里是9998
func InitZmq(statusport string) error { //初始化zmq，赋值statuspubport
	var err error
	newPublisher, err = zmq.NewSocket(zmq.PUB) //todo 需要研究  ?????
	if err != nil {
		return err
	}
	statuspubport = statusport
	fmt.Println("zmq bind to ", statusport)
	_ = newPublisher.Bind(statusport)  //todo 需要研究  ?????
	time.Sleep(200 * time.Millisecond) //休眠200ms
	return nil
}
func ZmqInit() error { //zmq初始化，
	context, err := zmq.NewContext() //todo 需要研究 ?????
	if err != nil {
		return err
	}
	commandRep, err := context.NewSocket(zmq.REP) //todo 需要研究?????
	if err != nil {
		return err
	}
	defer func() {
		err = commandRep.Close()
		if err != nil {
			return
		}
	}()
	err = commandRep.Connect("tcp://127.0.0.1:9998")
	if err != nil {
		return err
	}
	var commandzmq CommandZmq
	for {
		msg, err := commandRep.Recv(0) //recieve message by commandrep todo 需要研究  ?????
		if err != nil {
			return err //有err输出err
		}
		msgbyte := []byte(msg)                             //将字符串类型的msg强制转换成为字节数组类型的msgbyte
		err = json.Unmarshal([]byte(msgbyte), &commandzmq) //json非序列化
		if err != nil {                                    //有err输出err
			log.Println(err)
			return err
		}
		fmt.Println("Got", string(msg))  //todo msg本身的类型就是string，为什么还要加上string
		_, err = commandRep.Send(msg, 0) //todo 需要研究  ?????
		if err != nil {
			return err
		}
		if commandzmq.Command.Name == "init" {
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string) //todo  需要研究  ?????
			for i := range device.Accessarysenders {
				var deviceid = device.Accessarysenders[i].ID
				for n := range device.Accessarysenders[i].Commands {
					var commandid = device.Accessarysenders[i].Commands[n].ID
					statuscommand := commandform(commandid, deviceid)
					fmt.Println("statuscommand: ", statuscommand)
					result, err := httpsender.GetMessage(statuscommand)
					if err != nil {
						return err
					}
					if string(result) != "" {
						err = EventHanler(string(result))
						if err != nil {
							return err
						}
					}
				}
			}
		} else {
			params, commandname, err := getEdgexParams(commandzmq)
			if err != nil {
				return err
			}
			id := commandzmq.ID
			go sendcommand(id, params, commandname)
		}
	}
}
func getEdgexParams(commandzmq CommandZmq) (edgexParams string, commandname string, err error) {
	params := commandzmq.Command.Params
	fmt.Println("params: ", params)
	data := make(map[string]string)
	if params.(map[string]interface{})["onOrOff"] != nil { //todo commandname=onoff/percent/brightness 需要研究  ?????
		onoroff := params.(map[string]interface{})["onOrOff"].(bool)
		data["onoff"] = strconv.FormatBool(onoroff)
		commandname = "onoff"
	} else if params.(map[string]interface{})["percent"] != nil {
		percent := params.(map[string]interface{})["percent"].(float64)
		data["percent"] = strconv.FormatInt(int64(percent), 10)
		commandname = "percent"
	} else if params.(map[string]interface{})["brightness"] != nil {
		brightness := params.(map[string]interface{})["brightness"].(float64)
		data["brightness"] = strconv.FormatInt(int64(brightness), 10)
		commandname = "brightness"
	} else {
		fmt.Println("other type")
	}
	datajson, err := json.Marshal(data) //data进行json序列化
	if err != nil {                     //有err输出err
		return "", "", err
	}
	edgexParams = string(datajson)
	return edgexParams, commandname, nil //返回函数的3个要输出的参数
}
func sendcommand(proxyid string, params string, commandname string) {
	for j := range device.Accessarysenders {
		deviceid := device.Accessarysenders[j].ID
		if deviceid == proxyid {
			for k := range device.Accessarysenders[j].Commands {
				switch device.Accessarysenders[j].Commands[k].Name {
				case "brightness":
					if commandname == "brightness" {
						commandid := device.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := httpsender.Put(controlcommand, params)
						if err != nil {
							return
						}
						fmt.Println("put result", string(result))
					}
				case "percent":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					result, err := httpsender.Put(controlcommand, params)
					if err != nil {
						return
					}
					fmt.Println("put result", string(result))
				case "onoff":
					if commandname == "onoff" {
						commandid := device.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := httpsender.Put(controlcommand, params)
						if err != nil {
							return
						}
						fmt.Println("put result", string(result))
					}
				default:
					fmt.Println("in default")
				}
			}
		}
	}
}

//这个commandform看明白了
func commandform(commandid string, deviceid string) string {
	controlstring := "http://localhost:48082/api/v1/device/"
	controlcommand := controlstring + deviceid + "/command/" + commandid
	return controlcommand
}

//todo 需要研究  ?????
func commandHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	log.Print("n: ", n)
	var bd string
	bd = string(buf[:n])
	log.Print("222", bd)
	err := EventHanler(bd)
	if err != nil {
		return
	}
	//4.对收到的event进行处理，然后发给js   status
}

//todo 需要研究  半知半解  ?????
func EventHanler(bd string) (err error) {
	var event Event
	var status map[string]interface{}
	status = make(map[string]interface{})
	fmt.Println("收到的event： ", bd)
	bytestr := []byte(bd)
	err = json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		log.Println(err)
		return err
	}
	devicename := event.Device
	for i := range device.Accessaries {
		defaultname := device.Accessarysenders[i].Name
		defaultid := device.Accessaries[i].ProxyID
		defaulttype := device.Accessaries[i].Service
		if devicename == defaultname {
			var dimmerablelightstatus DimmerableLightStatus
			var curtainstatus CurtainStatus
			var lightstatus LightStatus
			for j := range event.Readings {
				switch event.Readings[j].Name {
				case "brightness":
					if device.Accessaries[i].Dimmerable == "true" {
						dimmerablelightstatus.Characteristic.Brightness, _ = strconv.Atoi(event.Readings[j].Value)
						if dimmerablelightstatus.Characteristic.Brightness > 0 {
							dimmerablelightstatus.Characteristic.On = true
						} else {
							dimmerablelightstatus.Characteristic.On = false
						}
						dimmerablelightstatus.Id = defaultid
						dimmerablelightstatus.Name = defaultname
						dimmerablelightstatus.Service = defaulttype
						status["status"] = dimmerablelightstatus
					}
				case "percent":
					curtainstatus.Characteristic.Percent, _ = strconv.Atoi(event.Readings[j].Value)
					curtainstatus.Id = defaultid
					curtainstatus.Name = defaultname
					curtainstatus.Service = defaulttype
					status["status"] = curtainstatus
				case "onoff":
					lightstatus.Characteristic.On, _ = strconv.ParseBool(event.Readings[j].Value)
					lightstatus.Id = defaultid
					lightstatus.Name = defaultname
					lightstatus.Service = defaulttype
					status["status"] = lightstatus
				default:
					return
				}
			}
		}
	}
	data, err := json.MarshalIndent(status, "", " ")
	if err != nil {
		return err
	}
	if string(data) != "{}" {
		fmt.Println("send to js ", string(data))
		if newPublisher != nil {
			_, err = newPublisher.SendMessage("status", data)
		}
	}
	return
}

////todo LoadRestRoutes最好单独放到一个rest.go
//LoadRestRoutes 定义REST资源  todo  需要研究 ?????
func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/version", versionHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/reboot", rebootHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/ping", pingHandler).Methods(http.MethodGet)
	return r
}

// Respond with PINGRESPONSE to see if the service is alive
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request", r)
	w.Header().Set("Content-Type", "text/plain")
	pincode := device.Pincode
	if pincode == "" {
		fmt.Println("ErrPincodeNil")
		_, err := w.Write([]byte("ErrPincodeNil")) //多个homebridge的数据再组
		if err != nil {
			return
		}
	}
	var data map[string]string = map[string]string{}
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = QRcode
	datasend = append(datasend, data)
	datajson, err := json.MarshalIndent(datasend, "", " ")
	if err != nil {
		return
	}
	_, err = w.Write([]byte(datajson)) //多个homebridge的数据再组
	if err != nil {
		return
	}
}
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	version := common.Version
	currentTime := common.BuildTime
	fmt.Println("version", version)
	fmt.Println("Time", currentTime)
	datastring := strings.Join([]string{version, currentTime}, " ")
	_, err := w.Write([]byte(datastring)) //多个homebridge的数据再组
	if err != nil {
		return
	}
}
func rebootHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	var bd string
	bd = string(buf[:n])
	log.Print("333", bd)
	device.Accessaries = nil
	device.Accessarysenders = nil
	labels := []string{"Light", "Curtain"}
	for _, label := range labels {
		projectUrl := "http://localhost:52030/api/v1/project/" + label
		var projectlist, _ = httpsender.GetMessage(projectUrl)
		_ = device.Decode(projectlist, label, statuspubport)
	}
}
