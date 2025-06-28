package response

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/peiyouyao/gorder/common/handler/errors"
	"github.com/peiyouyao/gorder/common/tracing"
)

type BaseResponse struct{}

type response struct {
	Errno   int    `json:"errno"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	TraceID string `json:"trace_id"`
}

func (base *BaseResponse) Response(c *gin.Context, err error, data interface{}) {
	if err != nil {
		base.error(c, err)
	} else {
		base.success(c, data)
	}
}

func (base *BaseResponse) success(c *gin.Context, data interface{}) {
	errno, errmsg := errors.Output(nil)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    data,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, r)
	rJson, _ := json.Marshal(r)
	c.Set("response", rJson)
}

func (base *BaseResponse) error(c *gin.Context, err error) {
	errno, errmsg := errors.Output(err)
	r := response{
		Errno:   errno,
		Message: errmsg,
		Data:    nil,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, r)
	rJson, _ := json.Marshal(r)
	c.Set("response", rJson)
}
