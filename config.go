package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
)

const (
	DevoUsbPID = "0C5B"
	DevoUsbVID = "16D0"
)

var defaultHeaders = []headerConfig{
	headerConfig{
		Name:      "Time",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SetT1",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Temp1",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "dc1",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Err1",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SetT2",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Temp2",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "dc2",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Err2",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SetT3",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Temp3",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "dc3",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Err3",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SetT4",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Temp4",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "dc4",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Err4",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "intT4",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "ExtCur",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "ExtPWM",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "ExtTmp",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Overht",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "FAULT",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SetRPM",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "RPM",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "FT",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "FTAVG",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Puller",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "MemFree",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Status",
		Validator: `[a-zA-Z]+`},
	headerConfig{
		Name:      "WndrSpd",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "PosSpd",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Length",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "Volume",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SpDia",
		Validator: `-?[0-9]\d*(\.\d+)?`},
	headerConfig{
		Name:      "SpFill",
		Validator: `-?[0-9]\d*(\.\d+)?`}}

type headerConfig struct {
	Name      string `yaml:"name"`
	Validator string `yaml:"validator"`
}

// Configuration is a struct that uses the config.yaml to initialize certain configurable aspects of the application
// For example what validator to use for a header in the serial communication
type Configuration struct {
	Headers []headerConfig `yaml:"headers"`
}

// createDefaultConfiguration creates a config.yaml file with default values and returns the created configuration struct
func createDefaultConfiguration(configFile string) (*Configuration, error) {
	defaultConfig := &Configuration{Headers: defaultHeaders}
	configInBytes, err := yaml.Marshal(&defaultConfig)

	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(configFile, configInBytes, os.ModePerm)

	if err != nil {
		return nil, err
	}

	return defaultConfig, nil
}

// LoadConfiguration tries to read in an existing config file and parse and return it.
// If no existing file is found it will generate a default one
func LoadConfiguration(configFile string) (*Configuration, error) {
	var configuration *Configuration
	if _, err := os.Stat(configFile); err == nil {
		configBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(configBytes, &configuration)
		if err != nil {
			return nil, err
		}
		return configuration, nil

	}
	log.Println("Config file not found creating a default version")
	configuration, err := createDefaultConfiguration(configFile)
	if err != nil {
		return nil, err
	}
	return configuration, nil
}
