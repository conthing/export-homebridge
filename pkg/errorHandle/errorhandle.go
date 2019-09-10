package errorHandle

import "errors"

//todo 这些不是errorhandler，不需要单独一个package
var ErrGetFail = errors.New("ErrGetFail")
var ErrReadFail = errors.New("ErrReadFail")
var ErrRequestFail = errors.New("ErrRequestFail")
var ErrPutFail = errors.New("ErrPutFail")
var ErrContextFail = errors.New("ErrContextFail")
var ErrSocketFail = errors.New("ErrSocketFail")
var ErrConnectFail = errors.New("ErrConnectFail")
var ErrRevFail = errors.New("ErrRevFail")
var ErrUnmarshalFail = errors.New("ErrUnmarshalFail")
var ErrSendFail = errors.New("ErrSendFail")
var ErrBindFail = errors.New("ErrBindFail")
var ErrMarshalFail = errors.New("ErrMarshalFail")
var ErrWriteFail = errors.New("ErrWriteFail")
var ErrParseFail = errors.New("ErrParseFail")
var ErrMacInvalid = errors.New("ErrMacInvalid")
var ErrPincodeNil = errors.New("ErrPincodeNil")
var ErrCreateFail = errors.New("ErrCreateFail")
var ErrSizeNil = errors.New("ErrSizeNil")
