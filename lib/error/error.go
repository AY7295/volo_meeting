package error

import (
	"errors"
	"fmt"
	"volo_meeting/consts"

	"github.com/spf13/viper"
	"net/http"
)

type errorType struct {
	Code consts.ErrorCode `json:"code"`
	Err  error            `json:"error,omitempty"`
	Type consts.ErrorType `json:"type"`
}

func (e *errorType) Error() string {
	return fmt.Sprintf("[ code: %d, type: %s, error: %v ]", e.Code, e.Type, e.Err)
}

func New(errCode consts.ErrorCode, err error) error {

	if err == nil {
		return nil
	}

	typeInfo, ok := consts.Code2Type[errCode]
	if !ok {
		typeInfo = consts.Code2Type[consts.UnknownError]
	}

	e := &errorType{}
	if errors.As(err, &e) {
		e.Code = errCode
		e.Type = typeInfo
		return e
	}

	e.Err = err
	e.Code = errCode
	e.Type = typeInfo
	return e
}

func Format(err error) (int, map[string]any) {
	e := &errorType{}
	if !errors.As(err, &e) {
		e = unknown(err)
	}

	result := map[string]any{
		"code": e.Code,
		"type": e.Type,
	}

	if viper.GetBool("DEBUG") {
		result["error"] = e.Err.Error()
	}

	httpStatus := http.StatusBadRequest
	if status, ok := consts.Code2HttpStatus[e.Code]; ok {
		httpStatus = status
	}

	return httpStatus, result
}

func unknown(err error) *errorType {
	return &errorType{
		Code: consts.UnknownError,
		Err:  err,
		Type: consts.Code2Type[consts.UnknownError],
	}
}
