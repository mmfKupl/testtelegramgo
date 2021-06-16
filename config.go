package testtelegramgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type AppConfig struct {
	Password                string `json:"password"`
	Tel                     string `json:"tel"`
	ApiId                   string `json:"apiId"`
	ApiHash                 string `json:"apiHash"`
	ChatID                  int64  `json:"chatID"`
	DisplayedMessagesAmount uint   `json:"displayedMessagesAmount"`
	Port                    string `json:"port"`
	AdminPassword           string `json:"adminPassword"`
}

const defaultConfigName string = "config.json"
const configDir string = "configs"

func GetAppConfig(configName string) (*AppConfig, error) {
	if configName == "" {
		configName = defaultConfigName
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	pathToConfig := path.Join(dir, configDir, configName)

	if _, err := os.Stat(pathToConfig); err != nil {
		return nil, fmt.Errorf("Config file (%s) not found. ", pathToConfig)
	}

	bytesConfig, err := os.ReadFile(pathToConfig)
	if err != nil {
		return nil, err
	}

	var config *AppConfig
	err = json.Unmarshal(bytesConfig, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
