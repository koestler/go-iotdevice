package main

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"log"
)

type Storages struct {
	raw     *dataflow.ValueStorageInstance
	rounded *dataflow.ValueStorageInstance
}

func runStorageAndDataFlow() Storages {
	log.Printf("storage: setup rawStorage and roundedStorage")

	// setup dataflow pipeline
	// 1. sources:
	// those are appended by separate routines

	// 2. storage for raw values
	rawStorage := dataflow.ValueStorageCreate()

	// 3. rounder
	rounder := dataflow.RounderCreate()

	// 4. storage for rounded values
	roundedStorage := dataflow.ValueStorageCreate()

	// chain those
	rawStorage.Append(rounder)
	rounder.Append(roundedStorage)

	return Storages{
		rawStorage,
		roundedStorage,
	}
}
