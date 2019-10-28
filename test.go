package main

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type CPU struct {
	pc    int
	mutex sync.Mutex
}

func (cpu *CPU) handleInterrupts(interrupts <-chan int) {
	for interrupt := range interrupts {
		cpu.mutex.Lock()
		fmt.Printf("handled at: %d\n", cpu.pc)
		cpu.pc = interrupt
		fmt.Printf("New pc: %d\n", cpu.pc)
		cpu.mutex.Unlock()
	}
}

func (cpu *CPU) run(interrupts chan<- int) {
	time.Sleep(1000000)
	cpu.mutex.Lock()
	cpu.pc++
	if (cpu.pc % 10000) == 0 {
		fmt.Printf("Sync interrupt at: %d. ", cpu.pc)
		interrupts <- 0
	}
	cpu.mutex.Unlock()
}

func (cpu *CPU) getUserInput(interrupts chan<- int) {
	var input string
	for {
		fmt.Print("Enter new interrupt: ")
		fmt.Scanln(&input)
		interrupt, _ := strconv.Atoi(input)
		cpu.mutex.Lock()
		fmt.Printf("Async interrupt at %d. ", cpu.pc)
		cpu.mutex.Unlock()
		interrupts <- interrupt
	}
}

func main() {
	runtime.GOMAXPROCS(4)
	cpu := CPU{pc: 1}
	interrupts := make(chan int)
	go cpu.handleInterrupts(interrupts)
	go cpu.getUserInput(interrupts)
	for {
		cpu.run(interrupts)
	}
	close(interrupts)
}
