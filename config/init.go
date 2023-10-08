package config

import (
	"embed"
	"github.com/spf13/viper"
	"volo_meeting/consts"
)

type Option func(*conf)

type conf struct {
	configFileType string
	configFilename string
}

// c: immutable default config
var c = conf{
	configFileType: consts.DefaultConfigFileType,
	configFilename: consts.DefaultConfigFileName,
}

func apply(opts ...Option) *conf {
	newConf := c
	for _, opt := range opts {
		opt(&newConf)
	}
	return &newConf
}

func WithFileType(fileType string) Option {
	return func(c *conf) {
		c.configFileType = fileType
	}
}

func WithFileName(filename string) Option {
	return func(c *conf) {
		c.configFilename = filename
	}
}

//go:embed config.json
var configFile embed.FS

func Init(opts ...Option) {
	cur := apply(opts...)

	config, err := configFile.Open(cur.configFilename + "." + cur.configFileType)
	if err != nil {
		panic(err)
	}

	viper.SetConfigType(cur.configFileType)
	viper.AutomaticEnv()
	err = viper.ReadConfig(config)
	if err != nil {
		panic(err)
	}
}
