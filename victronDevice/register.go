package victronDevice

import (
	"github.com/koestler/go-iotdevice/v3/dataflow"
)

func FilterRegisters(input []VictronRegister, registerFilter dataflow.RegisterFilterConf) (output []VictronRegister) {
	output = make([]VictronRegister, 0, len(input))
	f := dataflow.RegisterFilter(registerFilter)
	for _, r := range input {
		if f(r) {
			output = append(output, r)
		}
	}
	return
}
func addToRegisterDb(rdb *dataflow.RegisterDb, registers []VictronRegister) {
	dataflowRegisters := make([]dataflow.RegisterStruct, len(registers))
	for i, r := range registers {
		dataflowRegisters[i] = r.RegisterStruct
	}
	rdb.AddStruct(dataflowRegisters...)
}
