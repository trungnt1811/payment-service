package http

import (
	"github.com/gin-gonic/gin"
)

func JSON(c *gin.Context, status int, payload any, isCached ...bool) {
	var res Response
	res.Status = status
	res.Data = payload
	if len(isCached) > 0 {
		res.IsCached = isCached[0]
	}
	c.Abort()
	c.JSON(status, res)
}

func Error(c *gin.Context, status int, msg string, errors ...error) {
	errResp := GeneralError{
		Code:    status,
		Message: msg,
		Errors:  make([]string, 0),
	}
	for _, err := range errors {
		if err != nil {
			errResp.Errors = append(errResp.Errors, err.Error())
		}
	}
	c.Abort()
	c.JSON(status, errResp)
}

func Errors(c *gin.Context, status int, payload any) {
	var res ErrorResponse
	res.Status = status
	res.Errors = payload
	c.Abort()
	c.JSON(status, res)
}

func NewErrorMap(key string, err error) map[string]any {
	res := ErrorMap{
		Errors: make(map[string]any),
	}
	res.Errors[key] = err.Error()
	return res.Errors
}
