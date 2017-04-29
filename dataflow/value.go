package dataflow

type Value struct {
	Name string
	Device *Device
	Value float64
	Unit string
}

/*
// this function must not be used concurrently
func ValueCreate(name, unit string) (valueId ValueId) {
	valueId = ValueId(len(valueDb) + 1)

	valueDb[valueId] = Value{
		Id:   valueId,
		Name: name,
		Unit: unit,
	}

	return
}

func ValuePrintToLog() {
	log.Printf("valueDb holds the current devices:")
	for deviceId, device := range deviceDb {
		log.Printf("- %v: %v", deviceId, device)
	}
}

*/