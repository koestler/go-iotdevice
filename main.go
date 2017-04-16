package main

import (
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/mongo"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vehttp"
	"log"
)

func main() {
	log.Print("start go-ve-sensor...")

	// start http server
	httpdConfig, err := config.GetHttpdConfig()
	if err == nil {
		log.Print("start http server, config=%v", httpdConfig)
		go vehttp.Run(httpdConfig.Bind, httpdConfig.Port, HttpRoutes)
	} else {
		log.Printf("skip http server, err=%v", err)
	}

	// startup Bmv Device
	log.Print("start devices")
	for _, bmvConfig := range config.GetBmvConfigs() {
		BmvStart(bmvConfig)
	}

	// run database synchronization routine
	log.Print("start database")
	vedata.Run()

	// initialize mongodb
	mongoConfig, err := config.GetMongoConfig()
	log.Printf("mongoConfig=%v", mongoConfig)
	if err == nil {
		log.Printf("start mongodatabase writer")
		mongoSession := mongo.GetSession(mongoConfig.MongoHost)
		defer mongoSession.Close()
		mongo.Run(mongoSession, mongoConfig.DatabaseName, mongoConfig.RawValuesIntervall)
	} else {
		log.Printf("skip mongo database initialization")
	}

	log.Print("start completed")
	select {}

}
