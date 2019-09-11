package device

//定义Project结构体
type Project struct {
	Id       string
	Name     string
	Commands []Command
}

//定义上方Project结构体中的Command为另一个结构体
type Command struct {
	Id     string
	Name   string
	Value  string
	Device string
}
