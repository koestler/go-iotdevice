package config

import (
	"errors"
	"fmt"
)

type MongoConfig struct {
	MongoHost          string
	DatabaseName       string
	RawValuesIntervall int
}

func GetMongoConfig() (mongoConfig *MongoConfig, err error) {
	mongoConfig = &MongoConfig{
		MongoHost:          "127.0.0.1",
		DatabaseName:       "go-ve-sensor",
		RawValuesIntervall: 2000,
	}

	// check if mongo sections exists
	_, err = config.GetSection("Mongo")
	if err != nil {
		// Section Mongo does not exist
		return nil, errors.New("no mongo configuration found")
	}

	err = config.Section("Mongo").MapTo(mongoConfig)

	if err != nil {
		return nil, fmt.Errorf("cannot read mongo configuration: %v", err)
	}

	return
}
