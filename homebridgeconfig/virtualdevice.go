package homebridgeconfig

//注:这个VirtualDevice统一是指诸如light keypad curtain hvac等的虚拟设备
type VirtualDevice struct {
	Id       string    `json:"id"`       //注:这个Id是device的id，可称为deviceid
	Name     string    `json:"name"`     //注:这个Name是device的alias(别名)
	Commands []Command `json:"commands"` //注:Commands组成了core-command的控制方式(http://localhost:48082/api/v1/device/deviceid/command/commandid    外加传入的body即可)，为了实现对虚拟设备的控制
}

type Command struct {
	ID     string `json:"id"`     //注:这个Id是command的id，可称为commandid
	Name   string `json:"name"`   //注:这个Name是虚拟设备一些特征名字，如onoff brightness percent moving mode fanlevel ttarget等
	Value  string `json:"value"`  //注:这个Value是虚拟设备一些特征名字对应的数值，如true false 0-100之间的任意整数 high middle low heat cool dehumi vent等
	Device string `json:"device"` //注:这个Device也是指device的id
}

//以下定义的所有结构体只方便HVAC使用(edgex的onoff+mode=homebridge的mode)
type EdgexCommandDevice struct {
	Name     string         `json:"name"`
	ID       string         `json:"id"`
	Commands []EdgexCommand `json:"commands"`
}

type EdgexCommand struct {
	Name string   `json:"name"`
	GET  EdgexGET `json:"get"`
}

type EdgexGET struct {
	URL string `json:"url"`
}
