package internal

import (
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
	"github.com/unhealme/lakehouse-admin-tools/internal/dataarts"
)

const configFileName = "lakehouse-admin-tools.conf"

type Config struct {
	AccessKey    string `yaml:"access_key"`
	SecretKey    string `yaml:"secret_key"`
	SessionToken string `yaml:"session_token"`
	DomainId     string `yaml:"domain_id"`
	Region       string

	DataArts struct {
		InstanceId string `yaml:"instance_id"`
		Agent      struct {
			Id, Name string
		}
		HetuConfig dataarts.DwConfig `yaml:"hetu_config"`
	}

	Obs struct {
		Endpoint string
	}

	Uam struct {
		Url, User, Password, Domain, Realm string
		BaseDN                             string `yaml:"base_dn"`
		GroupBase                          string `yaml:"group_base"`
	}

	Yarn struct {
		RMAddress string `yaml:"rm_address"`
	}
}

func getConfigFromBuffer(logger *Logger, buf []byte) *Config {
	var c Config
	if err := yaml.Unmarshal(buf, &c); err != nil {
		logger.Debug("unable to load config.", logger.Args("error", err))
		return nil
	}
	logger.Debug("config loaded.", logger.Args(ToArgs(c)...))
	return &c
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

func GetConfig(logger *Logger, path string) *Config {
	if path != "" {
		if buf, err := os.ReadFile(path); err == nil {
			return getConfigFromBuffer(logger, buf)
		} else {
			logger.Warn("unable to load config from path.", logger.Args("path", path, "error", err))
			return getConfigFromBuffer(logger, nil)
		}
	}
	return getConfigFromBuffer(logger, readConfigPath())
}
