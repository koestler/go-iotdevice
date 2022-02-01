package vedirect

type VeCommand byte

const (
	VeCommandPing       VeCommand = 0x01
	VeCommandAppVersion VeCommand = 0x03
	VeCommandDeviceId   VeCommand = 0x04
	VeCommandRestart    VeCommand = 0x06
	VeCommandGet        VeCommand = 0x07
	VeCommandSet        VeCommand = 0x08
	VeCommandAsync      VeCommand = 0x0A
)

type VeResponse byte

const (
	VeResponseDone    VeResponse = 0x01
	VeResponseUnknown VeResponse = 0x03
	VeResponsePing    VeResponse = 0x05
	VeResponseGet     VeResponse = 0x07
	VeResponseSet     VeResponse = 0x08
	VeResponseAsync   VeResponse = 0x0A
)

func ResponseForCommand(command VeCommand) (response VeResponse) {
	switch command {
	case VeCommandPing:
		return VeResponsePing
	case VeCommandAppVersion:
		return VeResponseDone
	case VeCommandDeviceId:
		return VeResponseDone
	case VeCommandRestart:
		return VeResponseDone
	case VeCommandGet:
		return VeResponseGet
	case VeCommandSet:
		return VeResponseSet
	case VeCommandAsync:
		return VeResponseAsync
	}

	return VeResponseUnknown
}

type VeResponseFlag byte

const (
	VeResponseFlagOk             VeResponseFlag = 0x00
	VeResponseFlagUnknownId      VeResponseFlag = 0x01
	VeResponseFlagNotSupported   VeResponseFlag = 0x02
	VeResponseFlagParameterError VeResponseFlag = 0x04
)

type VeProduct uint16

const (
	VeProductBMV700                           VeProduct = 0x203
	VeProductBMV702                           VeProduct = 0x204
	VeProductBMV700H                          VeProduct = 0x205
	VeProductBlueSolarMPPT70_15               VeProduct = 0x0300
	VeProductBlueSolarMPPT75_50               VeProduct = 0xA040
	VeProductBlueSolarMPPT150_35              VeProduct = 0xA041
	VeProductBlueSolarMPPT75_15               VeProduct = 0xA042
	VeProductBlueSolarMPPT100_15              VeProduct = 0xA043
	VeProductBlueSolarMPPT100_30              VeProduct = 0xA044
	VeProductBlueSolarMPPT100_50              VeProduct = 0xA045
	VeProductBlueSolarMPPT150_70              VeProduct = 0xA046
	VeProductBlueSolarMPPT150_100             VeProduct = 0xA047
	VeProductBlueSolarMPPT100_50rev2          VeProduct = 0xA049
	VeProductBlueSolarMPPT100_30rev2          VeProduct = 0xA04A
	VeProductBlueSolarMPPT150_35rev2          VeProduct = 0xA04B
	VeProductBlueSolarMPPT75_10               VeProduct = 0xA04C
	VeProductBlueSolarMPPT150_45              VeProduct = 0xA04D
	VeProductBlueSolarMPPT150_60              VeProduct = 0xA04E
	VeProductBlueSolarMPPT150_85              VeProduct = 0xA04F
	VeProductSmartSolarMPPT250_100            VeProduct = 0xA050
	VeProductSmartSolarMPPT150_100            VeProduct = 0xA051
	VeProductSmartSolarMPPT150_85             VeProduct = 0xA052
	VeProductSmartSolarMPPT75_15              VeProduct = 0xA053
	VeProductSmartSolarMPPT75_10              VeProduct = 0xA054
	VeProductSmartSolarMPPT100_15             VeProduct = 0xA055
	VeProductSmartSolarMPPT100_30             VeProduct = 0xA056
	VeProductSmartSolarMPPT100_50             VeProduct = 0xA057
	VeProductSmartSolarMPPT150_35             VeProduct = 0xA058
	VeProductSmartSolarMPPT150_100rev2        VeProduct = 0xA059
	VeProductSmartSolarMPPT150_85rev2         VeProduct = 0xA05A
	VeProductSmartSolarMPPT250_70             VeProduct = 0xA05B
	VeProductSmartSolarMPPT250_85             VeProduct = 0xA05C
	VeProductSmartSolarMPPT250_60             VeProduct = 0xA05D
	VeProductSmartSolarMPPT250_45             VeProduct = 0xA05E
	VeProductSmartSolarMPPT100_20             VeProduct = 0xA05F
	VeProductSmartSolarMPPT100_2048V          VeProduct = 0xA060
	VeProductSmartSolarMPPT150_45             VeProduct = 0xA061
	VeProductSmartSolarMPPT150_60             VeProduct = 0xA062
	VeProductSmartSolarMPPT150_70             VeProduct = 0xA063
	VeProductSmartSolarMPPT250_85rev2         VeProduct = 0xA064
	VeProductSmartSolarMPPT250_100rev2        VeProduct = 0xA065
	VeProductBlueSolarMPPT100_20              VeProduct = 0xA066
	VeProductBlueSolarMPPT100_2048V           VeProduct = 0xA067
	VeProductSmartSolarMPPT250_60rev2         VeProduct = 0xA068
	VeProductSmartSolarMPPT250_70rev2         VeProduct = 0xA069
	VeProductSmartSolarMPPT150_45rev2         VeProduct = 0xA06A
	VeProductSmartSolarMPPT150_60rev2         VeProduct = 0xA06B
	VeProductSmartSolarMPPT150_70rev2         VeProduct = 0xA06C
	VeProductSmartSolarMPPT150_85rev3         VeProduct = 0xA06D
	VeProductSmartSolarMPPT150_100rev3        VeProduct = 0xA06E
	VeProductBlueSolarMPPT150_45rev2          VeProduct = 0xA06F
	VeProductBlueSolarMPPT150_60rev2          VeProduct = 0xA070
	VeProductBlueSolarMPPT150_70rev2          VeProduct = 0xA071
	VeProductSmartSolarMPPTVECan150_70        VeProduct = 0xA102
	VeProductSmartSolarMPPTVECan150_45        VeProduct = 0xA103
	VeProductSmartSolarMPPTVECan150_60        VeProduct = 0xA104
	VeProductSmartSolarMPPTVECan150_85        VeProduct = 0xA105
	VeProductSmartSolarMPPTVECan150_100       VeProduct = 0xA106
	VeProductSmartSolarMPPTVECan250_45        VeProduct = 0xA107
	VeProductSmartSolarMPPTVECan250_60        VeProduct = 0xA108
	VeProductSmartSolarMPPTVECan250_70        VeProduct = 0xA109
	VeProductSmartSolarMPPTVECan250_85        VeProduct = 0xA10A
	VeProductSmartSolarMPPTVECan250_100       VeProduct = 0xA10B
	VeProductSmartSolarMPPTVECan150_70rev2    VeProduct = 0xA10C
	VeProductSmartSolarMPPTVECan150_85rev2    VeProduct = 0xA10D
	VeProductSmartSolarMPPTVECan150_100rev2   VeProduct = 0xA10E
	VeProductBlueSolarMPPTVECan150_100        VeProduct = 0xA10F
	VeProductBlueSolarMPPTVECan250_70         VeProduct = 0xA112
	VeProductBlueSolarMPPTVECan250_100        VeProduct = 0xA113
	VeProductSmartSolarMPPTVECan250_70rev2    VeProduct = 0xA114
	VeProductSmartSolarMPPTVECan250_100rev2   VeProduct = 0xA115
	VeProductSmartSolarMPPTVECan250_85rev2    VeProduct = 0xA116
	VeProductPhoenixInverter12V250VA230V      VeProduct = 0xA231
	VeProductPhoenixInverter24V250VA230V      VeProduct = 0xA232
	VeProductPhoenixInverter48V250VA230V      VeProduct = 0xA234
	VeProductPhoenixInverter12V250VA120V      VeProduct = 0xA239
	VeProductPhoenixInverter24V250VA120V      VeProduct = 0xA23A
	VeProductPhoenixInverter48V250VA120V      VeProduct = 0xA23C
	VeProductPhoenixInverter12V375VA230V      VeProduct = 0xA241
	VeProductPhoenixInverter24V375VA230V      VeProduct = 0xA242
	VeProductPhoenixInverter48V375VA230V      VeProduct = 0xA244
	VeProductPhoenixInverter12V375VA120V      VeProduct = 0xA249
	VeProductPhoenixInverter24V375VA120V      VeProduct = 0xA24A
	VeProductPhoenixInverter48V375VA120V      VeProduct = 0xA24C
	VeProductPhoenixInverter12V500VA230V      VeProduct = 0xA251
	VeProductPhoenixInverter24V500VA230V      VeProduct = 0xA252
	VeProductPhoenixInverter48V500VA230V      VeProduct = 0xA254
	VeProductPhoenixInverter12V500VA120V      VeProduct = 0xA259
	VeProductPhoenixInverter24V500VA120V      VeProduct = 0xA25A
	VeProductPhoenixInverter48V500VA120V      VeProduct = 0xA25C
	VeProductPhoenixInverter12V800VA230V      VeProduct = 0xA261
	VeProductPhoenixInverter24V800VA230V      VeProduct = 0xA262
	VeProductPhoenixInverter48V800VA230V      VeProduct = 0xA264
	VeProductPhoenixInverter12V800VA120V      VeProduct = 0xA269
	VeProductPhoenixInverter24V800VA120V      VeProduct = 0xA26A
	VeProductPhoenixInverter48V800VA120V      VeProduct = 0xA26C
	VeProductPhoenixInverter12V1200VA230V     VeProduct = 0xA271
	VeProductPhoenixInverter24V1200VA230V     VeProduct = 0xA272
	VeProductPhoenixInverter48V1200VA230V     VeProduct = 0xA274
	VeProductPhoenixInverter12V1200VA120V     VeProduct = 0xA279
	VeProductPhoenixInverter24V1200VA120V     VeProduct = 0xA27A
	VeProductPhoenixInverter48V1200VA120V     VeProduct = 0xA27C
	VeProductPhoenixInverter12V1600VA230V     VeProduct = 0xA281
	VeProductPhoenixInverter24V1600VA230V     VeProduct = 0xA282
	VeProductPhoenixInverter48V1600VA230V     VeProduct = 0xA284
	VeProductPhoenixInverter12V2000VA230V     VeProduct = 0xA291
	VeProductPhoenixInverter24V2000VA230V     VeProduct = 0xA292
	VeProductPhoenixInverter48V2000VA230V     VeProduct = 0xA294
	VeProductPhoenixInverter12V3000VA230V     VeProduct = 0xA2A1
	VeProductPhoenixInverter24V3000VA230V     VeProduct = 0xA2A2
	VeProductPhoenixInverter48V3000VA230V     VeProduct = 0xA2A4
	VeProductPhoenixSmartIP43Charger12_50_1p1 VeProduct = 0xA340
	VeProductPhoenixSmartIP43Charger12_50_3   VeProduct = 0xA341
	VeProductPhoenixSmartIP43Charger24_25_1p1 VeProduct = 0xA342
	VeProductPhoenixSmartIP43Charger24_25_3   VeProduct = 0xA343
	VeProductPhoenixSmartIP43Charger12_30_1p1 VeProduct = 0xA344
	VeProductPhoenixSmartIP43Charger12_30_3   VeProduct = 0xA345
	VeProductPhoenixSmartIP43Charger24_16_1p1 VeProduct = 0xA346
	VeProductPhoenixSmartIP43Charger24_16_3   VeProduct = 0xA347
	VeProductBMV712Smart                      VeProduct = 0xA381
	VeProductBMV710HSmart                     VeProduct = 0xA382
	VeProductBMV712SmartRev2                  VeProduct = 0xA383
	VeProductSmartShunt500A_50mV              VeProduct = 0xA389
	VeProductSmartShunt1000A_50mV             VeProduct = 0xA38A
	VeProductSmartShunt2000A_50mV             VeProduct = 0xA38B
)

func (product VeProduct) String() string {
	switch product {
	case VeProductBMV700:
		return "BMV-700"
	case VeProductBMV702:
		return "BMV-702"
	case VeProductBMV700H:
		return "BMV-700H"
	case VeProductBlueSolarMPPT70_15:
		return "BlueSolar MPPT 70|15"
	case VeProductBlueSolarMPPT75_50:
		return "BlueSolar MPPT 75|50"
	case VeProductBlueSolarMPPT150_35:
		return "BlueSolar MPPT 150|35"
	case VeProductBlueSolarMPPT75_15:
		return "BlueSolar MPPT 75|15"
	case VeProductBlueSolarMPPT100_15:
		return "BlueSolar MPPT 100|15"
	case VeProductBlueSolarMPPT100_30:
		return "BlueSolar MPPT 100|30"
	case VeProductBlueSolarMPPT100_50:
		return "BlueSolar MPPT 100|50"
	case VeProductBlueSolarMPPT150_70:
		return "BlueSolar MPPT 150|70"
	case VeProductBlueSolarMPPT150_100:
		return "BlueSolar MPPT 150|100"
	case VeProductBlueSolarMPPT100_50rev2:
		return "BlueSolar MPPT 100|50 rev2"
	case VeProductBlueSolarMPPT100_30rev2:
		return "BlueSolar MPPT 100|30 rev2"
	case VeProductBlueSolarMPPT150_35rev2:
		return "BlueSolar MPPT 150|35 rev2"
	case VeProductBlueSolarMPPT75_10:
		return "BlueSolar MPPT 75|10"
	case VeProductBlueSolarMPPT150_45:
		return "BlueSolar MPPT 150|45"
	case VeProductBlueSolarMPPT150_60:
		return "BlueSolar MPPT 150|60"
	case VeProductBlueSolarMPPT150_85:
		return "BlueSolar MPPT 150|85"
	case VeProductSmartSolarMPPT250_100:
		return "SmartSolar MPPT 250|100"
	case VeProductSmartSolarMPPT150_100:
		return "SmartSolar MPPT 150|100"
	case VeProductSmartSolarMPPT150_85:
		return "SmartSolar MPPT 150|85"
	case VeProductSmartSolarMPPT75_15:
		return "SmartSolar MPPT 75|15"
	case VeProductSmartSolarMPPT75_10:
		return "SmartSolar MPPT 75|10"
	case VeProductSmartSolarMPPT100_15:
		return "SmartSolar MPPT 100|15"
	case VeProductSmartSolarMPPT100_30:
		return "SmartSolar MPPT 100|30"
	case VeProductSmartSolarMPPT100_50:
		return "SmartSolar MPPT 100|50"
	case VeProductSmartSolarMPPT150_35:
		return "SmartSolar MPPT 150|35"
	case VeProductSmartSolarMPPT150_100rev2:
		return "SmartSolar MPPT 150|100 rev2"
	case VeProductSmartSolarMPPT150_85rev2:
		return "SmartSolar MPPT 150|85 rev2"
	case VeProductSmartSolarMPPT250_70:
		return "SmartSolar MPPT 250|70"
	case VeProductSmartSolarMPPT250_85:
		return "SmartSolar MPPT 250|85"
	case VeProductSmartSolarMPPT250_60:
		return "SmartSolar MPPT 250|60"
	case VeProductSmartSolarMPPT250_45:
		return "SmartSolar MPPT 250|45"
	case VeProductSmartSolarMPPT100_20:
		return "SmartSolar MPPT 100|20"
	case VeProductSmartSolarMPPT100_2048V:
		return "SmartSolar MPPT 100|20 48V"
	case VeProductSmartSolarMPPT150_45:
		return "SmartSolar MPPT 150|45"
	case VeProductSmartSolarMPPT150_60:
		return "SmartSolar MPPT 150|60"
	case VeProductSmartSolarMPPT150_70:
		return "SmartSolar MPPT 150|70"
	case VeProductSmartSolarMPPT250_85rev2:
		return "SmartSolar MPPT 250|85 rev2"
	case VeProductSmartSolarMPPT250_100rev2:
		return "SmartSolar MPPT 250|100 rev2"
	case VeProductBlueSolarMPPT100_20:
		return "BlueSolar MPPT 100|20"
	case VeProductBlueSolarMPPT100_2048V:
		return "BlueSolar MPPT 100|20 48V"
	case VeProductSmartSolarMPPT250_60rev2:
		return "SmartSolar MPPT 250|60 rev2"
	case VeProductSmartSolarMPPT250_70rev2:
		return "SmartSolar MPPT 250|70 rev2"
	case VeProductSmartSolarMPPT150_45rev2:
		return "SmartSolar MPPT 150|45 rev2"
	case VeProductSmartSolarMPPT150_60rev2:
		return "SmartSolar MPPT 150|60 rev2"
	case VeProductSmartSolarMPPT150_70rev2:
		return "SmartSolar MPPT 150|70 rev2"
	case VeProductSmartSolarMPPT150_85rev3:
		return "SmartSolar MPPT 150|85 rev3"
	case VeProductSmartSolarMPPT150_100rev3:
		return "SmartSolar MPPT 150|100 rev3"
	case VeProductBlueSolarMPPT150_45rev2:
		return "BlueSolar MPPT 150|45 rev2"
	case VeProductBlueSolarMPPT150_60rev2:
		return "BlueSolar MPPT 150|60 rev2"
	case VeProductBlueSolarMPPT150_70rev2:
		return "BlueSolar MPPT 150|70 rev2"
	case VeProductSmartSolarMPPTVECan150_70:
		return "SmartSolar MPPT VE.Can 150/70"
	case VeProductSmartSolarMPPTVECan150_45:
		return "SmartSolar MPPT VE.Can 150/45"
	case VeProductSmartSolarMPPTVECan150_60:
		return "SmartSolar MPPT VE.Can 150/60"
	case VeProductSmartSolarMPPTVECan150_85:
		return "SmartSolar MPPT VE.Can 150/85"
	case VeProductSmartSolarMPPTVECan150_100:
		return "SmartSolar MPPT VE.Can 150/100"
	case VeProductSmartSolarMPPTVECan250_45:
		return "SmartSolar MPPT VE.Can 250/45"
	case VeProductSmartSolarMPPTVECan250_60:
		return "SmartSolar MPPT VE.Can 250/60"
	case VeProductSmartSolarMPPTVECan250_70:
		return "SmartSolar MPPT VE.Can 250/70"
	case VeProductSmartSolarMPPTVECan250_85:
		return "SmartSolar MPPT VE.Can 250/85"
	case VeProductSmartSolarMPPTVECan250_100:
		return "SmartSolar MPPT VE.Can 250/100"
	case VeProductSmartSolarMPPTVECan150_70rev2:
		return "SmartSolar MPPT VE.Can 150/70 rev2"
	case VeProductSmartSolarMPPTVECan150_85rev2:
		return "SmartSolar MPPT VE.Can 150/85 rev2"
	case VeProductSmartSolarMPPTVECan150_100rev2:
		return "SmartSolar MPPT VE.Can 150/100 rev2"
	case VeProductBlueSolarMPPTVECan150_100:
		return "BlueSolar MPPT VE.Can 150/100"
	case VeProductBlueSolarMPPTVECan250_70:
		return "BlueSolar MPPT VE.Can 250/70"
	case VeProductBlueSolarMPPTVECan250_100:
		return "BlueSolar MPPT VE.Can 250/100"
	case VeProductSmartSolarMPPTVECan250_70rev2:
		return "SmartSolar MPPT VE.Can 250/70 rev2"
	case VeProductSmartSolarMPPTVECan250_100rev2:
		return "SmartSolar MPPT VE.Can 250/100 rev2"
	case VeProductSmartSolarMPPTVECan250_85rev2:
		return "SmartSolar MPPT VE.Can 250/85 rev2"
	case VeProductPhoenixInverter12V250VA230V:
		return "Phoenix Inverter 12V 250VA 230V"
	case VeProductPhoenixInverter24V250VA230V:
		return "Phoenix Inverter 24V 250VA 230V"
	case VeProductPhoenixInverter48V250VA230V:
		return "Phoenix Inverter 48V 250VA 230V"
	case VeProductPhoenixInverter12V250VA120V:
		return "Phoenix Inverter 12V 250VA 120V"
	case VeProductPhoenixInverter24V250VA120V:
		return "Phoenix Inverter 24V 250VA 120V"
	case VeProductPhoenixInverter48V250VA120V:
		return "Phoenix Inverter 48V 250VA 120V"
	case VeProductPhoenixInverter12V375VA230V:
		return "Phoenix Inverter 12V 375VA 230V"
	case VeProductPhoenixInverter24V375VA230V:
		return "Phoenix Inverter 24V 375VA 230V"
	case VeProductPhoenixInverter48V375VA230V:
		return "Phoenix Inverter 48V 375VA 230V"
	case VeProductPhoenixInverter12V375VA120V:
		return "Phoenix Inverter 12V 375VA 120V"
	case VeProductPhoenixInverter24V375VA120V:
		return "Phoenix Inverter 24V 375VA 120V"
	case VeProductPhoenixInverter48V375VA120V:
		return "Phoenix Inverter 48V 375VA 120V"
	case VeProductPhoenixInverter12V500VA230V:
		return "Phoenix Inverter 12V 500VA 230V"
	case VeProductPhoenixInverter24V500VA230V:
		return "Phoenix Inverter 24V 500VA 230V"
	case VeProductPhoenixInverter48V500VA230V:
		return "Phoenix Inverter 48V 500VA 230V"
	case VeProductPhoenixInverter12V500VA120V:
		return "Phoenix Inverter 12V 500VA 120V"
	case VeProductPhoenixInverter24V500VA120V:
		return "Phoenix Inverter 24V 500VA 120V"
	case VeProductPhoenixInverter48V500VA120V:
		return "Phoenix Inverter 48V 500VA 120V"
	case VeProductPhoenixInverter12V800VA230V:
		return "Phoenix Inverter 12V 800VA 230V"
	case VeProductPhoenixInverter24V800VA230V:
		return "Phoenix Inverter 24V 800VA 230V"
	case VeProductPhoenixInverter48V800VA230V:
		return "Phoenix Inverter 48V 800VA 230V"
	case VeProductPhoenixInverter12V800VA120V:
		return "Phoenix Inverter 12V 800VA 120V"
	case VeProductPhoenixInverter24V800VA120V:
		return "Phoenix Inverter 24V 800VA 120V"
	case VeProductPhoenixInverter48V800VA120V:
		return "Phoenix Inverter 48V 800VA 120V"
	case VeProductPhoenixInverter12V1200VA230V:
		return "Phoenix Inverter 12V 1200VA 230V"
	case VeProductPhoenixInverter24V1200VA230V:
		return "Phoenix Inverter 24V 1200VA 230V"
	case VeProductPhoenixInverter48V1200VA230V:
		return "Phoenix Inverter 48V 1200VA 230V"
	case VeProductPhoenixInverter12V1200VA120V:
		return "Phoenix Inverter 12V 1200VA 120V"
	case VeProductPhoenixInverter24V1200VA120V:
		return "Phoenix Inverter 24V 1200VA 120V"
	case VeProductPhoenixInverter48V1200VA120V:
		return "Phoenix Inverter 48V 1200VA 120V"
	case VeProductPhoenixInverter12V1600VA230V:
		return "Phoenix Inverter 12V 1600VA 230V"
	case VeProductPhoenixInverter24V1600VA230V:
		return "Phoenix Inverter 24V 1600VA 230V"
	case VeProductPhoenixInverter48V1600VA230V:
		return "Phoenix Inverter 48V 1600VA 230V"
	case VeProductPhoenixInverter12V2000VA230V:
		return "Phoenix Inverter 12V 2000VA 230V"
	case VeProductPhoenixInverter24V2000VA230V:
		return "Phoenix Inverter 24V 2000VA 230V"
	case VeProductPhoenixInverter48V2000VA230V:
		return "Phoenix Inverter 48V 2000VA 230V"
	case VeProductPhoenixInverter12V3000VA230V:
		return "Phoenix Inverter 12V 3000VA 230V"
	case VeProductPhoenixInverter24V3000VA230V:
		return "Phoenix Inverter 24V 3000VA 230V"
	case VeProductPhoenixInverter48V3000VA230V:
		return "Phoenix Inverter 48V 3000VA 230V"
	case VeProductPhoenixSmartIP43Charger12_50_1p1:
		return "Phoenix Smart IP43 Charger 12|50 (1+1)"
	case VeProductPhoenixSmartIP43Charger12_50_3:
		return "Phoenix Smart IP43 Charger 12|50 (3)"
	case VeProductPhoenixSmartIP43Charger24_25_1p1:
		return "Phoenix Smart IP43 Charger 24|25 (1+1)"
	case VeProductPhoenixSmartIP43Charger24_25_3:
		return "Phoenix Smart IP43 Charger 24|25 (3)"
	case VeProductPhoenixSmartIP43Charger12_30_1p1:
		return "Phoenix Smart IP43 Charger 12|30 (1+1)"
	case VeProductPhoenixSmartIP43Charger12_30_3:
		return "Phoenix Smart IP43 Charger 12|30 (3)"
	case VeProductPhoenixSmartIP43Charger24_16_1p1:
		return "Phoenix Smart IP43 Charger 24|16 (1+1)"
	case VeProductPhoenixSmartIP43Charger24_16_3:
		return "Phoenix Smart IP43 Charger 24|16 (3)"
	case VeProductBMV712Smart:
		return "BMV-712 Smart"
	case VeProductBMV710HSmart:
		return "BMV-710H Smart"
	case VeProductBMV712SmartRev2:
		return "BMV-712 Smart Rev2"
	case VeProductSmartShunt500A_50mV:
		return "SmartShunt 500A/50mV"
	case VeProductSmartShunt1000A_50mV:
		return "SmartShunt 1000A/50mV"
	case VeProductSmartShunt2000A_50mV:
		return "SmartShunt 2000A/50mV"
	}
	return ""
}
