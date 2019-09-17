package device

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestDecode(t *testing.T) {
	file, err := ioutil.ReadFile("H:/gitproject/export-homebridge/pkg/device/expected.json")
	if err != nil {
		fmt.Println(err)
	}
	File := string(file)
	assert.Equal(t, AutoGenerated{}, File, "Equal")
}
