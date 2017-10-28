package main

import (
	"github.com/koestler/go-ve-sensor/config"
	"github.com/koestler/go-ve-sensor/mongo"
	"github.com/koestler/go-ve-sensor/vedata"
	"github.com/koestler/go-ve-sensor/vehttp"
	"log"
	"github.com/koestler/go-ve-sensor/cam"
)

func main() {
	log.Print("main: start go-ve-sensor...")

	// start http server
	httpdConfig, err := config.GetHttpdConfig()
	if err == nil {
		log.Print("main: start http server, config=%v", httpdConfig)
		go vehttp.Run(httpdConfig.Bind, httpdConfig.Port, HttpRoutes)
	} else {
		log.Printf("main: skip http server, err=%v", err)
	}

	// startup Bmv devices
	log.Print("start bmv devices")
	for _, bmvConfig := range config.GetBmvConfigs() {
		BmvStart(bmvConfig)
	}

	// startup Cam devices
	log.Print("start camera devices")
	for _, camConfig := range config.GetCamConfigs() {
		cam.FtpCamStart(camConfig)
	}

	// run database synchronization routine
	log.Print("main: start database")
	vedata.Run()

	// initialize mongodb
	mongoConfig, err := config.GetMongoConfig()
	if err == nil {
		log.Printf("main: start mongo database connection, config=%v", mongoConfig)
		mongoSession := mongo.GetSession(mongoConfig.MongoHost)
		defer mongoSession.Close()
		mongo.Run(mongoSession, mongoConfig.DatabaseName, mongoConfig.RawValuesIntervall)
	} else {
		log.Printf("main: skip mongo database initialization (%v)", err)
	}

	log.Print("main: start completed")
	select {}

}
