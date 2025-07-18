package constants

const (
	ErrnoSuccess       = 0
	ErrnoUnknown       = 1
	ErrnoBindRequest   = 100
	ErrnoInvalidParams = 200
)

var (
	ErrMsg = map[int]string{
		ErrnoSuccess:       "success",
		ErrnoUnknown:       "unknown error",
		ErrnoBindRequest:   "bind request error",
		ErrnoInvalidParams: "invalid parameters",
	}
)
