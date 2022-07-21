package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// GotifyConfig -
type GotifyConfig struct {
	URL         string `yaml:"url"`
	Token       string `yaml:"token"`
	MinPriority int    `yaml:"minPriority"`
}

// LogConfig -
type LogConfig struct {
	LogLevel string `yaml:"logLevel"`
}

// Config -
type Config struct {
	Gotify GotifyConfig `yaml:"gotify"`
	Log    LogConfig    `yaml:"log"`
}

func (lc LogConfig) get() log.Level {
	lvl, err := log.ParseLevel(lc.LogLevel)
	if err != nil {
		log.Warnf("cannot parse log level '%s': %s", lc.LogLevel, err.Error())
		return log.InfoLevel
	}
	return lvl
}

func loadConfig() (c Config, err error) {
	var confDir = os.Getenv("XDG_CONFIG_HOME")
	if confDir == "" {
		confDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	confDir = filepath.Join(confDir, "gotify-notify")
	var confPath = filepath.Join(confDir, "config.yaml")

	info, err := os.Stat(confPath)
	if err != nil {
		return
	}

	if info.IsDir() {
		return c, fmt.Errorf("config file path '%s' is directory", confPath)
	}

	data, err := ioutil.ReadFile(confPath)
	if err != nil {
		return
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return
	}

	return
}
