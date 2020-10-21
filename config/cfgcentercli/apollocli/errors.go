package apollocli

import (
	"errors"
	"fmt"
)

var (
	ExistErr        = errors.New("file exist exception")
	InvalidParamErr = errors.New("request param invalid")
	HttpNot200Err   = errors.New(fmt.Sprintf("%s", "Not success status 200"))
)
