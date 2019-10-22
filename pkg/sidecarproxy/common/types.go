package common

import (
	"strings"
)

type StringArrayFlag []string

func (ssf *StringArrayFlag) String() string {
	return strings.Join(*ssf, ", ")
}

func (ssf *StringArrayFlag) Set(value string) error {
	*ssf = append(*ssf, value)
	return nil
}

func (ssf *StringArrayFlag) Type() string {
	return "String"
}
