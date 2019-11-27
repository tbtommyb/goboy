package registers

type Single byte
type Pair byte
type Registers struct {
	single map[Single]byte
	ioram  [0x100]byte
}

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
	AF      = 0x4 // TODO: are these correct?
)

func Init() *Registers {
	return &Registers{
		ioram:  [0x100]byte{},
		single: make(map[Single]byte),
	}
}

func (r *Registers) Write(register Single, value byte) {
	r.single[register] = value
}

func (r *Registers) Read(register Single) byte {
	return r.single[register]
}

func (r *Registers) WriteIO(address uint16, value byte) {
	r.ioram[address] = value
}

func (r *Registers) ReadIO(address uint16) byte {
	return r.ioram[address]
}
