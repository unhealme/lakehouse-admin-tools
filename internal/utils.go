package internal

import (
	"fmt"
	"os"
	"reflect"
	"strings"
)

func GetEnv(k, def string) string {
	if v, e := os.LookupEnv(k); e {
		return v
	}
	return def
}

func ToArgs(a any) []any {
	var (
		args []any
		v    = reflect.ValueOf(a)
	)
	for _, f := range reflect.VisibleFields(reflect.TypeOf(a)) {
		if !strings.HasPrefix(f.Name, "_") {
			args = append(args, f.Name)
			args = append(args, fmt.Sprintf("%#v", v.FieldByName(f.Name)))
		}
	}
	return args
}
