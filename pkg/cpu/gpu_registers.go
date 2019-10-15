package cpu

import "github.com/tbtommyb/goboy/pkg/utils"

type GPUStatus byte
type GPUControl byte
type LCDCFlag byte
type LCDStatusFlag byte

const (
	LCDDisplayEnable           LCDCFlag = 0x80
	WindowTileMapDisplaySelect          = 0x40
	WindowDisplayEnable                 = 0x20
	DataSelect                          = 0x10
	BGTileMapDisplaySelect              = 0x8
	SpriteSize                          = 0x4
	SpriteEnable                        = 0x2
	WindowDisplayPriority               = 0x1
)

const (
	MatchFlag              LCDStatusFlag = 2
	AccessEnabledInterrupt               = 3
	VBlankInterrupt                      = 4
	OAMInterrupt                         = 5
	MatchInterrupt                       = 6
)

// LCD Control

func (gpu *GPU) getControl() GPUControl {
	return GPUControl(gpu.cpu.getLCDC())
}

func (gpu *GPU) setControl(control GPUControl) {
	gpu.cpu.setLCDC(byte(control))
}

func (control GPUControl) isDisplayEnabled() bool {
	return control.isSet(LCDDisplayEnable)
}

func (control GPUControl) isWindowEnabled() bool {
	return control.isSet(WindowDisplayEnable)
}

func (control GPUControl) isSpriteEnabled() bool {
	return control.isSet(SpriteEnable)
}

func (control GPUControl) isHighWindowAddress() bool {
	return control.isSet(WindowTileMapDisplaySelect)
}

func (control GPUControl) isHighBGDataAddress() bool {
	return !control.isSet(DataSelect)
}

func (control GPUControl) isHighBGStartAddress() bool {
	return control.isSet(BGTileMapDisplaySelect)
}

func (control GPUControl) isSet(flag LCDCFlag) bool {
	return (byte(control) & byte(flag)) > 0
}

// LCD Status

func (gpu *GPU) getStatus() GPUStatus {
	return GPUStatus(gpu.cpu.getSTAT())
}

func (gpu *GPU) setStatus(status GPUStatus) {
	gpu.cpu.setSTAT(byte(status))
}

func (status GPUStatus) mode() Mode {
	return Mode(status & ModeMask)
}

func (status GPUStatus) setMode(mode Mode) GPUStatus {
	return GPUStatus((byte(status) & StatusModeResetMask) | byte(mode))
}

func (status GPUStatus) isModeInterruptSet() bool {
	mode := status.mode()
	if mode == TransferringMode {
		return false
	}
	return utils.IsSet(byte(mode)+3, byte(status))
}

func (status GPUStatus) setMatchFlag() GPUStatus {
	return GPUStatus(utils.SetBit(byte(MatchFlag), byte(status), 1))
}

func (status GPUStatus) resetMatchFlag() GPUStatus {
	return GPUStatus(utils.SetBit(byte(MatchFlag), byte(status), 0))
}

func (status GPUStatus) isMatchInterruptSet() bool {
	return utils.IsSet(byte(MatchInterrupt), byte(status))
}