package main

import "fmt"

func main() {
	// Chip8 has 16 8-bit registers
	registers := map[string]byte{
		"V0": 0,
		"V1": 0,
		"V2": 0,
		"V3": 0,
		"V4": 0,
		"V5": 0,
		"V6": 0,
		"V7": 0,
		"V8": 0,
		"V9": 0,
		"VA": 0,
		"VB": 0,
		"VC": 0,
		"VD": 0,
		"VE": 0,
		"VF": 0,
	}

	// 4k bytes of memory
	memory := [4096]byte{}

	// The Index Register is a special register used to store memory addresses for use in operations
	// It's a 16-bit register because the maximum memory address (0xFFF) is too big for an 8-bit register
	var indexRegister uint16

	// The Program Counter (PC) is a special register that holds the address of the next instruction to execute
	// Again, it's 16 bits because it has to be able to hold the maximum memory address (0xFFF)
	var programCounter uint16

	// 16-level stack used to hold PCs. Can push and pull instructions to it for execution flow
	stack := make([]uint16, 16)

	// The Stack Pointer keeps track of our position in the stack
	var stackPointer byte = 0

	// The CHIP-8 has a simple timer used for timing
	// If the timer value is zero, it stays zero
	// If it is loaded with a value, it will decrement at a rate of 60Hz
	// We will just be decrementing based on clock cycle for this application
	var delayTimer byte

	// Same behavior as the Delay Timer
	var soundTimer byte

	fmt.Println("EXITED")
}
