package zmqinit

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/conthing/export-homebridge/pkg/device"
	httpsender "github.com/conthing/export-homebridge/pkg/http"
	"github.com/gorilla/mux"
	zmq "github.com/pebbe/zmq4"
)

//CommandZmq is the command from zmq
type CommandZmq struct {
	Name    string `json:"name"`
	Service string `json:"service"`
	ID string `json:"id"`
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

type Status struct {
	Id              string           `json:"id"`
	Name            string           `json:"name"`
	Service         string           `json:"service"`
	Characteristic  StCharacteristic `json:"characteristic"`
}

type StCharacteristic struct {
	Brightness     int         `json:"brightness"`
	Percent        int		   `json:"percent"` 
}

var Statusport string
var QRcode string
var newPublisher *zmq.Socket

func ZmqInit() {
	context, _ := zmq.NewContext()
	commandRep, _ := context.NewSocket(zmq.REP)
	defer commandRep.Close()

	newPublisher, _ = zmq.NewSocket(zmq.PUB)

	commandRep.Connect("tcp://127.0.0.1:9998")

	var commandzmq CommandZmq

	for {
		msg, _ := commandRep.Recv(0) //recieve message by commandrep

		msgbyte := []byte(msg)
		err := json.Unmarshal([]byte(msgbyte), &commandzmq)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("Got", string(msg))
		commandRep.Send(msg, 0)
		fmt.Println("send ok!")
		if commandzmq.Command.Name == "init" {
			Statusport = commandzmq.Command.Params.(map[string]interface{})["statusport"].(string)
			QRcode = commandzmq.Command.Params.(map[string]interface{})["QRcode"].(string)

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
						result := httpsender.GetMessage(statuscommand)
						EventHanler(result)
		//			case "percent":

		//			default:
		//			}

				}

			}
		} else {
			params := getEdgexParams(commandzmq)
			id := commandzmq.ID
			sendcommand(id, params)
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

func getEdgexParams(commandzmq CommandZmq) string{
	params := commandzmq.Command.Params
	var edgexParams string
	data :=make(map[string]string)
if params.(map[string]interface{})["onOrOff"] != nil{
onoroff := params.(map[string]interface{})["onOrOff"].(bool)
if onoroff {
data["brightness"] = "100"
}else{
	data["brightness"] = "0"
}
}else if params.(map[string]interface{})["percent"] != nil{
	 percent := params.(map[string]interface{})["percent"].(float64)
	 data["percent"] = strconv.FormatInt(int64(percent), 10)
}else{
fmt.Println("other type")
}
datajson,_ :=json.Marshal(data)
edgexParams =string(datajson)
return edgexParams
}

func sendcommand(proxyid string, params string) {
	for j := range device.Accessarysenders {
		deviceid := device.Accessarysenders[j].ID
		if deviceid == proxyid {
			for k := range device.Accessarysenders[j].Commands {
				commandid := device.Accessarysenders[j].Commands[k].ID
				controlcommand := commandform(commandid, deviceid)
				httpsender.Put(controlcommand, params)
				// switch device.Accessarysenders[j].Commands[k].Name {
				// case "Light":
				// 	fmt.Println("in Light")
				// 	commandid := device.Accessarysenders[j].Commands[k].ID
				// 	controlcommand := commandform(commandid, deviceid)
				// 	httpsender.Put(controlcommand, params)
				// case "brightness":
				// 	fmt.Println("in Brightness")
				// 	commandid := device.Accessarysenders[j].Commands[k].ID
				// 	controlcommand := commandform(commandid, deviceid)
				// 	httpsender.Put(controlcommand, params)
				// case "Percent":
				// 	fmt.Println("in Brightness")
				// 	commandid := device.Accessarysenders[j].Commands[k].ID
				// 	controlcommand := commandform(commandid, deviceid)
				// 	httpsender.Put(controlcommand, params)
				// default:
				// 	fmt.Println("in default")
				// }

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
	var event Event
	var status map[string]interface{}
	status = make(map[string]interface{})
	bytestr := []byte(bd)
	err := json.Unmarshal([]byte(bytestr), &event)
	if err != nil {
		log.Println(err)
		return
	}
	devicename := event.Device
	for i := range device.Accessaries {
		defaultname := device.Accessaries[i].Name
		defaultid := device.Accessaries[i].ProxyID
		defaulttype := device.Accessaries[i].Service
		if defaultname == devicename {
			var ststatus Status
			for j := range event.Readings {
				switch event.Readings[j].Name{
				case "brightness":
					ststatus.Characteristic.Brightness,_ = strconv.Atoi(event.Readings[j].Value)
				case "percent":
					ststatus.Characteristic.Percent,_ = strconv.Atoi(event.Readings[j].Value)
				default:
					return
				}
			}

			ststatus.Id = defaultid
			ststatus.Name = defaultname
			ststatus.Service = defaulttype
			status["status"] = ststatus
		}
	}

	data, _ := json.MarshalIndent(status, "", " ")
	if Statusport != "" {
		log.Printf("zmq bind to %s", Statusport)
		_ = newPublisher.Bind(Statusport)
		time.Sleep(200 * time.Millisecond)
		fmt.Println("send to js ", string(data))
		_, _ = newPublisher.SendMessage("status", data)
	}


}

//LoadRestRoutes 定义REST资源
func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)

	return r
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	pincode := device.Pincode
	var data map[string]string = map[string]string{}
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = QRcode
	datasend = append(datasend, data)
	datajson, _ := json.MarshalIndent(datasend, "", " ")
	w.Write([]byte(datajson)) //多个homebridge的数据再组
}
