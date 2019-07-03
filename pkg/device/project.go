package device

type Project struct {
	Id       string
	Name     string
	Commands []Command
}

type Command struct {
	Id     string
	Name   string
	Value  string
	Device string
}
