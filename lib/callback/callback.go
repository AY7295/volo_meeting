package callback

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	error2 "volo_meeting/lib/error"
)

type Finally func() (any, error)

func Final(ctx *gin.Context, finally Finally) {
	data, err := finally()
	if err != nil {
		Error(ctx, err)
		return
	}

	if data != nil {
		Success(ctx, data)
		return
	}

	response(ctx, http.StatusOK, "")
}

func Error(ctx *gin.Context, err error) {
	if err == nil {
		response(ctx, http.StatusOK, "")
		return
	}

	zap.L().Debug("callback.Error", zap.Error(err))

	httpStatus, data := error2.Format(err)
	response(ctx, httpStatus, data)
}

func Success(ctx *gin.Context, data any) {
	if data == nil {
		data = ""
	}
	response(ctx, http.StatusOK, data)
}

func response(ctx *gin.Context, httpStatus int, data any) {
	ctx.JSON(httpStatus, data)
	ctx.AbortWithStatus(httpStatus)
}
