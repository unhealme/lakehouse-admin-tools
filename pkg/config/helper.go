package config

import (
	"strings"

	"github.com/goccy/go-yaml"
)

type CommaSeparatedString []string

func (s *CommaSeparatedString) UnmarshalText(buf []byte) (e error) {
	var ss CommaSeparatedString
	for args := range strings.SplitSeq(string(buf), ",") {
		ss = append(ss, args)
	}
	*s = ss
	return
}

func (s *CommaSeparatedString) UnmarshalYAML(buf []byte) error {
	var ss []string
	if err := yaml.Unmarshal(buf, &ss); err != nil {
		return err
	}
	*s = CommaSeparatedString(ss)
	return nil
}
