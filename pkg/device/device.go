package device

import (
	"encoding/json"
	"fmt"
	"github.com/conthing/export-homebridge/pkg/errorHandle"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

//var ErrUnknownType = errors.New("ErrUnknownType")

//AutoGenerated means the construction of config.json
type AutoGenerated struct {
	Description string       `json:"description"`
	Bridge      BridgeStruct `json:"bridge"`
	Platforms   []Platform   `json:"platforms"`
}

type BridgeStruct struct {
	Serialnumber string `json:"serialNumber"`
	Pin          string `json:"pin"`
	Port         int    `json:"port"`
	Name         string `json:"name"`
	Model        string `json:"model"`
	Manufacturer string `json:"manufacturer"`
	Username     string `json:"username"`
	Repport      string `json:"repport"`
}

type Platform struct {
	Accessories []Accessary `json:"accessories"`
	Name        string      `json:"name"`
	ConfigPath  string      `json:"configPath"`
	Platform    string      `json:"platform"`
}

type Accessary struct {
	Service   string `json:"service"`
	Name      string `json:"name"`
	ProxyID   string `json:"proxy_id"`
	Accessory string `json:"accessory"`
	Dimmerable string  `json:"dimmerable,omitempty"`
}

//Envelope means the data transformed from coredata
type Envelope []struct {
	Name    string
	ID      string
	Profile Profile
}

type Profile struct {
	Name     string
	Commands []Commands
}

//Command means control
type Commands struct {
	ID   string
	Name string
}

type Accessarysender struct {
	Service  string
	Name     string
	ID       string
	Commands []Commands
}

var Accessaries []Accessary
var Accessarysenders []Accessarysender
var Pincode string

func Decode(jsonStr []byte, label string, statusport string) {

	var projects []Project

	err := json.Unmarshal(jsonStr, &projects)
	if err != nil {
		log.Println(err)
		return
	}
	index := 0
	// for the love of Gopher DO NOT DO THIS

	for _, project := range projects {
		var accessary Accessary
		var accessarysender Accessarysender
		accessary.ProxyID = project.Id
		accessary.Accessory = "Control4"
		commands := accessarysender.Commands
		
		switch label {
		case "Light":
			accessary.Service = "Lightbulb"
		case "Curtain":
			accessary.Service = "WindowCovering"
		default:
			fmt.Println("不存相应设备")

		}

		for _, projectcommand := range project.Commands {
			if projectcommand.Name == "alias" {
				if Accessaries != nil{
					for _,access := range Accessaries {
						if access.Name != projectcommand.Value{
							accessary.Name = projectcommand.Value
						}else{
							accessary.Name = fmt.Sprintf("%s(%d)",projectcommand.Value,index)
							fmt.Println("accessary.Name: ",accessary.Name)
							index++
							break
						}
					}
				}else{
					accessary.Name = projectcommand.Value
				}

			}else if projectcommand.Name == "dimmerable"{
				accessary.Dimmerable = projectcommand.Value
			}

			var command Commands
			command.ID = projectcommand.Id
			command.Name = projectcommand.Name
			commands = append(commands, command)
		}
		accessarysender.Commands = commands
		accessarysender.Name = project.Name
		accessarysender.ID = project.Id
		accessarysender.Service = label

		Accessaries = append(Accessaries, accessary)
		Accessarysenders = append(Accessarysenders, accessarysender) //store deviceid and commandid
	}
	configdata, err := createConfigData(Accessaries, statusport)
	if err != nil {
		log.Fatal(err)
	}
	b, _ := json.MarshalIndent(configdata, "", " ") //变成json字符串

	err = ioutil.WriteFile("/root/.homebridge/config.json", b, os.ModeAppend) //create config.json
	if err != nil {
		fmt.Println(err)
		return
	}

	return
}

func createConfigData(accessaries []Accessary, statusport string) (configdata AutoGenerated, err error) {
	mac := mac()
	pinstring := strings.Split(mac, ":")
	pinnum := make([]int, 6)
	if len(pinstring) == 6 {
		for i, pin := range pinstring {
			n, err := strconv.ParseUint(pin, 16, 8)
			if err != nil {
				return configdata, errorHandle.ErrParseFail
			} else {
				pinnum[i] = int(n)
			}
		}
	} else {
		err = errorHandle.ErrMacInvalid
		return
	}
	username := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", pinnum[5], pinnum[4], pinnum[3], pinnum[2], pinnum[1], pinnum[0])

	Pincode = fmt.Sprintf("%03d-%02d-%03d", pinnum[5]%90+10, pinnum[4]%90+10, pinnum[3]+100)
	sernum := fmt.Sprintf("%02d.%02d.%02d.%02d", pinnum[3]%100, pinnum[2]%100, pinnum[1]%100, pinnum[0]%100)

	configdata = AutoGenerated{
		Description: "This is an inSona plugin configuration file",
		Bridge: BridgeStruct{
			Serialnumber: sernum,
			Pin:          Pincode,
			Port:         51826,
			Name:         "homebridge-0",
			Model:        "homebridge-inSona",
			Manufacturer: "inSona",
			Username:     username,
			Repport:      statusport,
		},
		Platforms: []Platform{
			{
				Accessories: accessaries,
				Name:        "Control4",
				ConfigPath:  "/root/.homebridge/config.json",
				Platform:    "Control4",
			},
		},
	}
	return
}

func mac() (mac string) {
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	for _, inter := range interfaces {
		//fmt.Println(inter.Name)
		mac = strings.ToUpper(inter.HardwareAddr.String()) //获取本机MAC地址
		if mac != "" {
			return
		}
	}
	return
}

//serialnumber,pin,username create auto
