package config

import "strings"

type CommaSeparatedString []string

func (s *CommaSeparatedString) UnmarshalText(buf []byte) (e error) {
	var ss CommaSeparatedString
	for args := range strings.SplitSeq(string(buf), ",") {
		ss = append(ss, args)
	}
	*s = ss
	return
}
