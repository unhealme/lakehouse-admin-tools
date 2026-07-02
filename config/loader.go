package config

import (
	"os"
	"path/filepath"

	yaml "github.com/goccy/go-yaml"
	"github.com/pterm/pterm"
)

const configFileName = "lakehouse-admin-tools.conf"

func getConfigFromBuffer(logger *pterm.Logger, buf []byte) *Arguments {
	var c Arguments
	if err := yaml.Unmarshal(buf, &c); err != nil {
		logger.Debug("unable to load config.", logger.Args("error", err))
		return nil
	}
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

func GetConfig(logger *pterm.Logger, path string) *Arguments {
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
