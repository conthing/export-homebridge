package zmqreceivesendhandler

import (
	"encoding/json"
	"github.com/conthing/export-homebridge/dto"
	"github.com/conthing/export-homebridge/errors"
	"github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/conthing/export-homebridge/getedgexparams"
	"github.com/conthing/export-homebridge/homebridgeconfig"

	"github.com/conthing/utils/common"
	zmq "github.com/pebbe/zmq4"
)

const (
	CONTROLSTRING      = "http://localhost:48082/api/v1/device/"
	GETDEVICEBYNAMEURL = CONTROLSTRING + "name/"
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

//定义空调的状态结构体
type HvacStatus struct {
	Id             string               `json:"id"`
	Name           string               `json:"name"`
	Service        string               `json:"service"`
	Characteristic StHvacCharacteristic `json:"characteristic"`
}
type StHvacCharacteristic struct {
	On       bool   `json:"on"`
	Ttarget  string `json:"ttarget"`
	Mode     string `json:"mode"`
	Fanlevel string `json:"fanlevel"`
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
			value, ok := commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string)
			if !ok {
				return errors.QRCodeAssertErr
			}
			QRcode = value

			for i := range homebridgeconfig.Accessarysenders {
				var deviceID = homebridgeconfig.Accessarysenders[i].ID
				for n := range homebridgeconfig.Accessarysenders[i].Commands {
					var commandID = homebridgeconfig.Accessarysenders[i].Commands[n].ID
					coreCommandURL := commandform(commandID, deviceID)
					//common.Log.Info("coreCommandURL: ", coreCommandURL)
					result, err := getedgexparams.GetMessage(coreCommandURL)
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
			getEdgexParams(commandzmq)
		}
	}
}
func getEdgexParams(commandzmq CommandZmq) {
	commandname := ""
	id := commandzmq.ID
	params := commandzmq.Command.Params
	common.Log.Info("params: ", params)
	data := make(map[string]string)
	if params.(map[string]interface{})["onOrOff"] != nil {
		onoroff := params.(map[string]interface{})["onOrOff"].(bool)
		data["onoff"] = strconv.FormatBool(onoroff)
		commandname = "onoff"
		go sendcommand(id, data, commandname)
	} else if params.(map[string]interface{})["percent"] != nil {
		percent := params.(map[string]interface{})["percent"].(float64)
		data["percent"] = strconv.FormatInt(int64(percent), 10)
		commandname = "percent"
		go sendcommand(id, data, commandname)
	} else if params.(map[string]interface{})["brightness"] != nil {
		brightness := params.(map[string]interface{})["brightness"].(float64)
		data["brightness"] = strconv.FormatInt(int64(brightness), 10)
		commandname = "brightness"
		go sendcommand(id, data, commandname)
	} else if params.(map[string]interface{})["t_target"] != nil {
		ttarget := params.(map[string]interface{})["t_target"].(float64)
		data["ttarget"] = strconv.FormatInt(int64(ttarget), 10)
		commandname = "ttarget"
		go sendcommand(id, data, commandname)
	} else if params.(map[string]interface{})["mode"] != nil {
		mode := params.(map[string]interface{})["mode"].(string)
		switch mode {
		case "HEAT":
			go sendcommand(id, map[string]string{"onoff": "true"}, "onoff")
			data["mode"] = "HEATER"
			commandname = "mode"
			go sendcommand(id, data, commandname)
		case "OFF":
			data["onoff"] = "false"
			commandname = "onoff"
			go sendcommand(id, data, commandname)
		case "COOL":
			go sendcommand(id, map[string]string{"onoff": "true"}, "onoff")
			data["mode"] = "AC"
			commandname = "mode"
			go sendcommand(id, data, commandname)
		case "AUTO":
			go sendcommand(id, map[string]string{"onoff": "true"}, "onoff")
			data["mode"] = "AUTO"
			commandname = "mode"
			go sendcommand(id, data, commandname)
		}
	} else if params.(map[string]interface{})["fanlevel"] != nil {
		fanlevel := params.(map[string]interface{})["fanlevel"].(string)
		data["fanlevel"] = string(fanlevel) //加的空调的风速设置fanlevel，fanlevel属性是string,输入string,输出也是string，fanlevel取值"LOW,MIDDLE,HIGH"
		commandname = "fanlevel"
		go sendcommand(id, data, commandname)
	} else {
		common.Log.Info("other type")
	}
}
func sendcommand(proxyid string, data map[string]string, commandname string) {
	datajson, err := json.Marshal(data)
	if err != nil {
		common.Log.Errorf("json.Marshal(data) failed: %v", err)
	}
	params := string(datajson)
	common.Log.Debugf("sendcommand(%s, %s, %s)", proxyid, params, commandname)
	for j := range homebridgeconfig.Accessarysenders {
		deviceid := homebridgeconfig.Accessarysenders[j].ID
		if deviceid == proxyid {
			//common.Log.Info("deviceid: ", deviceid, commandname, params)
			for k := range homebridgeconfig.Accessarysenders[j].Commands {
				if homebridgeconfig.Accessarysenders[j].Commands[k].Name == commandname {
					switch commandname {
					case "brightness":
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case brightness getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					case "percent":
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case percent getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					case "onoff":
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case onoff getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					case "ttarget": //sendcommand加的空调的温度设置ttarget
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case ttarget getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					case "mode": //sendcommand加的空调的模式设置mode
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case mode getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					case "fanlevel": //sendcommand加的空调的风速设置fanlevel
						commandid := homebridgeconfig.Accessarysenders[j].Commands[k].ID
						controlcommand := commandform(commandid, deviceid)
						result, err := getedgexparams.Put(controlcommand, params)
						if err != nil {
							common.Log.Errorf("sendcommand(proxyid string, params string, commandname string) case fanlevel getedgexparams.Put failed: %v", err)
						}
						common.Log.Info("put result", string(result))
					default:
						common.Log.Info("in default")
					}
					return
				}
			}
			common.Log.Errorf("command %s not found, commands %+v", commandname, homebridgeconfig.Accessarysenders[j].Commands)
			return
		}
	}
	common.Log.Errorf("proxyid %s not found, accsender %+v", proxyid, homebridgeconfig.Accessarysenders)
}
func commandform(commandid string, deviceid string) string {
	controlcommand := CONTROLSTRING + deviceid + "/command/" + commandid
	return controlcommand
}

/**  --------------------------------------以下代码发给homebridge------------------------------------------------ **/
func EventHanler(bd string) (err error) { //edgex将初始状态传递给onoff、percent、fanlevel、mode、brightness等
	/* ------------ 转化event数据结构------------- */
	common.Log.Info("收到的event： ", bd)
	var event Event
	bytestr := []byte(bd)
	err = json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		common.Log.Errorf("EventHanler(bd string) bytestr json.Umarshal([]byte(bytestr), &event) failed: %v", err)
		return err
	}
	/* ------------ 转成homebridge的协议格式------------- */
	for _, accessary := range homebridgeconfig.Accessaries {
		// 根据 event.name 找到 id
		content, err := FetchContent(GETDEVICEBYNAMEURL + event.Device)
		if err != nil {
			return err
		}
		id := jsoniter.Get(content, "id").ToString()
		// 找到同一个id
		if accessary.ProxyID != id {
			continue
		}
		service := accessary.Service
		if service == "Lightbulb" {
			if accessary.Dimmerable == "true" {
				DimmerableLightService(accessary, event)
			} else {
				LightService(accessary, event)
			}
		}
		if service == "WindowCovering" {
			CurtainService(accessary, event)
		}
		if service == "Thermostat" {
			HVACService(accessary, event)
		}
	}
	return nil
}

func DimmerableLightService(accessary homebridgeconfig.Accessary, event Event) {
	status := make(map[string]interface{})
	reading := event.Readings[0]
	var dimmerablelightstatus DimmerableLightStatus
	dimmerablelightstatus.Characteristic.Brightness, _ = strconv.Atoi(reading.Value)
	if dimmerablelightstatus.Characteristic.Brightness > 0 {
		dimmerablelightstatus.Characteristic.On = true
	} else {
		dimmerablelightstatus.Characteristic.On = false
	}
	dimmerablelightstatus.Id = accessary.ProxyID
	dimmerablelightstatus.Name = accessary.Name
	dimmerablelightstatus.Service = accessary.Service
	status["status"] = dimmerablelightstatus
	sendToHomebridge(status)
}

func LightService(accessary homebridgeconfig.Accessary, event Event) {
	status := make(map[string]interface{})
	var lightstatus LightStatus
	reading := event.Readings[0]
	if reading.Name == "onoff" {
		lightstatus.Characteristic.On, _ = strconv.ParseBool(reading.Value)
		lightstatus.Id = accessary.ProxyID
		lightstatus.Name = accessary.Name
		lightstatus.Service = accessary.Service
		status["status"] = lightstatus
	}
	sendToHomebridge(status)
}

func CurtainService(accessary homebridgeconfig.Accessary, event Event) {
	var curtainstatus CurtainStatus
	status := make(map[string]interface{})
	reading := event.Readings[0]
	if reading.Name == "percent" {
		curtainstatus.Characteristic.Percent, _ = strconv.Atoi(reading.Value)
		curtainstatus.Id = accessary.ProxyID
		curtainstatus.Name = accessary.Name
		curtainstatus.Service = accessary.Service
		status["status"] = curtainstatus
	}
	sendToHomebridge(status)
}

func HVACService(accessary homebridgeconfig.Accessary, event Event) {
	var hvacstatus HvacStatus
	status := make(map[string]interface{})
	reading := event.Readings[0]
	name := event.Device

	switch reading.Name {
	case "onoff":
		if reading.Value == "false" {
			hvacstatus.Characteristic.Mode = "OFF"
		}

		if reading.Value == "true" {
			content, err := FetchContent(GETDEVICEBYNAMEURL + name)
			if err != nil {
				return
			}
			id := jsoniter.Get(content, "id").ToString()
			if id == "" {
				common.Log.Error(" id 为空")
				return
			}
			url := FindSingleDeviceCommands(content, id)
			if url == "" {
				common.Log.Error("URL 为空")
				return
			}
			data, err := FetchContent(url)
			common.Log.Info(string(data))
			if err != nil {
				return
			}
			modeValue := jsoniter.Get(data, "readings", "0", "value").ToString()
			hvacstatus.Characteristic.Mode = EdgexToHomebridgeHvacModeMap[modeValue]

		}
		hvacstatus.Id = accessary.ProxyID
		hvacstatus.Name = accessary.Name
		hvacstatus.Service = accessary.Service
		status["status"] = hvacstatus
	case "mode":
		hvacstatus.Characteristic.Mode = reading.Value
		hvacstatus.Id = accessary.ProxyID
		hvacstatus.Name = accessary.Name
		hvacstatus.Service = accessary.Service
		status["status"] = hvacstatus
	case "fanlevel":
		hvacstatus.Characteristic.Fanlevel = reading.Value
		hvacstatus.Id = accessary.ProxyID
		hvacstatus.Name = accessary.Name
		hvacstatus.Service = accessary.Service
		status["status"] = hvacstatus
	case "ttarget":
		hvacstatus.Characteristic.Ttarget = reading.Value //温度为整数
		hvacstatus.Id = accessary.ProxyID
		hvacstatus.Name = accessary.Name
		hvacstatus.Service = accessary.Service
		status["status"] = hvacstatus
	}

	sendToHomebridge(status)
}

func sendToHomebridge(status map[string]interface{}) {
	/* ------------ 发送 data------------- */
	data, err := json.MarshalIndent(status, "", " ")
	common.Log.Info("data: ", data)
	if err != nil {
		common.Log.Errorf("EventHanler(bd string) data json.MarshalIndent failed: %v", err)
	}
	if string(data) != "{}" {
		common.Log.Info("send to homebridge ", string(data))
		if newPublisher != nil {
			_, err = newPublisher.SendMessage("status", data)
		}
	}
}

// FindSingleDeviceCommands 针对 、GETDEVICEBYNAMEURL 获取commands
func FindSingleDeviceCommands(content []byte, id string) string {
	var device dto.EdgexCommandDevice
	json.Unmarshal(content, device)
	common.Log.Info(device)
	for _, command := range device.Commands {
		if command.Name == "mode" {
			return command.GET.URL
		}
	}
	return ""
}

// FetchContent （通常为第一步）拉取内容
func FetchContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		common.Log.Error(" |Fetch Content ERROR| ")
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			common.Log.Error(" |Close Body ERROR| ")
		}
	}()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.Log.Error(" |Read Body ERROR| ")
		return nil, err
	}
	return content, nil
}
