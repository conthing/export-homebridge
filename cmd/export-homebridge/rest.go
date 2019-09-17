package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/conthing/export-homebridge/pkg/device"
	httpsender "github.com/conthing/export-homebridge/pkg/http"
	"github.com/conthing/export-homebridge/pkg/zmqinit"

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
	pincode := device.Pincode
	if pincode == "" {
		common.Log.Error("ErrPincodeNil")
		_, err := w.Write([]byte("ErrPincodeNil")) //多个homebridge的数据再组
		if err != nil {
			return
		}
	}
	var data map[string]string = map[string]string{}
	var datasend []map[string]string
	data["pincode"] = pincode
	data["QRcode"] = zmqinit.QRcode //直接调用zmqinit.go里面生成的QRcode
	datasend = append(datasend, data)
	datajson, err := json.MarshalIndent(datasend, "", " ")
	if err != nil {
		return
	}
	_, err = w.Write([]byte(datajson)) //多个homebridge的数据再组
	if err != nil {
		return
	}
}

// Respond with PINGRESPONSE to see if the service is alive
func pingHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("pong"))
}

func commandHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	var bd string
	bd = string(buf[:n])
	err := zmqinit.EventHanler(bd)
	if err != nil {
		return
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
		return
	}
}

func rebootHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()
	buf := make([]byte, 1024) // 1024为缓存大小，即每次读出的最大数据
	n, _ := r.Body.Read(buf)  // 为这次读出的数据大小
	var bd string
	bd = string(buf[:n])
	common.Log.Info("rebootHandler: ", bd)
	device.Accessaries = nil
	device.Accessarysenders = nil
	labels := []string{"Light", "Curtain"}
	for _, label := range labels {
		projectUrl := "http://localhost:52030/api/v1/project/" + label
		var projectlist, _ = httpsender.GetMessage(projectUrl)
		_ = device.Decode(projectlist, label, zmqinit.Statuspubport) //以前第3个参数是Statuspubport现在我改成直接调用zmqinit.Statuspubport
	}
}
