package homebridgeconfig

type VirtualDevice struct {
	Id       string    `json:"id"`
	Name     string    `json:"name"`
	Commands []Command `json:"commands"`
}

type Command struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Value  string `json:"value"`
	Device string `json:"device"`
}
