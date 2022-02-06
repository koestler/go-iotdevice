package victron

import (
	"github.com/koestler/go-iotdevice/dataflow"
	"github.com/koestler/go-iotdevice/vedirect"
)

func RegisterFactoryByProduct(product vedirect.VeProduct) dataflow.Registers {
	switch product {
	case vedirect.VeProductBMV700:
		return RegisterListBmv700
	case vedirect.VeProductBMV702:
		return RegisterListBmv702
	case vedirect.VeProductBMV700H:
		return RegisterListBmv700
	case vedirect.VeProductBlueSolarMPPT70_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT75_50:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_35:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT75_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_15:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_30:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_50:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_70:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_100:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_50rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_30rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_35rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT75_10:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_45:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_60:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT75_15:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT75_10:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT100_15:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT100_30:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT100_50:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_35:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_100rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_85rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_70:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_60:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_45:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT100_20:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT100_2048V:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_45:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_60:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_70:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_85rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_100rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_20:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT100_2048V:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_60rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT250_70rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_45rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_60rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_70rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_85rev3:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPT150_100rev3:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_45rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_60rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPT150_70rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_70:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_45:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_60:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_45:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_60:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_70:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_85:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_70rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_85rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan150_100rev2:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPTVECan150_100:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPTVECan250_70:
		return RegisterListSolar
	case vedirect.VeProductBlueSolarMPPTVECan250_100:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_70rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_100rev2:
		return RegisterListSolar
	case vedirect.VeProductSmartSolarMPPTVECan250_85rev2:
		return RegisterListSolar
	case vedirect.VeProductPhoenixInverter12V250VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V250VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V250VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V250VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V250VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V250VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V375VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V375VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V375VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V375VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V375VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V375VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V500VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V500VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V500VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V500VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V500VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V500VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V800VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V800VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V800VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V800VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V800VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V800VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V1200VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V1200VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V1200VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V1200VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V1200VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V1200VA120V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V1600VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V1600VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V1600VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V2000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V2000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V2000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter12V3000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter24V3000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixInverter48V3000VA230V:
		return RegisterPhoenixInverter
	case vedirect.VeProductPhoenixSmartIP43Charger12_50_1p1:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger12_50_3:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger24_25_1p1:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger24_25_3:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger12_30_1p1:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger12_30_3:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger24_16_1p1:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductPhoenixSmartIP43Charger24_16_3:
		return RegisterPhoenixSmartCharger
	case vedirect.VeProductBMV712Smart:
		return RegisterListBmv712
	case vedirect.VeProductBMV710HSmart:
		return RegisterListBmv712
	case vedirect.VeProductBMV712SmartRev2:
		return RegisterListBmv712
	case vedirect.VeProductSmartShunt500A_50mV:
		return RegisterListBmv712
	case vedirect.VeProductSmartShunt1000A_50mV:
		return RegisterListBmv712
	case vedirect.VeProductSmartShunt2000A_50mV:
		return RegisterListBmv712
	}
	return nil
}