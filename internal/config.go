package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	yaml "github.com/goccy/go-yaml"
)

const configFileName = "lakehouse-admin-tools.conf"

type Config struct {
	AccessKey    string  `yaml:"access_key"`
	SecretKey    string  `yaml:"secret_key"`
	SessionToken *string `yaml:"session_token"`

	OBS struct {
		Endpoint string
	}
}

func (c Config) ToArgs() []any {
	var (
		args []any
		v    = reflect.ValueOf(c)
	)
	for _, f := range reflect.VisibleFields(reflect.TypeFor[Config]()) {
		if !strings.HasPrefix(f.Name, "_") {
			args = append(args, f.Name)
			args = append(args, fmt.Sprintf("%#v", v.FieldByName(f.Name)))
		}
	}
	return args
}

var logger = DefaultLogger()

func getConfigFromBuffer(buf []byte) *Config {
	c := &Config{}
	if err := yaml.Unmarshal(buf, c); err != nil {
		logger.Debug("unable to load config.", logger.Args("error", err))
	} else {
		logger.Debug("config loaded.", logger.Args(c.ToArgs()...))
	}
	return c
}

func readConfigPath() []byte {
	if pwd, err := os.Getwd(); err == nil {
		if buf, err := os.ReadFile(filepath.Join(pwd, configFileName)); err == nil {
			return buf
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		if buf, err := os.ReadFile(filepath.Join(home, ".config", configFileName)); err == nil {
			return buf
		}
	}
	return nil
}

func GetConfig(path *string) *Config {
	if path != nil {
		buf, err := os.ReadFile(*path)
		if err == nil {
			return getConfigFromBuffer(buf)
		} else {
			logger.Warn("unable to load config from path.", logger.Args("path", *path, "error", err))
			return getConfigFromBuffer(nil)
		}
	}
	return getConfigFromBuffer(readConfigPath())
}
