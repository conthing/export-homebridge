package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/conthing/export-homebridge/getedgexparams"
	"github.com/conthing/export-homebridge/homebridgeconfig"
	"github.com/conthing/export-homebridge/zmqreceivesendhandler"

	"github.com/conthing/utils/common"
	"github.com/gorilla/mux"
)

func LoadRestRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/rest", commandHandler).Methods(http.MethodGet, http.MethodPost)
	r.HandleFunc("/api/v1/version", versionHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/reboot", rebootHandler).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/homebridge/qrcode", qrcodeHandler).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/ping", pingHandler).Methods(http.MethodGet)
	return r
}

func qrcodeHandler(w http.ResponseWriter, r *http.Request) {
	common.Log.Info("Request", r)
	w.Header().Set("Content-Type", "text/plain")
	pincode := homebridgeconfig.Pincode
	if pincode == "" {
		common.Log.Error("ErrPincodeNil")
		_, err := w.Write([]byte("ErrPincodeNil")) //多个homebridge的数据再组
		if err != nil {
			common.Log.Errorf("qrcodeHandler(w http.ResponseWriter, r *http.Request) w.Write([]byte(ErrPincodeNil) failed: %v", err)
		}
	}
	var data map[string]string = map[string]string{}
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = zmqreceivesendhandler.QRcode //直接调用zmqinit.go里面生成的QRcode
	datasend = append(datasend, data)
	datajson, err := json.MarshalIndent(datasend, "", " ")
	if err != nil {
		common.Log.Errorf("qrcodeHandler(w http.ResponseWriter, r *http.Request) datajson json.MarshalIndent failed: %v", err)
	}
	_, err = w.Write([]byte(datajson)) //多个homebridge的数据再组
	if err != nil {
		common.Log.Errorf("qrcodeHandler(w http.ResponseWriter, r *http.Request) w.Write([]byte(datajson)) failed: %v", err)
	}
}

// Respond with PINGRESPONSE to see if the service is alive
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, err := w.Write([]byte("pong"))
	if err != nil {
		common.Log.Error(err)
	}
}

func commandHandler(w http.ResponseWriter, r *http.Request) { //edgex传递参数
	defer func() {
		err := r.Body.Close()
		if err != nil {
			common.Log.Errorf("commandHandler(w http.ResponseWriter, r *http.Request) r.Body.Close() failed: %v", err)
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	var bd string
	bd = string(buf[:n])
	//common.Log.Info("commandHandler ", bd)
	err := zmqreceivesendhandler.EventHanler(bd)
	if err != nil {
		common.Log.Errorf("commandHandler(w http.ResponseWriter, r *http.Request) zmqreceivesendhandler.EventHanler(bd) failed: %v", err)
	}
	//4.对收到的event进行处理，然后发给js   status
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	version := common.Version
	currentTime := common.BuildTime
	common.Log.Info("version", version)
	common.Log.Info("Time", currentTime)
	datastring := strings.Join([]string{version, currentTime}, " ")
	_, err := w.Write([]byte(datastring)) //多个homebridge的数据再组
	if err != nil {
		common.Log.Errorf("versionHandler(w http.ResponseWriter, r *http.Request) w.Write([]byte(datastring)) failed: %v", err)
	}
}
func rebootHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			common.Log.Errorf("rebootHandler(w http.ResponseWriter, r *http.Request) r.Body.Close() failed: %v", err)
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	var bd string
	bd = string(buf[:n])
	common.Log.Info("rebootHandler: ", bd)
	homebridgeconfig.Accessaries = nil
	homebridgeconfig.Accessarysenders = nil
	var lightdevicelist, _ = getedgexparams.GetMessage(getedgexparams.LIGHTPROJECTURL)
	var curtaindevicelist, _ = getedgexparams.GetMessage(getedgexparams.CURTAINPROJECTURL)
	var hvacdevicelist, _ = getedgexparams.GetMessage(getedgexparams.HVACPROJECTURL)
	_ = homebridgeconfig.GenerateHomebridgeConfig(lightdevicelist, curtaindevicelist, hvacdevicelist, zmqreceivesendhandler.Statuspubport)
}
