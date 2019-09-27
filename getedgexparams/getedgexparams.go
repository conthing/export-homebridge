package getedgexparams

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/conthing/export-homebridge/homebridgeconfig"
	"github.com/conthing/utils/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

const (
	LIGHTPROJECTURL   = "http://localhost:52030/api/v1/project/Light"
	CURTAINPROJECTURL = "http://localhost:52030/api/v1/project/Curtain"
	REGISTRATIONURL   = "http://localhost:48071/api/v1/registration"
)

//第一步:向edgex48071中注册(写)export-homebridge(http://localhost:48071/api/v1/registration)
func HttpPost(statusport string) error {
	reg := models.Registration{}
	reg.Name = "RESTXMLClient"
	reg.Format = "JSON"
	//起筛选作用，目前灯光、窗帘等虚拟设备只有亮度值、开关状态、行程值、行程状态等4个变量
	reg.Filter.ValueDescriptorIDs = []string{"brightness", "percent", "moving", "onoff"}
	reg.Enable = true
	reg.Destination = "REST_ENDPOINT"
	reg.Addressable = models.Addressable{Name: "EdgeXTestRESTXML", Protocol: "HTTP", HTTPMethod: "POST",
		Address: "localhost", Port: 8111, Path: "/rest"}
	data, err := json.Marshal(reg)
	if err != nil {
		common.Log.Errorf("HttpPost(statusport string) data json.Marshal failed: %v", err)
	}
	resp, err := http.Post(REGISTRATIONURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		common.Log.Errorf("HttpPost(statusport string) http.Post failed: %v", err)
	}
	defer func() { //defer是一个延迟函数，在这里defer调用func()空函数，在这个函数之外出现panic、每当执行到return
		//时就会执行defer，此时会关闭   加defer延迟函数的好处是可以在有错误的时候可以重新执行defer函数之外的函数
		err = resp.Body.Close() //resp.Body.Close()会返回error类型的err，close()会发现最基本的错误
		if err != nil {
			common.Log.Errorf("HttpPost(statusport string) resp.Body.Close() failed: %v", err)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		common.Log.Errorf("HttpPost(statusport string) ioutil.ReadAll(resp.Body) failed: %v", err)
	}
	common.Log.Info(string(body)) //打印注册的数据

	// 第二步:获取设备列表
	lightdevicelist, err := GetMessage(LIGHTPROJECTURL)
	if err != nil {
		common.Log.Errorf("GetMessage(LIGHTPROJECTURL) failed: %v", err)
	}
	curtaindevicelist, err := GetMessage(CURTAINPROJECTURL)
	if err != nil {
		common.Log.Errorf("GetMessage(CURTAINPROJECTURL) failed: %v", err)
	}
	err = homebridgeconfig.GenerateHomebridgeConfig(lightdevicelist, curtaindevicelist, statusport)
	if err != nil {
		common.Log.Errorf("homebridgeconfig.GenerateHomebridgeConfig(lightdevicelist, curtaindevicelist, statusport) failed: %v", err)
	}
	return nil
}

func GetMessage(url string) (body []byte, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
		common.Log.Errorf(" GetMessage(url string) http.Get(url) failed: %v", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			common.Log.Errorf(" GetMessage(url string) resp.Body.Close() failed: %v", err)
		}
	}()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		common.Log.Errorf(" GetMessage(url string) ioutil.ReadAll(resp.Body) failed: %v", err)
		return nil, err
	}
	return
}

////todo commandstring应该命名成url  ?????? commanfstring是输入参数 是什么  为什么要命名为url
func Put(commandstring string, params string) (status string, err error) {
	common.Log.Info("commandstring :", commandstring)
	common.Log.Info("params :", params)
	payload := strings.NewReader(params)
	req, err := http.NewRequest("PUT", commandstring, payload)
	if err != nil {
		common.Log.Errorf("Put(commandstring string, params string) http.NewRequest(PUT, commandstring, payload) failed: %v", err)
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	////todo 这个author是什么，加注释
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	////todo 为什么是这个特定的时间，加注释
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		common.Log.Errorf("Put(commandstring string, params string) http.DefaultClient.Do(req) failed: %v", err)
		return "", err
	}
	status = res.Status
	defer func() {
		err = res.Body.Close()
		if err != nil {
			common.Log.Errorf("Put(commandstring string, params string) res.Body.Close() failed: %v", err)
		}
	}()
	return
}
