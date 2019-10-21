package common

import (
	"strings"
)

type StringSliceFlag []string

func (ssf *StringSliceFlag) String() string {
	return strings.Join(*ssf, ", ")
}

func (ssf *StringSliceFlag) Set(value string) error {
	*ssf = append(*ssf, value)
	return nil
}

func (ssf *StringSliceFlag) Type() string {
	return "String"
}
