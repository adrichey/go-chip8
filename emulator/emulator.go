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

/*
4xkk - SNE Vx, byte
Skip next instruction if Vx != kk.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) op4xkk() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	b := byte(c8.opcode & 0x00FF)

	if c8.registers[vx] != b {
		c8.programCounter += 2
	}
}

/*
5xy0 - SE Vx, Vy
Skip next instruction if Vx = Vy.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) op5xy0() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	if c8.registers[vx] == c8.registers[vy] {
		c8.programCounter += 2
	}
}

/*
6xkk - LD Vx, byte
Set Vx = kk.
*/
func (c8 *chip8) op6xkk() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	b := byte(c8.opcode & 0x00FF)

	c8.registers[vx] = b
}

/*
7xkk - ADD Vx, byte
Set Vx = Vx + kk.
*/
func (c8 *chip8) op7xkk() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	b := byte(c8.opcode & 0x00FF)

	c8.registers[vx] += b
}

/*
8xy0 - LD Vx, Vy
Set Vx = Vy.
*/
func (c8 *chip8) op8xy0() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	c8.registers[vx] = c8.registers[vy]
}

/*
8xy1 - OR Vx, Vy
Set Vx = Vx OR Vy.
*/
func (c8 *chip8) op8xy1() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	c8.registers[vx] |= c8.registers[vy]
}

/*
8xy2 - AND Vx, Vy
Set Vx = Vx AND Vy.
*/
func (c8 *chip8) op8xy2() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	c8.registers[vx] &= c8.registers[vy]
}

/*
8xy3 - XOR Vx, Vy
Set Vx = Vx XOR Vy.
*/
func (c8 *chip8) op8xy3() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	c8.registers[vx] ^= c8.registers[vy]
}

/*
8xy4 - ADD Vx, Vy
Set Vx = Vx + Vy, set VF = carry.
The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,) VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
This is an ADD with an overflow flag. If the sum is greater than what can fit into a byte (255), register VF will be set to 1 as a flag.
*/
func (c8 *chip8) op8xy4() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	sum := uint16(c8.registers[vx]) + uint16(c8.registers[vy])
	if sum > 255 {
		c8.registers[0xF] = 1
	} else {
		c8.registers[0xF] = 0
	}

	c8.registers[vx] = byte(sum & 0xFF)
}

/*
8xy5 - SUB Vx, Vy
Set Vx = Vx - Vy, set VF = NOT borrow.
If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx, and the results stored in Vx.
*/
func (c8 *chip8) op8xy5() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	if c8.registers[vx] > c8.registers[vy] {
		c8.registers[0xF] = 1
	} else {
		c8.registers[0xF] = 0
	}

	c8.registers[vx] -= c8.registers[vy]
}

/*
8xy6 - SHR Vx
Set Vx = Vx SHR 1.
If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0. Then Vx is divided by 2.
A right shift is performed (division by 2), and the least significant bit is saved in Register VF.
*/
func (c8 *chip8) op8xy6() {
	vx := byte((c8.opcode & 0x0F00) >> 8)

	// Save the least significant bit in register VF
	c8.registers[0xF] = c8.registers[vx] & 0x1

	// Division by two using bitwise shift
	c8.registers[vx] >>= 1
}

/*
8xy7 - SUBN Vx, Vy
Set Vx = Vy - Vx, set VF = NOT borrow.
If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy, and the results stored in Vx.
*/
func (c8 *chip8) op8xy7() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	if c8.registers[vy] > c8.registers[vx] {
		c8.registers[0xF] = 1
	} else {
		c8.registers[0xF] = 0
	}

	c8.registers[vx] = c8.registers[vy] - c8.registers[vx]
}

/*
8xyE - SHL Vx {, Vy}
Set Vx = Vx SHL 1.
If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0. Then Vx is multiplied by 2.
A left shift is performed (multiplication by 2), and the most significant bit is saved in Register VF.
*/
func (c8 *chip8) op8xyE() {
	vx := byte((c8.opcode & 0x0F00) >> 8)

	// Save the most significant bit in register VF
	c8.registers[0xF] = (c8.registers[vx] & 0x80) >> 7

	c8.registers[vx] <<= 1
}

/*
9xy0 - SNE Vx, Vy
Skip next instruction if Vx != Vy.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) op9xy0() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)

	if c8.registers[vx] != c8.registers[vy] {
		c8.programCounter += 2
	}
}

/*
Annn - LD I, addr
Set I = nnn.
*/
func (c8 *chip8) opAnnn() {
	address := c8.opcode & 0x0FFF
	c8.indexRegister = address
}

/*
Bnnn - JP V0, addr
Jump to location nnn + V0.
*/
func (c8 *chip8) opBnnn() {
	address := c8.opcode & 0x0FFF
	c8.programCounter = uint16(c8.registers[0]) + address
}

/*
Cxkk - RND Vx, byte
Set Vx = random byte AND kk.
*/
func (c8 *chip8) opCxkk() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	b := byte(c8.opcode & 0x00FF)

	c8.registers[vx] += randomByte() & b
}

/*
Dxyn - DRW Vx, Vy, nibble
Display n-byte sprite starting at memory location I at (Vx, Vy), set VF = collision.
We iterate over the sprite, row by row and column by column. We know there are eight columns because a sprite is guaranteed to be eight pixels wide.
If a sprite pixel is on then there may be a collision with what’s already being displayed, so we check if our screen pixel in the same location is set. If so we must set the VF register to express collision.
Then we can just XOR the screen pixel with 0xFFFFFFFF to essentially XOR it with the sprite pixel (which we now know is on). We can’t XOR directly because the sprite pixel is either 1 or 0 while our video pixel is either 0x00000000 or 0xFFFFFFFF.
TODO: Double check this
*/
func (c8 *chip8) opDxyn() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)
	height := uint16(c8.opcode & 0x000F)
	var pixel uint16

	c8.registers[0xF] = 0
	for row := uint16(0); row < height; row++ {
		pixel = uint16(c8.memory[c8.indexRegister+row])
		for col := uint16(0); col < 8; col++ {
			// If pixel is on...
			if (pixel & (0x80 >> col)) != 0 {
				// And screen pizel is also on: collision
				if c8.scrn.window[vy][vx] == 1 {
					c8.registers[0xF] = 1
				}

				// XOR with the screen pixel with the sprite pixel
				c8.scrn.window[vy][vx] ^= 1
			}
		}
	}
}

/*
Ex9E - SKP Vx
Skip next instruction if key with the value of Vx is pressed.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) opDxyn() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	vy := byte((c8.opcode & 0x00F0) >> 4)
	height := uint16(c8.opcode & 0x000F)
	var pixel uint16

	c8.registers[0xF] = 0
	for row := uint16(0); row < height; row++ {
		pixel = uint16(c8.memory[c8.indexRegister+row])
		for col := uint16(0); col < 8; col++ {
			// If pixel is on...
			if (pixel & (0x80 >> col)) != 0 {
				// And screen pizel is also on: collision
				if c8.scrn.window[vy][vx] == 1 {
					c8.registers[0xF] = 1
				}

				// XOR with the screen pixel with the sprite pixel
				c8.scrn.window[vy][vx] ^= 1
			}
		}
	}
}
