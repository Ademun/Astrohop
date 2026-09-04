package apperr

import (
	"fmt"
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
