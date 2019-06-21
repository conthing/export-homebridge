package device

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func TestDecode(t *testing.T) {

	expected := "{\n \"description\": \"This is an inSona plugin configuration file\",\n \"bridge\": {\n  \"serialNumber\": \"63.26.55.00\",\n  \"pin\": \"095-92-263\",\n  \"port\": 51826,\n  \"name\": \"homebridge-0\",\n  \"model\": \"homebridge-inSona\",\n  \"manufacturer\": \"inSona\",\n  \"username\": \"AF:52:A3:1A:FF:00\",\n  \"repport\": \"tcp://127.0.0.1:9999\"\n },\n \"platforms\": [\n  {\n   \"accessories\": [\n    {\n     \"service\": \"WindowCovering\",\n     \"name\": \"5cf8bccb0d22393226079bd7-Curtain-0\",\n     \"proxy_id\": \"5e1f737b-e324-4de0-b1ad-0c13cfa9cf99\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5cfdacb40d223944d0728332-Light-0\",\n     \"proxy_id\": \"ede2fbcc-c36a-4e38-a0c7-10842a13a8eb\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5cfdacb40d223944d0728332-Light-1\",\n     \"proxy_id\": \"135bdad9-f7d6-4cb9-8dd3-1865468c6794\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e9540d223905eaf45e00-Light-0\",\n     \"proxy_id\": \"78d5c4f1-dfe2-4d3b-b694-6003c8418b0f\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e96c0d223905eaf45e01-Light-0\",\n     \"proxy_id\": \"7e2b3001-d866-42ce-9af3-0890b6b87c11\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e9c10d223905eaf45e05-Light-0\",\n     \"proxy_id\": \"04935b4b-0347-43d6-bdcc-358114cf1991\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e9c10d223905eaf45e05-Light-1\",\n     \"proxy_id\": \"da6ed8e6-70d6-4573-98e9-7e1ca5c51135\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e97a0d223905eaf45e02-Light-0\",\n     \"proxy_id\": \"e40bc3a6-460e-4cc6-8103-605ce88b36da\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06e97a0d223905eaf45e02-Light-1\",\n     \"proxy_id\": \"90708333-c003-4d39-bee5-9c4103c36716\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06ea070d223905eaf45e06-Light-0\",\n     \"proxy_id\": \"3ac4dbc5-a2f7-4854-9129-15c3d48b277b\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"Lightbulb\",\n     \"name\": \"5d06ea070d223905eaf45e06-Light-1\",\n     \"proxy_id\": \"42528165-4836-4e7f-8acb-f367088feb9c\",\n     \"accessory\": \"Control4\"\n    },\n    {\n     \"service\": \"WindowCovering\",\n     \"name\": \"5d0837900d2239594f0b0ec4-Curtain-0\",\n     \"proxy_id\": \"699c07e1-bbda-400b-baa3-7097c35fbaed\",\n     \"accessory\": \"Control4\"\n    }\n   ],\n   \"name\": \"Control4\",\n   \"configPath\": \"/root/.homebridge/config.json\",\n   \"platform\": \"Control4\"\n  }\n ]\n}"
	msg := "http://192.168.1.90:48081/api/v1/device"
	resp, err := http.Get(msg)
	if err != nil {
		log.Println(err)
		return
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			return
		}
	}()
	devicelist, err := ioutil.ReadAll(resp.Body)
	actual, _ := Decode([]byte(devicelist))

	assert.Equal(t, expected, string(actual), "测试Device")

}
