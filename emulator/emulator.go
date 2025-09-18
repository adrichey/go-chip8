package emulator

import (
	"log"
	"os"
)

/*
There is relatively little register-space (because it's expensive), so a computer needs a large chunk of general
memory dedicated to holding program instructions, long-term data, and short-term data. It references different
locations in that memory using an address.

The CHIP-8 has 4096 bytes of memory, meaning the address space is from 0x000 to 0xFFF.
The address space is segmented into three sections:

	0x000-0x1FF: Originally reserved for the CHIP-8 interpreter, but in our modern emulator we will just never write to or read from that area. Except for...
	0x050-0x0A0: Storage space for the 16 built-in characters (0 through F), which we will need to manually put into our memory because ROMs will be looking for those characters.
	0x200-0xFFF: Instructions from the ROM will be stored starting at 0x200, and anything left after the ROM’s space is free to use.
*/
const START_ADDRESS uint = 0x200
const FONTSET_START_ADDRESS uint = 0x50

type chip8 struct {
	// Chip8 has 16 8-bit registers
	registers [16]byte

	// 4k bytes of memory
	memory [4096]byte

	// The Index Register is a special register used to store memory addresses for use in operations
	// It's a 16-bit register because the maximum memory address (0xFFF) is too big for an 8-bit register
	indexRegister uint16

	// The Program Counter (PC) is a special register that holds the address of the next instruction to execute
	// Again, it's 16 bits because it has to be able to hold the maximum memory address (0xFFF)
	programCounter uint16

	// 16-level stack used to hold PCs. Can push and pull instructions to it for execution flow
	stack [16]uint16

	// The Stack Pointer keeps track of our position in the stack
	stackPointer byte

	// The CHIP-8 has a simple timer used for timing
	// If the timer value is zero, it stays zero
	// If it is loaded with a value, it will decrement at a rate of 60Hz
	// We will just be decrementing based on clock cycle for this application
	delayTimer byte

	// Same behavior as the Delay Timer
	soundTimer byte

	// Store the opcode for instructions
	opcode uint16

	scrn screen
}

func newChip8() chip8 {
	c8 := chip8{}

	for k := range c8.registers {
		c8.registers[k] = 0
	}

	for k := range c8.memory {
		c8.memory[k] = 0
	}

	// Load fontset into memory
	fontset := [80]byte{
		0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0x90, 0x90, 0xF0, 0x10, 0x10, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80, // F
	}

	for k, v := range fontset {
		c8.memory[FONTSET_START_ADDRESS+uint(k)] = v
	}

	for k := range c8.stack {
		c8.stack[k] = 0
	}

	c8.indexRegister = 0
	c8.stackPointer = 0
	c8.delayTimer = 0
	c8.soundTimer = 0
	c8.opcode = 0

	c8.programCounter = uint16(START_ADDRESS)

	c8.scrn.reset()

	return c8
}

func LoadChip8ROM(filepath string) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	c8 := newChip8()

	for i := 0; i < len(data); i++ {
		c8.memory[START_ADDRESS+uint(i)] = data[i]
	}

	c8.scrn.draw()
}

/*
INSTRUCTIONS IMPLEMENTATION

The following section is a set of all instruction operations allowed to us in Chip8.
See this documentation for more details:
https://github.com/mattmikolay/chip-8/wiki/Mastering-CHIP%E2%80%908
https://github.com/mattmikolay/chip-8/wiki/CHIP%E2%80%908-Instruction-Set
*/

/*
00E0: CLS
Clear the display
*/
func (c8 *chip8) op00E0() {
	c8.scrn.reset()
}

/*
00EE: RET
Return from a subroutine
*/
func (c8 *chip8) op00EE() {
	c8.stackPointer -= 1
	c8.programCounter = c8.stack[c8.stackPointer]
}

/*
1nnn: JP addr
Jump to location nnn.
The interpreter sets the program counter to nnn.
A jump doesn’t remember its origin, so no stack interaction required.
*/
func (c8 *chip8) op1nnn() {
	// Use bitwise AND to find our jump location in our memory array
	address := c8.opcode & 0x0FFF
	c8.programCounter = address
}

/*
2nnn - CALL addr
Call subroutine at nnn.
*/
func (c8 *chip8) op2nnn() {
	address := c8.opcode & 0x0FFF
	c8.stack[c8.stackPointer] = c8.programCounter
	c8.stackPointer += 1
	c8.programCounter = address
}

/*
3xkk - SE Vx, byte
Skip next instruction if Vx = kk.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) op3xkk() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	b := byte(c8.opcode & 0x00FF)

	if c8.registers[vx] == b {
		c8.programCounter += 2
	}
}
