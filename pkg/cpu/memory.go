package cpu

type Memory []byte

const ProgramStartAddress = 0x150
const StackStartAddress = 0xFF80

// TODO: use pointer?
func (m Memory) Load(start int, data []byte) {
	for i := 0; i < len(data); i++ {
		m[start+i] = data[i]
	}
}

func (m Memory) Set(address uint16, value byte) byte {
	m[address] = value
	return value
}

func (m Memory) Get(address uint16) byte {
	return m[address]
}

func InitMemory() Memory {
	return make(Memory, 0xFFFF)
}
