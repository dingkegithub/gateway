package backend

import (
	"errors"
	"fmt"
)

var (
	HttpNot200Err = errors.New(fmt.Sprintf("%s", "Not success status 200"))
)
