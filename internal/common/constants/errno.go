package constants

const (
	ErrnoSuccess       = 0
	ErrnoUnknown       = 404
	ErrnoBindRequest   = 403
	ErrnoInvalidParams = 401
)

var (
	ErrMsg = map[int]string{
		ErrnoSuccess:       "success",
		ErrnoUnknown:       "unknown error",
		ErrnoBindRequest:   "bind request error",
		ErrnoInvalidParams: "invalid parameters",
	}
)
