package zmqinit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	//	"time"

	"github.com/conthing/export-homebridge/pkg/device"
	"github.com/conthing/export-homebridge/pkg/errorHandle"
	httpsender "github.com/conthing/export-homebridge/pkg/http"
	"github.com/gorilla/mux"
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

type LightStatus struct {
	Id             string                `json:"id"`
	Name           string                `json:"name"`
	Service        string                `json:"service"`
	Characteristic StLightCharacteristic `json:"characteristic"`
}

type StLightCharacteristic struct {
	Brightness int  `json:"brightness"`
	On         bool `json:"on"`
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

var Statusport string
var QRcode string
var newPublisher *zmq.Socket

//var statusReq *zmq.Socket

func ZmqInit() error {
	context, err := zmq.NewContext()
	if err != nil {
		return errorHandle.ErrContextFail
	}
	commandRep, err := context.NewSocket(zmq.REP)
	if err != nil {
		return errorHandle.ErrSocketFail
	}
	defer func() {
		err = commandRep.Close()
		if err != nil {
			return
		}
	}()

	// context1,_ := zmq.NewContext()
	// statusReq,_ = context1.NewSocket(zmq.REQ)

	err = commandRep.Connect("tcp://127.0.0.1:9998")
	if err != nil {
		return errorHandle.ErrConnectFail
	}

	var commandzmq CommandZmq

	for {
		msg, err := commandRep.Recv(0) //recieve message by commandrep
		if err != nil {
			return errorHandle.ErrRevFail
		}
		msgbyte := []byte(msg)
		err = json.Unmarshal([]byte(msgbyte), &commandzmq)
		if err != nil {
			log.Println(err)
			return errorHandle.ErrUnmarshalFail
		}
		fmt.Println("Got", string(msg))
		_, err = commandRep.Send(msg, 0)
		if err != nil {
			return errorHandle.ErrSendFail
		}
		if commandzmq.Command.Name == "init" {
			Statusport = commandzmq.Command.Params.(map[string]interface{})["statusport"].(string)
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string)

			newPublisher, err := zmq.NewSocket(zmq.PUB)
			if err != nil {
				return errorHandle.ErrSocketFail
			}
			log.Printf("zmq bind to %s", Statusport)
			err = newPublisher.Bind(Statusport)
			if err != nil {
				return errorHandle.ErrBindFail
			}

			//qrcode := commandzmq.Command.Params.QRcode
			for i := range device.Accessarysenders {
				//	var id = device.Accessaries[i].ProxyID
				//var name = device.Accessarysenders[i].Name
				var deviceid = device.Accessarysenders[i].ID
				for n := range device.Accessarysenders[i].Commands {
					//	switch device.Accessarysenders[i].Commands[n].Name {
					// case "Light":
					// 	var commandid = device.Accessarysenders[i].Commands[n].ID
					// 	statuscommand := commandform(commandid, deviceid)
					// 	result := httpsender.GetMessage(statuscommand)
					// 	fmt.Println("123", result)
					// 	EventHanler(result)
					//		case "brightness":
					var commandid = device.Accessarysenders[i].Commands[n].ID
					statuscommand := commandform(commandid, deviceid)
					result, err := httpsender.GetMessage(statuscommand)
					if err != nil {
						return err
					}
					err = EventHanler(string(result))
					if err != nil {
						return err
					}
					//			case "percent":

					//			default:
					//			}

				}

			}
		} else {
			params, err := getEdgexParams(commandzmq)
			if err != nil {
				return err
			}
			id := commandzmq.ID
			sendcommand(id, params)
		}

	}

}

func getEdgexParams(commandzmq CommandZmq) (edgexParams string, err error) {
	params := commandzmq.Command.Params
	data := make(map[string]string)
	if params.(map[string]interface{})["onOrOff"] != nil {
		onoroff := params.(map[string]interface{})["onOrOff"].(bool)
		if onoroff {
			data["brightness"] = "100"
		} else {
			data["brightness"] = "0"
		}
	} else if params.(map[string]interface{})["percent"] != nil {
		percent := params.(map[string]interface{})["percent"].(float64)
		data["percent"] = strconv.FormatInt(int64(percent), 10)
	} else {
		fmt.Println("other type")
	}
	datajson, err := json.Marshal(data)
	if err != nil {
		return "", errorHandle.ErrMarshalFail
	}
	edgexParams = string(datajson)
	return edgexParams, nil
}

func sendcommand(proxyid string, params string) {
	for j := range device.Accessarysenders {
		deviceid := device.Accessarysenders[j].ID
		if deviceid == proxyid {
			for k := range device.Accessarysenders[j].Commands {
				switch device.Accessarysenders[j].Commands[k].Name {
				case "brightness":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					result, err := httpsender.Put(controlcommand, params)
					if err != nil {
						return
					}
					fmt.Println("put result", string(result))
				case "percent":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					result, err := httpsender.Put(controlcommand, params)
					if err != nil {
						return
					}
					fmt.Println("put result", string(result))
				default:
					fmt.Println("in default")
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

func commandHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("writer", w, "HTTP ", r.Method, " ", r.URL)

	defer func() {
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()
	buf := make([]byte, 1024)  // 1024为缓存大小，即每次读出的最大数据
	n, err := r.Body.Read(buf) // 为这次读出的数据大小
	if err != nil {
		return
	}

	var bd string
	bd = string(buf[:n])
	log.Print("222", bd)

	err = EventHanler(bd)
	if err != nil {
		return
	}
	//4.对收到的event进行处理，然后发给js   status
}

func EventHanler(bd string) (err error) {
	var event Event
	var status map[string]interface{}
	status = make(map[string]interface{})
	bytestr := []byte(bd)
	err = json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		log.Println(err)
		return errorHandle.ErrReadFail
	}
	devicename := event.Device
	for i := range device.Accessaries {
		defaultname := device.Accessaries[i].Name
		defaultid := device.Accessaries[i].ProxyID
		defaulttype := device.Accessaries[i].Service
		if defaultname == devicename {
			var lightstatus LightStatus
			var curtainstatus CurtainStatus
			for j := range event.Readings {
				switch event.Readings[j].Name {
				case "brightness":
					lightstatus.Characteristic.Brightness, _ = strconv.Atoi(event.Readings[j].Value)
					if lightstatus.Characteristic.Brightness > 0 {
						lightstatus.Characteristic.On = true
					} else {
						lightstatus.Characteristic.On = false
					}
					lightstatus.Id = defaultid
					lightstatus.Name = defaultname
					lightstatus.Service = defaulttype
					status["status"] = lightstatus
				case "percent":
					curtainstatus.Characteristic.Percent, _ = strconv.Atoi(event.Readings[j].Value)
					curtainstatus.Id = defaultid
					curtainstatus.Name = defaultname
					curtainstatus.Service = defaulttype
					status["status"] = curtainstatus
				default:
					return
				}
			}

		}
	}

	data, err := json.MarshalIndent(status, "", " ")
	if err != nil {
		return errorHandle.ErrMarshalFail
	}
	if Statusport != "" {
		fmt.Println("send to js ", string(data))
		result, err := newPublisher.SendMessage("status", data)
		if err != nil {
			return errorHandle.ErrSendFail
		}
		fmt.Println(result)
	}

	// if Statusport != "" {
	// 	log.Printf("zmq bind to %s", Statusport)
	// 	statusReq.Bind(Statusport)
	// 	time.Sleep(200 * time.Millisecond)
	// 	fmt.Println("send to js ", string(data))
	// 	statusReq.Send(string(data))
	// }
	return
}

//LoadRestRoutes 定义REST资源
func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)

	return r
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request", r)
	w.Header().Set("Content-Type", "text/plain")
	pincode := device.Pincode
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
