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

func (r *BaseResponse) Response(c *gin.Context, err error, data interface{}) {
	if err != nil {
		r.error(c, err)
	} else {
		r.success(c, data)
	}
}

func (r *BaseResponse) success(c *gin.Context, data interface{}) {
	errno, errmsg := errors.Output(nil)
	resp := response{
		Errno:   errno,
		Message: errmsg,
		Data:    data,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)
	rJson, _ := json.Marshal(resp)
	c.Set("response", rJson)
}

func (r *BaseResponse) error(c *gin.Context, err error) {
	errno, errmsg := errors.Output(err)
	resp := response{
		Errno:   errno,
		Message: errmsg,
		Data:    nil,
		TraceID: tracing.TraceID(c.Request.Context()),
	}
	c.JSON(http.StatusOK, resp)
	rJson, _ := json.Marshal(resp)
	c.Set("response", rJson)
}
