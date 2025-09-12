package common

import (
	"runtime/debug"
)

func HandlePanic() {
	if r := recover(); r != nil {
		Logger.Sugar().Errorf("catch panic: %v \n stack: %s", r, string(debug.Stack()))
	}
}
