package zmqreceivesendhandler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/conthing/export-homebridge/getedgexparams"
	"github.com/conthing/export-homebridge/homebridgeconfig"

	"github.com/conthing/utils/common"
	zmq "github.com/pebbe/zmq4"
)

const (
	CONTROLSTRING = "http://localhost:48082/api/v1/device/"
)

//CommandZmq is the command from zmq
type CommandZmq struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	ID      string `json:"id"`
	Command struct {
		Name   string `json:"name"`
		Params interface{}
	} `json:"command"`
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

var newPublisher *zmq.Socket
var Statuspubport string
var QRcode string

func InitZmq(statusport string) error {
	var err error
	newPublisher, err = zmq.NewSocket(zmq.PUB)
	if err != nil {
		common.Log.Errorf("InitZmq(statusport string) zmq.NewSocket(zmq.PUB) failed: %v", err)
	}
	Statuspubport = statusport
	common.Log.Info("zmq bind to ", statusport)
	_ = newPublisher.Bind(statusport)
	time.Sleep(200 * time.Millisecond) //休眠200ms
	return nil
}
func ZmqInit() error {
	context, err := zmq.NewContext()
	if err != nil {
		common.Log.Errorf("ZmqInit() zmq.NewContext() failed: %v", err)
	}
	commandRep, err := context.NewSocket(zmq.REP)
	if err != nil {
		common.Log.Errorf("ZmqInit() context.NewContext(zmq.REP) failed: %v", err)
	}
	defer func() {
		err = commandRep.Close()
		if err != nil {
			common.Log.Errorf("ZmqInit() commandRep.Close() failed: %v", err)
		}
	}()
	err = commandRep.Connect("tcp://127.0.0.1:9998")
	if err != nil {
		common.Log.Errorf("ZmqInit() commandRep.Connect(tcp://127.0.0.1:9998) failed: %v", err)
	}
	var commandzmq CommandZmq
	for {
		msg, err := commandRep.Recv(0) //recieve message by commandrep
		if err != nil {
			common.Log.Errorf("ZmqInit() commandRep.Recv(0) failed: %v", err)
		}
		msgbyte := []byte(msg)
		err = json.Unmarshal([]byte(msgbyte), &commandzmq)
		if err != nil {
			common.Log.Errorf("ZmqInit() msgbyte json.Unmarshal([]byte(msgbyte), &commandzmq) failed: %v", err)
		}
		common.Log.Info("Got: ", string(msg))
		_, err = commandRep.Send(msg, 0)
		if err != nil {
			common.Log.Errorf("ZmqInit() commandRep.Send(msg, 0) failed: %v", err)
		}
		if commandzmq.Command.Name == "init" {
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string) //todo 类型断言有可能失败
			for i := range homebridgeconfig.Accessarysenders {
				var deviceid = homebridgeconfig.Accessarysenders[i].ID
				for n := range homebridgeconfig.Accessarysenders[i].Commands {
					var commandid = homebridgeconfig.Accessarysenders[i].Commands[n].ID
					statuscommand := commandform(commandid, deviceid)
					common.Log.Info("statuscommand: ", statuscommand)
					result, err := getedgexparams.GetMessage(statuscommand)
					if err != nil {
						common.Log.Errorf("ZmqInit() getedgexparams.GetMessag(statuscommand) failed: %v", err)
					}
					if string(result) != "" {
						err = EventHanler(string(result))
						if err != nil {
							common.Log.Errorf("ZmqInit() EventHanler(string(result)) failed: %v", err)
						}
					}
				}
			}
		} else {
			params, commandname, err := getEdgexParams(commandzmq)
			if err != nil {
				common.Log.Errorf("ZmqInit() getEdgexParams(commandzmq) failed: %v", err)
			}
			id := commandzmq.ID
			go sendcommand(id, params, commandname)
		}
	}
}
func getEdgexParams(commandzmq CommandZmq) (edgexParams string, commandname string, err error) {
	params := commandzmq.Command.Params
	common.Log.Info("params: ", params)
	data := make(map[string]string)
	if params.(map[string]interface{})["onOrOff"] != nil {
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
		common.Log.Info("other type")
	}
	datajson, err := json.Marshal(data)
	if err != nil {
		common.Log.Errorf("getEdgexParams(commandzmq CommandZmq datajson json.Marshal(data) failed: %v", err)
	}
	edgexParams = string(datajson)
	return edgexParams, commandname, nil //返回函数的3个要输出的参数
}
func sendcommand(proxyid string, params string, commandname string) {
	for j := range homebridgeconfig.Accessarysenders {
		deviceid := homebridgeconfig.Accessarysenders[j].ID
		if deviceid == proxyid {
			for k := range homebridgeconfig.Accessarysenders[j].Commands {
				switch homebridgeconfig.Accessarysenders[j].Commands[k].Name {
				case "brightness":
					if commandname == "brightness" {
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case brightness getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					}
				case "percent":
					commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					result, err := getedgexparams.Put(controlcommand, params)
					if err != nil {
						common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case percent getedgexparams.Put failed: %v", err)
					}
					common.Log.Info("put result", string(result))
				case "onoff":
					if commandname == "onoff" {
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case onoff getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					}
				default:
					common.Log.Info("in default")
				}
			}
		}
	}
}
func commandform(commandid string, deviceid string) string {
	controlcommand := CONTROLSTRING + deviceid + "/command/" + commandid
	return controlcommand
}
func EventHanler(bd string) (err error) {
	var event Event
	var status map[string]interface{}
	status = make(map[string]interface{})
	common.Log.Info("收到的event： ", bd)
	bytestr := []byte(bd)
	err = json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		common.Log.Errorf("EventHanler(bd string) bytestr json.Umarshal([]byte(bytestr), &event) failed: %v", err)
	}
	devicename := event.Device
	for i := range homebridgeconfig.Accessaries {
		defaultname := homebridgeconfig.Accessarysenders[i].Name
		defaultid := homebridgeconfig.Accessaries[i].ProxyID
		defaulttype := homebridgeconfig.Accessaries[i].Service
		if devicename == defaultname {
			var dimmerablelightstatus DimmerableLightStatus
			var curtainstatus CurtainStatus
			var lightstatus LightStatus
			for j := range event.Readings {
				switch event.Readings[j].Name {
				case "brightness":
					if homebridgeconfig.Accessaries[i].Dimmerable == "true" {
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
		common.Log.Errorf("EventHanler(bd string) data json.MarshalIndent failed: %v", err)
	}
	if string(data) != "{}" {
		common.Log.Info("send to js ", string(data))
		if newPublisher != nil {
			_, err = newPublisher.SendMessage("status", data)
		}
	}
	return
}
