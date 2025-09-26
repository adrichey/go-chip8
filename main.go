package main

import (
	"log"

	"github.com/adrichey/go-chip8/emulator"
)

func main() {
	c8, err := emulator.NewChip8()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = c8.LoadChip8ROM("./roms/pong.ch8")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer c8.Destroy()

	c8.Run()
}
