package apperr

import "fmt"

type Kind string

const (
	KindNotFound     Kind = "not_found"
	KindValidation   Kind = "validation"
	KindForbidden    Kind = "forbidden"
	KindUnauthorized Kind = "unauthorized"
	KindConflict     Kind = "conflict"
	KindInternal     Kind = "internal"
)

type Err struct {
	Kind Kind
	Code string
	Msg  string
	Err  error
}

func New(kind Kind, code, msg string, err error) *Err {
	return &Err{Kind: kind, Code: code, Msg: msg, Err: err}
}

func (e *Err) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Msg)
	}
	return fmt.Sprintf("%s: %s: %s", e.Code, e.Msg, e.Err)
}

func (e *Err) Unwrap() error {
	return e.Err
}

func NotFound(code, msg string, err error) *Err   { return New(KindNotFound, code, msg, err) }
func Forbidden(code, msg string, err error) *Err  { return New(KindForbidden, code, msg, err) }
func Validation(code, msg string, err error) *Err { return New(KindValidation, code, msg, err) }
func Internal(code, msg string, err error) *Err   { return New(KindInternal, code, msg, err) }
