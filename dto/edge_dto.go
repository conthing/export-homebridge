package dto

//LabeledDevice 表示打标签的设备
type LabeledDevice struct {
	Name     string        `json:"name"`
	ID       string        `json:"id"`
	Commands []FlatCommand `json:"commands"`
}

// FlatCommand 原子化Command
type FlatCommand struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Value  string `json:"value"`
	Device string `json:"device"`
}
