package internal

import (
	"fmt"
	"os"
	"reflect"
)

func GetEnv(k, def string) string {
	if v, e := os.LookupEnv(k); e {
		return v
	}
	return def
}

func StructToArgs(a any) (args []any) {
	va := reflect.ValueOf(a)
	if va.Kind() == reflect.Pointer {
		va = va.Elem()
	}
	if va.Kind() != reflect.Struct {
		return
	}

	for f, v := range va.Fields() {
		if f.Tag.Get("arg") == "-" {
			continue
		}
		args = append(args, f.Name, fmt.Sprintf("%+v", v))
	}
	return
}
