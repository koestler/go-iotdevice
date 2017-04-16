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
