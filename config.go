package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/lasaleks/go-utils/configsrv"
	"gopkg.in/yaml.v2"
)

// Конфигурация
type Config struct {
	SVSIGNAL struct {
		ID_DRIVER   int `yaml:"id_driver"`
		DEBUG_LEVEL int `yaml:"debug_level"`
		RABBITMQ    struct {
			URL        string `yaml:"url"`
			QUEUE_NAME string `yaml:"queue_name"`
			QOS        int    `yaml:"qos"`
		} `yaml:"rabbitmq"`
		HTTP *struct {
			Address    string `yaml:"address"`
			UnixSocket string `yaml:"unixsocket"`
			User       string `yaml:"user"`
			Password   string `yaml:"password"`
		} `yaml:"http"`
		BulkInsertBufferSize int   `yaml:"bulk_insert_buffer_size"`
		BufferSize           int   `yaml:"buffer_size"`
		PeriodSave           int64 `yaml:"period_save"`
	} `yaml:"svsignal"`
	SERVER_PATH_CFG string `yaml:"server_path_cfg"`
	CONFIG_SERVER   configsrv.ConfigSrv
}

func (conf *Config) ParseConfig(config_file string) error {
	yamlFile, err := os.ReadFile(config_file)
	if err != nil {
		return fmt.Errorf("yamlFile.Get err   #%v ", err)
	}
	// 	fmt.Printf("%s\n", yamlFile)
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return fmt.Errorf("unmarshal: %v", err)
	}
	if err = conf.CONFIG_SERVER.ParseConfig(conf.SERVER_PATH_CFG); err != nil {
		return errors.Join(fmt.Errorf("file server_path_cfg:%s parse error", conf.SERVER_PATH_CFG), err)
	}
	return nil
}
