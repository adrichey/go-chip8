package main

import (
	"fmt"
	"log"

	"github.com/adrichey/go-chip8/emulator"
)

func main() {
	c8, err := emulator.NewChip8()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = c8.LoadChip8ROM("./test_opcode.ch8")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer c8.Destroy()

	c8.Run()

	fmt.Println("EXITED")
}
