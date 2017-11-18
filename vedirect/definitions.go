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

type VeResponseFlag byte

const (
	VeResponseFlagOk             VeResponseFlag = 0x00
	VeResponseFlagUnknownId      VeResponseFlag = 0x01
	VeResponseFlagNotSupported   VeResponseFlag = 0x02
	VeResponseFlagParameterError VeResponseFlag = 0x04
)

type VeProduct uint16

const (
	VeProductBmv700                   VeProduct = 0x0203
	VeProductBmv702                   VeProduct = 0x0204
	VeProductBmv700H                  VeProduct = 0x0205
	VeProductBlueSolarMppt70_15       VeProduct = 0x0300
	VeProductBlueSolarMppt75_50       VeProduct = 0xA040
	VeProductBlueSolarMppt150_35_rev1 VeProduct = 0xA041
	VeProductBlueSolarMppt75_15       VeProduct = 0xA042
	VeProductBlueSolarMppt100_15      VeProduct = 0xA043
	VeProductBlueSolarMppt100_30_rev1 VeProduct = 0xA044
	VeProductBlueSolarMppt100_50_rev1 VeProduct = 0xA045
	VeProductBlueSolarMppt150_70      VeProduct = 0xA046
	VeProductBlueSolarMppt150_100     VeProduct = 0xA047
	VeProductBlueSolarMppt100_50_rev2 VeProduct = 0xA049
	VeProductBlueSolarMppt100_30_rev2 VeProduct = 0xA04A
	VeProductBlueSolarMppt150_35_rev2 VeProduct = 0xA04B
	VeProductBlueSolarMppt75_10       VeProduct = 0xA04C
	VeProductBlueSolarMppt150_45      VeProduct = 0xA04D
	VeProductBlueSolarMppt150_60      VeProduct = 0xA04E
	VeProductBlueSolarMppt150_85      VeProduct = 0xA04F
	VeProductSmartSolarMppt250_100    VeProduct = 0xA050
	VeProductSmartSolarMppt150_100    VeProduct = 0xA051
	VeProductSmartSolarMppt150_85     VeProduct = 0xA052
	VeProductSmartSolarMppt75_15      VeProduct = 0xA053
)

func (product VeProduct) String() string {
	switch product {
	case VeProductBmv700:
		return "Bmv700"
	case VeProductBmv702:
		return "Bmv702"
	case VeProductBmv700H:
		return "Bmv700H"
	case VeProductBlueSolarMppt70_15:
		return "BlueSolarMppt70_15"
	case VeProductBlueSolarMppt75_50:
		return "BlueSolarMppt75_50"
	case VeProductBlueSolarMppt150_35_rev1:
		return "BlueSolarMppt150_35_rev1"
	case VeProductBlueSolarMppt75_15:
		return "BlueSolarMppt75_15"
	case VeProductBlueSolarMppt100_15:
		return "BlueSolarMppt100_15"
	case VeProductBlueSolarMppt100_30_rev1:
		return "BlueSolarMppt100_30_rev1"
	case VeProductBlueSolarMppt100_50_rev1:
		return "BlueSolarMppt100_50_rev1"
	case VeProductBlueSolarMppt150_70:
		return "BlueSolarMppt150_70"
	case VeProductBlueSolarMppt150_100:
		return "BlueSolarMppt150_100"
	case VeProductBlueSolarMppt100_50_rev2:
		return "BlueSolarMppt100_50_rev2"
	case VeProductBlueSolarMppt100_30_rev2:
		return "BlueSolarMppt100_30_rev2"
	case VeProductBlueSolarMppt150_35_rev2:
		return "BlueSolarMppt150_35_rev2"
	case VeProductBlueSolarMppt75_10:
		return "BlueSolarMppt75_10"
	case VeProductBlueSolarMppt150_45:
		return "BlueSolarMppt150_45"
	case VeProductBlueSolarMppt150_60:
		return "BlueSolarMppt150_60"
	case VeProductBlueSolarMppt150_85:
		return "BlueSolarMppt150_85"
	case VeProductSmartSolarMppt250_100:
		return "SmartSolarMppt250_100"
	case VeProductSmartSolarMppt150_100:
		return "SmartSolarMppt150_100"
	case VeProductSmartSolarMppt150_85:
		return "SmartSolarMppt150_85"
	case VeProductSmartSolarMppt75_15:
		return "SmartSolarMppt75_15"
	}
	return ""
}
