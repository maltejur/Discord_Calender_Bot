package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

//Config Struct to store config data from config.yaml
type Config struct {
	Token        string `yaml:"token"`
	Logfile      string `yaml:"logfile"`
	Validtypes   string `yaml:"validtypes,flow"`
	DatabaseFile string `yaml:"databasefile"`
}

//ReadConfigFile parses file at path as yaml and returns result in Config struct
func ReadConfigFile(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var c Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&c)
	if err != nil {
		panic(err)
	}
	return c
}
