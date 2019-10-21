package dto

// FlatCommand 原子化Command
type EdgexCommandDevice struct {
	Name     string         `json:"name"`
	ID       string         `json:"id"`
	Commands []EdgexCommand `json:"command'`
}

type EdgexCommand struct {
	Name string   `json:"name"`
	GET  EdgexGET `json:"get"`
}

type EdgexGET struct {
	URL string `json:"url"`
}

//"id": "e63b578f-20d6-4b4f-a792-da4c1f6aac76",
//"name": "mode",
//"get": {
//"path": "/api/v1/device/{deviceId}/mode",

//"url": "http://localhost:48082/api/v1/device/98f04fb8-2476-480e-947a-f363e654cf00/command/e63b578f-20d6-4b4f-a792-da4c1f6aac76"
//},
