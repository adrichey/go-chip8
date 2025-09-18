package main

import (
	"fmt"

	"github.com/adrichey/go-chip8/emulator"
)

func main() {
	// The screen output consists of a monochrome 64x32 grid
	// Setting these to 32-bit ints for SDL compatibility and possible reuse
	screen := emulator.Screen{}
	screen.Draw()

	fmt.Println("EXITED")
}
