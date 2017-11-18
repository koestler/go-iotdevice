package vedevices

import (
	"log"
)

func BmvRegisterFactory(model string) Registers {
	switch model {
	case "bmv700Essential":
		return RegisterListBmv700Essential
	case "bmv700":
		return RegisterListBmv700
	case "bmv702":
		return RegisterListBmv702
	default:
		log.Fatalf("device: unknown Bmv.Model: %v", model)
	}
	return nil
}
