package registers

type Single byte
type Pair byte
type Registers map[Single]byte

const (
	A Single = 0x7
	B        = 0x0
	C        = 0x1
	D        = 0x2
	E        = 0x3
	H        = 0x4
	L        = 0x5
	M        = 0x6 // memory reference through H:L
)

const (
	BC Pair = 0x0
	DE      = 0x1
	HL      = 0x2
	SP      = 0x3
	AF      = 0x4 // are these correct?
)
