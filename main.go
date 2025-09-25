package main

import (
	"fmt"
	"log"

	"github.com/adrichey/go-chip8/emulator"
)

func main() {
	err := emulator.LoadChip8ROM("./test_opcode.ch8")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("EXITED")
}
