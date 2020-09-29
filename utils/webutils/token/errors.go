package token

import "fmt"

var (
	ParseError       = fmt.Errorf("parse token error")
	FormatError = fmt.Errorf("token format error")
	ValidError = fmt.Errorf("token invalid error")
)
