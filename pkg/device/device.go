package device

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/conthing/utils/common"
)

//AutoGenerated means the construction of config.json，前4个定义的结构体是生成config.json文件用的
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

//omitempty的注释:1、加上omitempty如果dimmerable为nil，则生成的dimmerable不会显示""；2、不加omitempty则如果dimmerable为nil，生成的dimmerable会显示""；
type Accessary struct {
	Service    string `json:"service"`
	Name       string `json:"name"`
	ProxyID    string `json:"proxy_id"`
	Accessory  string `json:"accessory"`
	Dimmerable string `json:"dimmerable,omitempty"`
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

//定义Decode函数，这个函数主要做:1、zigbee设备的name相同时则对应的虚拟设备的alias就会相同，这个函数就保证了homekit上的虚拟设备
//的alias没有重名的，相同alias的会在其后加((1)、(2)、(3)....等等表示)，就是下方的的index := 1；2、homekit上的虚拟设备如果是调
//光灯在控制的时候就显示百分百，如果是开关灯在控制的时候就显示开和关；3、指定homebridge的config.json文件的生成路径；
func Decode(jsonStr []byte, label string, statusport string) error {
	var projects []Project
	err := json.Unmarshal(jsonStr, &projects)
	if err != nil {
		return err
	}
	index := 1
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
			common.Log.Warn("不存相应设备")
		}
		//这个for循环用在web上的zigbee设备的name如果相同则对应的虚拟设备灯光的alias也相同，这是在后面加上(1、2、3....)以示区分
		for _, projectcommand := range project.Commands {
			if projectcommand.Name == "alias" {
				if Accessaries != nil {
					for _, access := range Accessaries {
						if access.Name != projectcommand.Value {
							accessary.Name = projectcommand.Value
						} else {
							accessary.Name = fmt.Sprintf("%s(%d)", projectcommand.Value, index)
							common.Log.Info("accessary.Name: ", accessary.Name)
							index++
							break
						}
					}
				} else {
					accessary.Name = projectcommand.Value
				}
			} else if projectcommand.Name == "dimmerable" {
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
		return err
	}
	b, err := json.MarshalIndent(configdata, "", " ")
	if err != nil {
		return err
	}
	//生成config.json文件的路径
	err = ioutil.WriteFile("/root/.homebridge/config.json", b, os.ModeAppend) //create config.json
	if err != nil {
		return err
	}
	return nil
}

//creatConfigData()函数主要是:解释config.json文件中的各个参数，其中一些是常量赋值，一些是生成的
func createConfigData(accessaries []Accessary, statusport string) (configdata AutoGenerated, err error) {
	macAddr := common.GetMacAddrByName("eth0") //获取本机有线网卡的mac地址，注:一般来说，ubuntu系统中有线网卡的name是eth0，无线网卡的name是wlan0,但具体有线网卡和无线网卡的名字是什么还要看实际的板子
	mac := strings.ToUpper(macAddr)            //将获取到的mac地址的英文字母由小写变成大写
	pinstring := strings.Split(mac, ":")       //分割mac地址中的:和在一起的两个16进制的两位字符，举个例子:A1:22:3C:34:45:AB，在这里就是把A1和:完全分开，其它依此类推
	pinnum := make([]int, 6)
	if len(pinstring) == 6 { //由于上述已经分开了mac地址的:和由英文字母或数字组成的连在一起的两个字符，而每个字符都是16进制，故总共有6个字节
		for i, pin := range pinstring {
			n, err := strconv.ParseUint(pin, 16, 8)
			if err != nil {
				return configdata, err
			} else {
				pinnum[i] = int(n)
			}
		}
	} else {
		return
	}
	//获取本机的mac地址要反过来写的原因是如果正着写则第2次就homebridge页面就生成不了二维码和pin码了
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
