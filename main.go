package main

import (
	"fmt"

	"github.com/adrichey/go-chip8/emulator"
)

func main() {
	emulator.LoadChip8ROM("./test_opcode.ch8")

	fmt.Println("EXITED")
}
