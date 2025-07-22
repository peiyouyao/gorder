package util

import (
	"strings"
)

type ArgFormatter interface {
	FormatArg() (string, error)
}

func FormatArgs(args []any) string {
	var item []string
	for _, arg := range args {
		item = append(item, FormatArg(arg))
	}
	return strings.Join(item, "||")
}

func FormatArg(arg any) string {
	var (
		str string
		err error
	)
	defer func() {
		if err != nil {
			str = "unsupported type in formatMySQLArg||err=" + err.Error()
		}
	}()
	switch v := arg.(type) {
	default:
		s, _ := MarshalString(v)
		return s
	case ArgFormatter:
		str, err = v.FormatArg()
	}
	return str
}
