package bmv

import (
	"log"
)

func BmvRegisterFactory(model string) Registers {
	switch model {
	case "bmv700Essential":
		return RegisterList700Essential
	case "bmv700":
		return RegisterList700
	case "bmv702":
		return RegisterList702
	default:
		log.Fatalf("device: unknown Bmv.Model: %v", model)
	}
	return nil
}
