package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/conthing/export-homebridge/pkg/device"
	"github.com/conthing/export-homebridge/pkg/errorHandle"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	jsoniter "github.com/json-iterator/go"
)

//todo 大写开头的函数、结构体、变量、常量 加注释
func HttpPost(statusport string) error {

	reg := models.Registration{}
	reg.Name = "RESTXMLClient"
	reg.Format = "JSON"
	//todo 为什么这几个字符串？加注释
	reg.Filter.ValueDescriptorIDs = []string{"brightness", "percent", "moving", "onoff"}
	reg.Enable = true
	reg.Destination = "REST_ENDPOINT"
	reg.Addressable = models.Addressable{Name: "EdgeXTestRESTXML", Protocol: "HTTP", HTTPMethod: "POST",
		Address: "localhost", Port: 8111, Path: "/rest"}

	// todo 无效的代码要删掉
	// var registration map[string]interface{}
	// var addressable map[string]interface{}
	// registration = make(map[string]interface{})
	// addressable = make(map[string]interface{})
	// addressable["name"] = "EdgeXTestRESTXML"
	// addressable["protocol"] = "HTTP"
	// addressable["method"] = "POST"
	// addressable["address"] = "localhost"
	// addressable["port"] = 8111
	// addressable["path"] = "/rest"
	// registration["name"] = "RESTXMLClient"
	// registration["format"] = "JSON"
	// registration["filter"] = filter
	// registration["enable"] = true
	// registration["destination"] = "REST_ENDPOINT"
	// registration["addressable"] = addressable
	data, err := json.Marshal(reg)
	if err != nil {
		return errorHandle.ErrMarshalFail
	}

	//str := "{\"origin\":1471806386919,\"name\":\"RESTXMLClient\",\"addressable\":{\"origin\":1471806386919,\"name\":\"EdgeXTestRESTXML\",\"protocol\":\"HTTP\",\"method\": \"POST\",\"address\":\"localhost\",\"port\":8111,\"path\":\"/rest\"},\"format\":\"JSON\",\"enable\":true,\"destination\":\"REST_ENDPOINT\"}"

	// todo data转成string又转回来，没必要
	var jsonstr = []byte(string(data))
	// todo 固定的url用const定义，或配置文件
	resp, err := http.Post("http://localhost:48071/api/v1/registration",
		"application/json",
		bytes.NewBuffer(jsonstr))
	if err != nil {
		log.Println(err)
		return err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		return err
	}

	log.Println(string(body))

	//todo 排版缩进用工具调整下
	//todo 固定的url用const
	lightprojectUrl := "http://localhost:52030/api/v1/project/Light"
	lightprojectlist, err := GetMessage(lightprojectUrl)
	if err != nil {
		return err
	}
	curtainprojectUrl := "http://localhost:52030/api/v1/project/Curtain"
	curtainprojectlist, err := GetMessage(curtainprojectUrl)
	if err != nil {
		return err
	}

	if jsoniter.Get(lightprojectlist).Size() == 0 && jsoniter.Get(curtainprojectlist).Size() == 0 {
		return errorHandle.ErrSizeNil
	}
	//todo 此处err没有判断，所有的返回值必须判断
	err = device.Decode(lightprojectlist, "Light", statusport)
	err = device.Decode(curtainprojectlist, "Curtain", statusport)
	if err != nil {
		return err
	}
	return nil
}

//todo msg应该命名成url，命名必须自注释，驼峰命名法
func GetMessage(msg string) (body []byte, err error) {
	resp, err := http.Get(msg)
	if err != nil {
		return nil, errorHandle.ErrGetFail
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		// todo 日志统一采用utils中的package
		log.Println(err)
		return nil, errorHandle.ErrReadFail
	}

	return
}

//todo commandstring应该命名成url
func Put(commandstring string, params string) (status string, err error) {

	fmt.Println("commandstring :", commandstring)
	fmt.Println("params :", params)

	payload := strings.NewReader(params)
	req, err := http.NewRequest("PUT", commandstring, payload)
	if err != nil {
		return "", errorHandle.ErrRequestFail
	}

	req.Header.Add("Content-Type", "application/json")
	//todo 这个author是什么，加注释
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	//todo 为什么是这个特定的时间，加注释
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", errorHandle.ErrPutFail
	}
	status = res.Status
	defer func() {
		err = res.Body.Close()
		if err != nil {
			return
		}
	}()
	//result, _ = ioutil.ReadAll(res.Body)
	return
}
