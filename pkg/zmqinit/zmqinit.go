package zmqinit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/conthing/export-homebridge/pkg/device"
	httpsender "github.com/conthing/export-homebridge/pkg/http"
	"github.com/gorilla/mux"
	zmq "github.com/pebbe/zmq4"
)

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

type Event struct {
	Device   string
	Readings []Reading
}

//Reading means readings
type Reading struct {
	Name  string
	Value string
}

var Statusport string
var QRcode string

func ZmqInit() {
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
			Statusport = commandzmq.Command.Params.(map[string]interface{})["statusport"].(string)
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
						result := httpsender.GetMessage(statuscommand)
						fmt.Println("123", result)
						EventHanler(result)
					case "Brightness":
						var commandid = device.Accessarysenders[i].Commands[n].ID
						statuscommand := commandform(commandid, deviceid)
						result := httpsender.GetMessage(statuscommand)
						fmt.Println(result)
						EventHanler(result)
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
					httpsender.Put(controlcommand, params)
				case "Brightness":
					commandid := device.Accessarysenders[j].Commands[k].ID
					controlcommand := commandform(commandid, deviceid)
					httpsender.Put(controlcommand, params)
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

func commandHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("HTTP ", r.Method, " ", r.URL)

	defer r.Body.Close()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小

	var bd string
	bd = string(buf[:n])
	log.Print("222", bd)

	EventHanler(bd)
	//4.对收到的event进行处理，然后发给js   status
}

func EventHanler(bd string) {
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
	if Statusport != "" {
		statusReq.Connect(Statusport)
		statusReq.Send(string(data), 0)
		defer statusReq.Close()
	}
	fmt.Println(string(data))
}

//LoadRestRoutes 定义REST资源
func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)

	return r
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	pincode := device.Pinbuffer.String()
	var data map[string]string = map[string]string{}
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = QRcode
	datasend = append(datasend, data)
	datajson, _ := json.MarshalIndent(datasend, "", " ")
	w.Write([]byte(datajson)) //多个homebridge的数据再组
}
