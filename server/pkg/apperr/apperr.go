package apperr

import (
	"astrohop/pkg/logger"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

type ErrCode int

type Err struct {
	Code int
	Msg  string
	Err  error
}

func New(code int, msg string, err error) *Err {
	return &Err{Code: code, Msg: msg, Err: err}
}

func (e Err) Error() string {
	return fmt.Sprintf("Code %d: %s. Err: %s", e.Code, e.Msg, e.Err.Error())
}

func HandleHttp(c *gin.Context, err error) {
	if err != nil {
		logger.L().Error(err.Error())
	}
	if e, ok := errors.AsType[*Err](err); ok {
		c.AbortWithStatusJSON(e.Code, gin.H{"error": e.Msg})
	}
}
