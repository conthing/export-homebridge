package errors

import "errors"

var (
	ProjectUnfinishedErr = errors.New("ha-project 初始化未完成")
	QRCodeAssertErr      = errors.New("QRCode 断言失败")
)
