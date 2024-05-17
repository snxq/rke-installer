package rkeinstaller

import (
	"fmt"
)

func ErrHandle(msg string, err error) {
	if err != nil {
		panic(fmt.Sprintf("%s: %v", msg, err))
	}
}
