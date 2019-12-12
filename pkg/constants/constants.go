package constants

const (
	ScreenWidth  int = 160
	ScreenHeight     = 144
)

const (
	ScreenScaling float64 = 2.0
)

const (
	DIVAddress             uint16 = 0xFF04
	TIMAAddress                   = 0xFF05
	TMAAddress                    = 0xFF06
	TACAddress                    = 0xFF07
	LCDCAddress                   = 0xFF40
	STATAddress                   = 0xFF41
	ScrollYAddress                = 0xFF42
	ScrollXAddress                = 0xFF43
	LYAddress                     = 0xFF44
	LYCAddress                    = 0xFF45
	BGPAddress                    = 0xFF47
	OBP0Address                   = 0xFF48
	OBP1Address                   = 0xFF49
	WindowYAddress                = 0xFF4A
	WindowXAddress                = 0xFF4B
	JoypadRegisterAddress         = 0xFF00
	InterruptFlagAddress          = 0xFF0F
	InterruptEnableAddress        = 0xFFFF
	StackStartAddress             = 0xFFFE
)
