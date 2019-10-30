package cpu

import "github.com/tbtommyb/goboy/pkg/utils"

type GPUStatus byte
type GPUControl byte
type GPUControlFlag byte
type GPUStatusFlag byte

const (
	LCDDisplayEnable           GPUControlFlag = 0x80
	WindowTileMapDisplaySelect                = 0x40
	WindowDisplayEnable                       = 0x20
	DataSelect                                = 0x10
	BGTileMapDisplaySelect                    = 0x8
	SpriteSize                                = 0x4
	SpriteEnable                              = 0x2
	WindowDisplayPriority                     = 0x1
)

const (
	MatchFlag              GPUStatusFlag = 2
	AccessEnabledInterrupt               = 3
	VBlankInterrupt                      = 4
	OAMInterrupt                         = 5
	MatchInterrupt                       = 6
)

const LCDCAddress = 0xFF40
const STATAddress = 0xFF41
const ScrollYAddress = 0xFF42
const ScrollXAddress = 0xFF43
const LYAddress = 0xFF44
const LYCAddress = 0xFF45
const BGPAddress = 0xFF47
const OBP0Address = 0xFF48
const OBP1Address = 0xFF49
const WindowYAddress = 0xFF4A
const WindowXAddress = 0xFF4B

// LCD Control

func (cpu *CPU) getGPUControl() GPUControl {
	return GPUControl(cpu.memory.get(LCDCAddress))
}

func (cpu *CPU) setGPUControl(control GPUControl) {
	cpu.memory.set(LCDCAddress, byte(control)|1)
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

func (control GPUControl) useHighWindowAddress() bool {
	return control.isSet(WindowTileMapDisplaySelect)
}

func (control GPUControl) useHighBGDataAddress() bool {
	return !control.isSet(DataSelect)
}

func (control GPUControl) useHighBGStartAddress() bool {
	return control.isSet(BGTileMapDisplaySelect)
}

func (control GPUControl) isSet(flag GPUControlFlag) bool {
	return (byte(control) & byte(flag)) > 0
}

// LCD Status

func (cpu *CPU) getGPUStatus() GPUStatus {
	return GPUStatus(cpu.memory.get(STATAddress))
}

func (cpu *CPU) setGPUStatus(status GPUStatus) {
	cpu.memory.set(STATAddress, byte(status))
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

func (cpu *CPU) getScrollY() byte {
	return cpu.memory.get(ScrollYAddress)
}

func (cpu *CPU) setScrollY(value byte) {
	cpu.memory.set(ScrollYAddress, value)
}

func (cpu *CPU) getScrollX() byte {
	return cpu.memory.get(ScrollXAddress)
}

func (cpu *CPU) setScrollX(value byte) {
	cpu.memory.set(ScrollXAddress, value)
}

func (cpu *CPU) getWindowY() byte {
	return cpu.memory.get(WindowYAddress)
}

func (cpu *CPU) getWindowX() byte {
	return cpu.memory.get(WindowXAddress)
}

func (cpu *CPU) getLY() byte {
	return cpu.memory.get(LYAddress)
}

func (cpu *CPU) setLY(value byte) {
	cpu.memory.ioram[LYAddress-0xFF00] = value
}

func (cpu *CPU) getLYC() byte {
	return cpu.memory.get(LYCAddress)
}

func (cpu *CPU) getBGP() byte {
	return cpu.memory.get(BGPAddress)
}

func (cpu *CPU) setBGP(value byte) {
	cpu.memory.set(BGPAddress, value)
}

func (cpu *CPU) setOBP0(value byte) {
	cpu.memory.set(OBP0Address, value)
}

func (cpu *CPU) getOBP0() byte {
	return cpu.memory.get(OBP0Address)
}

func (cpu *CPU) setOBP1(value byte) {
	cpu.memory.set(OBP1Address, value)
}

func (cpu *CPU) getOBP1() byte {
	return cpu.memory.get(OBP1Address)
}
