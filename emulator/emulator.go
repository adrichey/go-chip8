package emulator

import (
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
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

const VIDEO_HEIGHT = 32
const VIDEO_WIDTH = 64
const WINDOW_TITLE = "Chip8 Emulator" // TODO: Add file to this??

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

	/*
		Key Mappings:
		Keypad       Keyboard
		+-+-+-+-+    +-+-+-+-+
		|1|2|3|C|    |1|2|3|4|
		+-+-+-+-+    +-+-+-+-+
		|4|5|6|D|    |Q|W|E|R|
		+-+-+-+-+ => +-+-+-+-+
		|7|8|9|E|    |A|S|D|F|
		+-+-+-+-+    +-+-+-+-+
		|A|0|B|F|    |Z|X|C|V|
		+-+-+-+-+    +-+-+-+-+
	*/
	keypad [16]byte

	// Holds our screen pixels
	pixels [VIDEO_HEIGHT][VIDEO_WIDTH]uint32

	// SDL2 specific properties
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	rect     *sdl.Rect
}

func newChip8() (*chip8, error) {
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

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, err
	}
	defer sdl.Quit() // TODO: May need to move these to a specific "destructor" method

	var winWidth, winHeight int32 = VIDEO_WIDTH * 100, VIDEO_HEIGHT * 100

	window, err := sdl.CreateWindow(WINDOW_TITLE, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return nil, err
	}
	c8.window = window
	defer c8.window.Destroy() // TODO: May need to move these to a specific "destructor" method

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}
	c8.renderer = renderer
	c8.renderer.Clear()
	defer c8.renderer.Destroy() // TODO: May need to move these to a specific "destructor" method

	texture, err := c8.renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_STREAMING, VIDEO_WIDTH, VIDEO_HEIGHT)
	if err != nil {
		return nil, err
	}
	c8.texture = texture
	defer c8.texture.Destroy() // TODO: May need to move these to a specific "destructor" method

	c8.rect = &sdl.Rect{X: 0, Y: 0, W: winWidth, H: winHeight}

	c8.op00E0()

	return &c8, nil
}

func LoadChip8ROM(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	c8, err := newChip8()
	if err != nil {
		return err
	}

	for i := 0; i < len(data); i++ {
		c8.memory[START_ADDRESS+uint(i)] = data[i]
	}

	return nil
}

func (c8 *chip8) processInput() bool {
	quit := false

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			quit = true
		case *sdl.KeyboardEvent:
			var s byte = 0
			if t.Type == sdl.KEYDOWN {
				s = 1
			}

			switch t.Keysym.Sym {
			case sdl.K_ESCAPE:
				if s == 1 {
					quit = true
				}
			case sdl.K_x:
				c8.keypad[0] = s
			case sdl.K_1:
				c8.keypad[1] = s
			case sdl.K_2:
				c8.keypad[2] = s
			case sdl.K_3:
				c8.keypad[3] = s
			case sdl.K_q:
				c8.keypad[4] = s
			case sdl.K_w:
				c8.keypad[5] = s
			case sdl.K_e:
				c8.keypad[6] = s
			case sdl.K_a:
				c8.keypad[7] = s
			case sdl.K_s:
				c8.keypad[8] = s
			case sdl.K_d:
				c8.keypad[9] = s
			case sdl.K_z:
				c8.keypad[0xA] = s
			case sdl.K_c:
				c8.keypad[0xB] = s
			case sdl.K_4:
				c8.keypad[0xC] = s
			case sdl.K_r:
				c8.keypad[0xD] = s
			case sdl.K_f:
				c8.keypad[0xE] = s
			case sdl.K_v:
				c8.keypad[0xF] = s
			}
		}
	}

	return quit
}

/*
When we talk about one cycle of this primitive CPU that we’re emulating, we’re talking about it doing three things:
- Fetch the next instruction in the form of an opcode
- Decode the instruction to determine what operation needs to occur
- Execute the instruction
*/
func (c *chip8) cycle() {
	fmt.Println("MEMORY: ", c.memory)

	// Fetch
	c.opcode = uint16(c.memory[c.programCounter])<<8 | uint16(c.memory[c.programCounter+1]) // TODO: TEST

	// Increment the PC before we execute anything
	c.programCounter += 2

	// Decode and Execute
	switch c.opcode & 0xF000 {
	case 0x0000:
		switch c.opcode & 0x000F {
		case 0x0000:
			c.op00E0()
		case 0x000E:
			c.op00EE()
		}
	case 0x1000:
		c.op1nnn()
	case 0x2000:
		c.op2nnn()
	case 0x3000:
		c.op3xkk()
	case 0x4000:
		c.op4xkk()
	case 0x5000:
		c.op5xy0()
	case 0x6000:
		c.op6xkk()
	case 0x7000:
		c.op7xkk()
	case 0x8000:
		switch c.opcode & 0x000F {
		case 0x0000:
			c.op8xy0()
		case 0x0001:
			c.op8xy1()
		case 0x0002:
			c.op8xy2()
		case 0x0003:
			c.op8xy3()
		case 0x0004:
			c.op8xy4()
		case 0x0005:
			c.op8xy5()
		case 0x0006:
			c.op8xy6()
		case 0x0007:
			c.op8xy7()
		case 0x000E:
			c.op8xyE()
		}
	case 0x9000:
		c.op9xy0()
	case 0xA000:
		c.opAnnn()
	case 0xB000:
		c.opBnnn()
	case 0xC000:
		c.opCxkk()
	case 0xD000:
		c.opDxyn()
	case 0xE000:
		switch c.opcode & 0x000F {
		case 0x0001:
			c.opExA1()
		case 0x000E:
			c.opEx9E()
		}
	case 0xF000:
		switch c.opcode & 0x00FF {
		case 0x0007:
			c.opFx07()
		case 0x000A:
			c.opFx0A()
		case 0x0015:
			c.opFx15()
		case 0x0018:
			c.opFx18()
		case 0x001E:
			c.opFx1E()
		case 0x0029:
			c.opFx29()
		case 0x0033:
			c.opFx33()
		case 0x0055:
			c.opFx55()
		case 0x0065:
			c.opFx65()
		}
	default:
		log.Fatal("cannot interpret instruction:", c.opcode)
	}

	// Decrement the delay timer if it's been set
	if c.delayTimer > 0 {
		c.delayTimer -= 1
	}

	// Decrement the sound timer if it's been set
	if c.soundTimer > 0 {
		c.soundTimer -= 1
	}
}

// Update the display
func (c8 *chip8) update() {
	videoPitch := len(c8.pixels[0]) * VIDEO_WIDTH

	// TODO: May need to change the following call: https://github.com/veandco/go-sdl2/blob/7f43f67a3a12d53b3d69f142b9bb67678081313a/sdl/render.go#L575
	c8.texture.Update(c8.rect, unsafe.Pointer(&c8.pixels), videoPitch)
	c8.renderer.Clear()
	c8.renderer.Copy(c8.texture, nil, nil)
	c8.renderer.Present()
}

/*
Our main loop that will call our cycle() receiver method continuously until exit, handle input, and render with SDL.

With each iteration of the loop: input from the keyboard is parsed, a delay is checked to see if enough time has
passed between cycles and a cycle is run if so, and the screen is updated.

Due to the way SDL works, we can simply pass in the video parameter to SDL and it will scale it automatically for
us to the size of our window texture.
*/
func (c8 *chip8) Run() {
	lastCycleTime := time.Now()
	quit := false

	for !quit {
		quit = c8.processInput()

		d := float64(time.Since(lastCycleTime).Milliseconds())

		var cycleDelay float64 = 1 // TODO: May need to convert this to a command line arg if timing feels off between ROMs

		if d > cycleDelay {
			lastCycleTime = time.Now()
			c8.cycle()
			c8.update()
		}
	}
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
	for k := range c8.pixels {
		for i := range c8.pixels[k] {
			c8.pixels[k][i] = 0x00000000
		}
	}
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
				// And screen pixel is also on: collision
				if c8.pixels[vy][vx] == 1 {
					c8.registers[0xF] = 1
				}

				// XOR with the screen pixel with the sprite pixel
				c8.pixels[vy][vx] ^= 1
			}
		}
	}
}

/*
Ex9E - SKP Vx
Skip next instruction if key with the value of Vx is pressed.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) opEx9E() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	key := c8.registers[vx]

	if c8.keypad[key] != 0 {
		c8.programCounter += 2
	}
}

/*
ExA1 - SKNP Vx
Skip next instruction if key with the value of Vx is not pressed.
Since our PC has already been incremented by 2 in Cycle(), we can just increment by 2 again to skip the next instruction.
*/
func (c8 *chip8) opExA1() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	key := c8.registers[vx]

	if c8.keypad[key] == 0 {
		c8.programCounter += 2
	}
}

/*
Fx07 - LD Vx, DT
Set Vx = delay timer value.
*/
func (c8 *chip8) opFx07() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	c8.registers[vx] = c8.delayTimer
}

/*
Fx0A - LD Vx, K
Wait for a key press, store the value of the key in Vx.
The easiest way to "wait" is to decrement the PC by 2 whenever a keypad value is not detected.
This has the effect of running the same instruction repeatedly.
*/
func (c8 *chip8) opFx0A() {
	vx := byte((c8.opcode & 0x0F00) >> 8)

	for k, v := range c8.keypad {
		if v != 0 {
			c8.registers[vx] = byte(k)
			return
		}
	}

	c8.programCounter -= 2
}

/*
Fx15 - LD DT, Vx
Set delay timer = Vx.
*/
func (c8 *chip8) opFx15() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	c8.delayTimer = c8.registers[vx]
}

/*
Fx18 - LD ST, Vx
Set sound timer = Vx.
*/
func (c8 *chip8) opFx18() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	c8.soundTimer = c8.registers[vx]
}

/*
Fx1E - ADD I, Vx
Set I = I + Vx.
*/
func (c8 *chip8) opFx1E() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	c8.indexRegister += uint16(c8.registers[vx])
}

/*
Fx29 - LD F, Vx
Set I = location of sprite for digit Vx.
We know the font characters are located at 0x50, and we know they’re five bytes each, so we can get the address of the first byte of any character by taking an offset from the start address.
*/
func (c8 *chip8) opFx29() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	digit := uint16(c8.registers[vx])

	c8.indexRegister = uint16(FONTSET_START_ADDRESS) + (5 * digit)
}

/*
Fx33 - LD B, Vx
Store BCD representation of Vx in memory locations I, I+1, and I+2.
The interpreter takes the decimal value of Vx, and places the hundreds digit in memory at location in I, the tens digit at location I+1, and the ones digit at location I+2.
We can use the modulus operator to get the right-most digit of a number, and then do a division to remove that digit.
A division by ten will either completely remove the digit (340 / 10 = 34), or result in a float which will be truncated (345 / 10 = 34.5 = 34).
*/
func (c8 *chip8) opFx33() {
	vx := byte((c8.opcode & 0x0F00) >> 8)
	value := c8.registers[vx]

	// Ones-place
	c8.memory[c8.indexRegister+2] = value % 10
	value /= 10

	// Tens-place
	c8.memory[c8.indexRegister+1] = value % 10
	value /= 10

	// Hundreds-place
	c8.memory[c8.indexRegister] = value % 10
}

/*
Fx55 - LD [I], Vx
Store registers V0 through Vx in memory starting at location I.
*/
func (c8 *chip8) opFx55() {
	vx := byte((c8.opcode & 0x0F00) >> 8)

	// TODO: This may cause an overflow or indexing issues. Need to do some thorough testing
	for i := byte(0); i <= vx; i++ {
		c8.memory[byte(c8.indexRegister)+i] = c8.registers[i]
	}
}

/*
Fx65 - LD Vx, [I]
Read registers V0 through Vx from memory starting at location I.
*/
func (c8 *chip8) opFx65() {
	vx := byte((c8.opcode & 0x0F00) >> 8)

	// TODO: This may cause an overflow or indexing issues. Need to do some thorough testing
	for i := byte(0); i <= vx; i++ {
		c8.registers[i] = c8.memory[byte(c8.indexRegister)+i]
	}
}

func randomByte() byte {
	return byte(rand.IntN(255))
}
