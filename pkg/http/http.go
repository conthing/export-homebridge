package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/conthing/export-homebridge/pkg/device"
)

func HttpPost() {

	var registration map[string]interface{}
	var addressable map[string]interface{}
	registration = make(map[string]interface{})
	addressable = make(map[string]interface{})
	addressable["name"] = "EdgeXTestRESTXML"
	addressable["protocol"] = "HTTP"
	addressable["method"] = "POST"
	addressable["address"] = "localhost"
	addressable["port"] = 8111
	addressable["path"] = "/rest"
	registration["name"] = "RESTXMLClient"
	registration["format"] = "JSON"
	registration["enable"] = true
	registration["destination"] = "REST_ENDPOINT"
	registration["addressable"] = addressable
	data, _ := json.MarshalIndent(registration, "", " ")

	//str := "{\"origin\":1471806386919,\"name\":\"RESTXMLClient\",\"addressable\":{\"origin\":1471806386919,\"name\":\"EdgeXTestRESTXML\",\"protocol\":\"HTTP\",\"method\": \"POST\",\"address\":\"localhost\",\"port\":8111,\"path\":\"/rest\"},\"format\":\"JSON\",\"enable\":true,\"destination\":\"REST_ENDPOINT\"}"

	var jsonstr = []byte(string(data))
	resp, err := http.Post("http://localhost:48071/api/v1/registration",
		"application/json",
		bytes.NewBuffer(jsonstr))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))

	devicelisturl := "http://localhost:48081/api/v1/device"
	var devicelist = GetMessage(devicelisturl)
	device.Decode([]byte(devicelist))
	fmt.Println(devicelist)
}

func GetMessage(msg string) string {
	resp, err := http.Get(msg)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		fmt.Println(err)
		return ""
	}

	result := string(body)

	fmt.Println(string(body))

	return result
}

func Put(commandstring string, params string) {

	payload := strings.NewReader(params)
	fmt.Println("command", commandstring)
	fmt.Println("command", params)
	req, _ := http.NewRequest("PUT", commandstring, payload)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
}
