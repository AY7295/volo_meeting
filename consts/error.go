package consts

type ErrorCode uint8

const (
	UnknownError ErrorCode = iota
	SeverError
	SqlError
	CacheError
	Forbidden
	AuthError
	PermissionDenied
	NotFound
	ParamError
)

type ErrorType string

var Code2Type = map[ErrorCode]ErrorType{
	UnknownError:     "Unknown Error",
	SeverError:       "Server Error",
	SqlError:         "Sql Error",
	AuthError:        "Auth Error",
	Forbidden:        "Forbidden",
	PermissionDenied: "Permission Denied",
	NotFound:         "Not Found",
	ParamError:       "Param Error",
}

var Code2HttpStatus = map[ErrorCode]int{
	UnknownError: 400,
	SeverError:   500,
	SqlError:     500,
	AuthError:    401,
	Forbidden:    403,
}
