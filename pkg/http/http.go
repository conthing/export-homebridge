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

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	jsoniter "github.com/json-iterator/go"
)

////todo 大写开头的函数、结构体、变量、常量 加注释
//第一步:向edgex48071中注册(写)export-homebridge(http://localhost:48071/api/v1/registration)
func HttpPost(statusport string) error {
	//todo 注册中的ID、Created、Modified、Origin、Encryption等在哪里  虽然是edgex分配的，但是两者是如何交互的，定义的变量在哪里   ?????
	reg := models.Registration{}
	reg.Name = "RESTXMLClient"
	reg.Format = "JSON"
	////todo 为什么这几个字符串？加注释
	reg.Filter.ValueDescriptorIDs = []string{"brightness", "percent", "moving", "onoff"}
	reg.Enable = true
	reg.Destination = "REST_ENDPOINT"
	reg.Addressable = models.Addressable{Name: "EdgeXTestRESTXML", Protocol: "HTTP", HTTPMethod: "POST",
		Address: "localhost", Port: 8111, Path: "/rest"}
	//todo Addressable中的created、modified、origin、id、baseURL、url在哪里  ???????????
	//// todo 无效的代码要删掉
	data, err := json.Marshal(reg)
	if err != nil {
		return err
	}
	////todo data转成string又转回来，没必要
	////todo 固定的url用const定义，或配置文件
	resp, err := http.Post("http://localhost:48071/api/v1/registration",
		"application/json",
		bytes.NewBuffer(data))
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() { //defer是一个延迟函数，在这里defer调用func()空函数，在这个函数之外出现panic、每当执行到return
		//时就会执行defer，此时会关闭   加defer延迟函数的好处是可以在有错误的时候可以重新执行defer函数之外的函数
		err = resp.Body.Close() //resp.Body.Close()会返回error类型的err，close()会发现最基本的错误
		if err != nil {
			return
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(string(body)) //打印注册的数据
	////todo 排版缩进用工具调整下
	////todo 固定的url用const
	//第二步:获取设备列表
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
	//如果灯光、窗帘等虚拟设备一个都没有则export-homebridge起不起来
	if jsoniter.Get(lightprojectlist).Size() == 0 && jsoniter.Get(curtainprojectlist).Size() == 0 {
		return err
	}
	////todo 此处err没有判断，所有的返回值必须判断
	err = device.Decode(lightprojectlist, "Light", statusport) //todo 虽然decode函数定义返回的类型是error，可是为什么要两句代码输出的都是err
	err = device.Decode(curtainprojectlist, "Curtain", statusport)
	if err != nil {
		return err
	}
	return nil
}

////todo msg应该命名成url，命名必须自注释，驼峰命名法
//定义GetMessage函数，方便上方第二步调用
func GetMessage(msg string) (body []byte, err error) {
	resp, err := http.Get(msg)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		////todo 日志统一采用utils中的package
		log.Println(err)
		return nil, err
	}
	return
}

////todo commandstring应该命名成url
func Put(commandstring string, params string) (status string, err error) {

	fmt.Println("commandstring :", commandstring)
	fmt.Println("params :", params)

	payload := strings.NewReader(params)
	req, err := http.NewRequest("PUT", commandstring, payload)
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	////todo 这个author是什么，加注释
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	////todo 为什么是这个特定的时间，加注释
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	status = res.Status
	defer func() {
		err = res.Body.Close()
		if err != nil {
			return
		}
	}()
	return
}
