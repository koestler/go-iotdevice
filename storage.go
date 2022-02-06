package main

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

type Storages struct {
	raw *dataflow.ValueStorageInstance
}

func runStorageAndDataFlow() Storages {
	log.Printf("storage: setup rawStorage")
	return Storages{
		raw: dataflow.ValueStorageCreate(),
	}
}
