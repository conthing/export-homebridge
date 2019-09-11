package http //约定:package的名字和文件夹的名字一样

import ( //import中的包GoLand会根据代码自动加入，github上的包需要自己添加
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/conthing/export-homebridge/pkg/device"
	"github.com/conthing/export-homebridge/pkg/errorHandle"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	jsoniter "github.com/json-iterator/go" //虽然不知道怎么使用，但是从代码的使用情况反推出它是"github.com/json-iterator
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	//  /go"中的包go，可以进行函数的调用，比如调用json-iterator/go中的Get函数:jsoniter.Get(lightprojectlist)
)

////todo 大写开头的函数、结构体、变量、常量 加注释
//第一步:向edgex48071中注册(写)export-homebridge(http://localhost:48071/api/v1/registration)
func HttpPost(statusport string) error { //todo 注册中的ID、Created、Modified、Origin、Encryption等在哪里  虽然是edgex分配的，但是两者是如何交互的，定义的变量在哪里   ?????
	reg := models.Registration{} //定义reg为一个名字为Registration的结构体变量
	reg.Name = "RESTXMLClient"   //以下的都是给结构体中的一些变量赋值
	reg.Format = "JSON"
	////todo 为什么这几个字符串？加注释
	reg.Filter.ValueDescriptorIDs = []string{"brightness", "percent", "moving", "onoff"}
	reg.Enable = true
	reg.Destination = "REST_ENDPOINT"
	reg.Addressable = models.Addressable{Name: "EdgeXTestRESTXML", Protocol: "HTTP", HTTPMethod: "POST",
		Address: "localhost", Port: 8111, Path: "/rest"} //todo Addressable中的created、modified、origin、id、baseURL、
	//todo url在哪里  ???????????
	//// todo 无效的代码要删掉
	data, err := json.Marshal(reg) // 对变量reg进行json序列化
	if err != nil {                //有err，export-homebridge就起不起来，因为没有注册成功
		return errorHandle.ErrMarshalFail
	}
	////todo data转成string又转回来，没必要
	//var jsonstr = []byte(string(data))   //todo data本身就是[]byte类型，有点多此一举
	////todo 固定的url用const定义，或配置文件
	resp, err := http.Post("http://localhost:48071/api/v1/registration",
		"application/json", //将需要注册的数据传给resp
		bytes.NewBuffer(data))
	if err != nil { //有err打印err并返回err，可根据日志定位解决问题
		log.Println(err)
		return err
	}
	defer func() { //todo defer是一个延迟函数，在这里defer调用func()空函数，在这个函数之外出现panic、每当执行到return
		//todo 时就会执行defer，此时会关闭   加defer延迟函数的好处是可以在有错误的时候可以重新执行defer函数之外的函数
		err = resp.Body.Close() //todo resp.Body.Close()会返回error类型的err，close()会发现最基本的错误
		if err != nil {         //有错误就返回到以前重新执行程序
			return
		}
	}() //没错误就继续向下执行
	body, err := ioutil.ReadAll(resp.Body) //todo 还得好好研究一下 ?????
	if err != nil {                        //todo 为什么不打印日志  问:什么时候的err日志需要打印目的是定位解决问题，什么时候不需要  ?????
		return err
	}
	log.Println(string(body)) //打印注册的数据
	////todo 排版缩进用工具调整下
	////todo 固定的url用const
	//第二步:获取设备列表
	lightprojectUrl := "http://localhost:52030/api/v1/project/Light" //定义light的url
	lightprojectlist, err := GetMessage(lightprojectUrl)             //调用GetMessage函数，获取52030project上light的列表
	if err != nil {                                                  //有err返回到重新获取light列表
		return err
	}
	curtainprojectUrl := "http://localhost:52030/api/v1/project/Curtain" //定义curtain的url
	curtainprojectlist, err := GetMessage(curtainprojectUrl)             //调用GetMessage函数，获取52030project上curtain的列表
	if err != nil {                                                      //有err返回到重新获取curtain列表
		return err
	}

	if jsoniter.Get(lightprojectlist).Size() == 0 && jsoniter.Get(curtainprojectlist).Size() == 0 {
		return errorHandle.ErrSizeNil //如果灯光、窗帘等虚拟设备一个都没有则export-homebridge起不起来
	} //todo 为什么有Size()函数  ?????
	////todo 此处err没有判断，所有的返回值必须判断
	err = device.Decode(lightprojectlist, "Light", statusport) //todo 虽然decode函数定义返回的类型是error，可是为什么要两句代码输出的都是err
	err = device.Decode(curtainprojectlist, "Curtain", statusport)
	if err != nil { //有err返回err，否则返回空
		return err
	}
	return nil
}

////todo msg应该命名成url，命名必须自注释，驼峰命名法
//定义GetMessage函数，方便上方第二步调用
func GetMessage(msg string) (body []byte, err error) { //GetMessage函数定义
	resp, err := http.Get(msg) //http.Get函数返回两个参数:resp err
	if err != nil {            //如果有err则返回ErrGetFail
		return nil, errorHandle.ErrGetFail
	}

	defer func() { //defer延迟函数
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err = ioutil.ReadAll(resp.Body) //同上
	if err != nil {                       //有err打印err
		// handle error
		////todo 日志统一采用utils中的package
		log.Println(err)
		return nil, errorHandle.ErrReadFail
	}

	return
}

////todo commandstring应该命名成url
func Put(commandstring string, params string) (status string, err error) {

	fmt.Println("commandstring :", commandstring) //打印命令字符串
	fmt.Println("params :", params)               //打印参数

	payload := strings.NewReader(params)                       //todo 不知道为什么这样?????
	req, err := http.NewRequest("PUT", commandstring, payload) //todo 不知道为什么需要这样?????
	if err != nil {                                            //有错返回错
		return "", errorHandle.ErrRequestFail
	}
	req.Header.Add("Content-Type", "application/json") //todo 这3个header一定需要吗  可有可无的最好全部不要  必不可少的才要 保证代码的精简
	////todo 这个author是什么，加注释
	req.Header.Add("Authorization", "bmgAAGI155F6MJ3N2Tk9ruL_6XQpx-uxkkg:yKx_OYDtI3njD7-c7Y87Oov0GpI=:eyJyZXNvdXJBvcy93aF9mbG93RGF0YVNvdXJjZTEiLCJleHBpcmVzIjoxNTM2NzU1MjkwLCJjb250ZW50TUQ1IjoiIiwiY29udGVudFR5cGUiOiJhcHBsaWNhdGlvbi9qc29uIiwiaGVhZGVycyI6IiIsIm1ldGhvZCI6IlBVVCJ9")
	////todo 为什么是这个特定的时间，加注释
	req.Header.Add("Date", "Wed, 12 Sep 2018 02:10:09 GMT")

	res, err := http.DefaultClient.Do(req) //todo 这个牵涉的面很多  得深入了解  ?????
	if err != nil {                        //有err打印err
		fmt.Println(err)
		return "", errorHandle.ErrPutFail
	}
	status = res.Status //todo res里面有Status从哪里判断出来的?????
	defer func() {      //defer延迟函数
		err = res.Body.Close()
		if err != nil {
			return
		}
	}()
	return
}
