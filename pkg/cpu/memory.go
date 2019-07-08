package cpu

type Memory []byte

const ProgramStartAddress = 0x150

// TODO: use pointer?
func (m Memory) Load(start int, data []byte) {
	for i := 0; i < len(data); i++ {
		m[start+i] = data[i]
	}
}

func InitMemory() Memory {
	return make(Memory, 8000)
}
