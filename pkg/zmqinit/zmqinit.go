package zmqinit

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/conthing/export-homebridge/pkg/device"
	httpsender "github.com/conthing/export-homebridge/pkg/http"

	"github.com/conthing/utils/common"
	zmq "github.com/pebbe/zmq4"
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
		return err
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
		return err
	}
	commandRep, err := context.NewSocket(zmq.REP)
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
		msg, err := commandRep.Recv(0) //recieve message by commandrep
		if err != nil {
			return err
		}
		msgbyte := []byte(msg)
		err = json.Unmarshal([]byte(msgbyte), &commandzmq)
		if err != nil {
			common.Log.Error(err)
			return err
		}
		common.Log.Info("Got: ", string(msg))
		_, err = commandRep.Send(msg, 0)
		if err != nil {
			return err
		}
		if commandzmq.Command.Name == "init" {
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string)
			for i := range device.Accessarysenders {
				var deviceid = device.Accessarysenders[i].ID
				for n := range device.Accessarysenders[i].Commands {
					var commandid = device.Accessarysenders[i].Commands[n].ID
					statuscommand := commandform(commandid, deviceid)
					common.Log.Info("statuscommand: ", statuscommand)
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
						common.Log.Info("put result", string(result))
					}
				case "percent":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					result, err := httpsender.Put(controlcommand, params)
					if err != nil {
						return
					}
					common.Log.Info("put result", string(result))
				case "onoff":
					if commandname == "onoff" {
						commandid := device.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := httpsender.Put(controlcommand, params)
						if err != nil {
							return
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
	controlstring := "http://localhost:48082/api/v1/device/"
	controlcommand := controlstring + deviceid + "/command/" + commandid
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
		common.Log.Error(err)
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
		common.Log.Info("send to js ", string(data))
		if newPublisher != nil {
			_, err = newPublisher.SendMessage("status", data)
		}
	}
	return
}
