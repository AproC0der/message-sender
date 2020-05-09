package config

import (
	"fmt"

	"github.com/toolkits/pkg/file"
)

type Config struct {
	Redis    redisSection    `yaml:"redis"`
	Consumer consumerSection `yaml:"consumer"`
	Url urlSection `yaml:"url"`
}

type redisSection struct {
	Addr    string         `yaml:"addr"`
	Pass    string         `yaml:"pass"`
	Idle    int            `yaml:"idle"`
	DB      int            `yaml:"db"`
	Timeout timeoutSection `yaml:"timeout"`
}

type urlSection struct {
	Ip string `yaml:"ip"`
	Port string `yaml:"port"`
	Corpid string `yaml:"corpid"`
	Pwd string `yanl:"pwd"`
	Mobile string `yaml:"mobile"`
}
type timeoutSection struct {
	Conn  int `yaml:"conn"`
	Read  int `yaml:"read"`
	Write int `yaml:"write"`
}

type consumerSection struct {
	Queue  string `yaml:"queue"`
	Worker int    `yaml:"worker"`
}

var yaml Config

func Get() Config {
	return yaml
}

func ParseConfig(yf string) error {
	err := file.ReadYaml(yf, &yaml)
	if err != nil {
		return fmt.Errorf("cannot read yml[%s]: %v", yf, err)
	}
	return nil
}
