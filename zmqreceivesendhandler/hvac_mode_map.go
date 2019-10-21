package zmqreceivesendhandler

var EdgexToHomebridgeHvacModeMap = map[string]string{
	"AC":     "COOL",
	"HEATER": "HEAT",
	"VENT":   "AUTO",
	"DEHUMI": "AUTO",
}
