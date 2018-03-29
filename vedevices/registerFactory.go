package vedevices

import (
	"log"
	"github.com/koestler/go-ve-sensor/vedirect"
)

func RegisterFactoryByModel(model string) Registers {
	switch model {
	case "bmv700Essential":
		return RegisterListBmv700Essential
	case "bmv700":
		return RegisterListBmv700
	case "bmv702":
		return RegisterListBmv702
	case "blueSolarMppt75_15":
		return RegisterListSolar
	default:
		log.Fatalf("device: unknown Bmv.Model: %v", model)
	}
	return nil
}

func RegisterFactoryByProduct(product vedirect.VeProduct) Registers {
	switch product {
	case vedirect.VeProductBmv700:
		return RegisterListBmv700
	case vedirect.VeProductBmv702:
		return RegisterListBmv702
	case vedirect.VeProductBmv700H:
		return RegisterListBmv700
	case vedirect.VeProductBlueSolarMppt70_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt75_50:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_35_rev1:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt75_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt100_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt100_30_rev1:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt100_50_rev1:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_70:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_100:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt100_50_rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt100_30_rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_35_rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt75_10:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_45:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_60:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMppt150_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMppt250_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMppt150_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMppt150_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMppt75_15:
		return RegisterListSolar
	}
	return nil
}
