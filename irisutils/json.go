package irisutils

import (
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

func ReadJSON(ctx iris.Context, reqData interface{}) bool {
	if err := ctx.ReadJSON(reqData); err != nil {
		golog.Errorf("Failed to parse input sent to %s by %s: %v", ctx.Path(), ctx.RemoteAddr(), err)
		ctx.StopWithStatus(iris.StatusBadRequest)
		return false
	}

	return true
}
