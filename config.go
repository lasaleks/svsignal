package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Конфигурация
type Config struct {
	CONFIG struct {
		ID_DRIVER   int `yaml:"id_driver"`
		DEBUG_LEVEL int `yaml:"debug_level"`
		RABBITMQ    struct {
			URL        string `yaml:"url"`
			QUEUE_NAME string `yaml:"queue_name"`
			QOS        int    `yaml:"qos"`
		} `yaml:"rabbitmq"`
		MYSQL struct {
			HOST     string `yaml:"host"`
			USER     string `yaml:"user"`
			PASSWORD string `yaml:"password"`
			PORT     int    `yaml:"port"`
			DATABASE string `yaml:"database"`
		} `yaml:"mysql"`
		HTTP *struct {
			Address    string `yaml:"address"`
			UnixSocket string `yaml:"unixsocket"`
			User       string `yaml:"user"`
			Password   string `yaml:"password"`
		} `yaml:"http"`
		TimeZone          string `yaml:"TimeZone"`
		MaxMultiplyInsert int    `yaml:"max_multiply_insert"`
		BuffSize          int    `yaml:"buff_size"`
	}
}

func (conf *Config) parseConfig(config_file string) *Config {
	yamlFile, err := ioutil.ReadFile(config_file)
	if err != nil {
		panic(fmt.Errorf("yamlFile.Get err   #%v ", err))
	}
	// 	fmt.Printf("%s\n", yamlFile)
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		fmt.Printf("Unmarshal: %v", err)
	}
	return conf
}
