package http

import (
	easyjson "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPut(t *testing.T) {
	expected := "200 OK"
	commandstring := "http://192.168.1.90:48082/api/v1/device/b8eb00ce-57ac-4455-9132-262d24c4068e/command/5e780f0d-fcfa-4747-902d-d7b831e609dc"
	param := "{\"brightness\":\"100\"}"
	actual, _ := Put(commandstring, param)
	assert.Equal(t, expected, string(actual), "测试Put")
}

func TestGetMessage1(t *testing.T) {
	expectedName := "brightness"
	msg := "http://192.168.1.90:48082/api/v1/device/b8eb00ce-57ac-4455-9132-262d24c4068e/command/5e780f0d-fcfa-4747-902d-d7b831e609dc"
	data, _ := GetMessage(msg)

	actualName := easyjson.Get(data, "readings", 0, "name").ToString()
	//
	assert.Equal(t, expectedName, actualName, "测试名字")
	//assert.EqualError(t, err, "asd","测试ERR" )

	// assert.Equal(t,expectedName, actualName,"测试名字")

}

func TestGetMessage2(t *testing.T) {
	//expectedName := "brightness"
	//msg := "http://192.168.1.90:48082/api/v1/device/b8eb00ce-57ac-4455-9132-262d24c4068e/command/5e780f0d-fcfa-4747-902d-d7b831e609dc"
	msg := "asdsadas"
	_, err := GetMessage(msg)

	//actualName := easyjson.Get(data, "readings",0,"name").ToString()
	//
	//assert.Equal(t,expectedName, actualName,"测试名字")
	assert.EqualError(t, err, "ErrGetFail", "测试ERR")

	// assert.Equal(t,expectedName, actualName,"测试名字")

}

func TestGetMessage3(t *testing.T) {
	//expectedName := "brightness"
	//msg := "http://192.168.1.90:48082/api/v1/device/b8eb00ce-57ac-4455-9132-262d24c4068e/command/5e780f0d-fcfa-4747-902d-d7b831e609dc"
	msg := "asdsadas"
	_, err := GetMessage(msg)

	//actualName := easyjson.Get(data, "readings",0,"name").ToString()
	//
	//assert.Equal(t,expectedName, actualName,"测试名字")
	assert.EqualError(t, err, "ErrGetFail", "测试ERR")

	// assert.Equal(t,expectedName, actualName,"测试名字")

}

func TestHttpPost(t *testing.T) {

}
