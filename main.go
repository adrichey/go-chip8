package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/adrichey/go-chip8/emulator"
)

var help bool
var romFile string
var cycleDelay float64
var videoScale int

func init() {
	flag.BoolVar(&help, "help", false, "Help")
	flag.StringVar(&romFile, "f", "", "Path to a Chip8 ROM file")
	flag.Float64Var(&cycleDelay, "d", 5, "Specifies the cycle delay to control the emulator cycle and update speeds (optional, default 5)")
	flag.IntVar(&videoScale, "s", 10, "Specifies the video scale for the emulator; Chip8 is 64x32 so 10 == 640x320 (optional, default 10)")

	flag.Parse()
}

func main() {
	if help {
		displayHelp()
		return
	}

	c8, err := emulator.NewChip8(videoScale, cycleDelay)
	if err != nil {
		log.Fatal(err)
		return
	}

	err = c8.LoadChip8ROM(romFile)
	if err != nil {
		log.Fatal("Error loading ROM file - ", err)
		return
	}
	defer c8.Destroy()

	c8.Run()
}

func displayHelp() {
	fmt.Println("How to use this script:")
	fmt.Println("-f: Path to a Chip8 ROM file")
	fmt.Println("-d: Specifies the cycle delay to control the emulator cycle and update speeds (optional, default 5)")
	fmt.Println("-s: Specifies the video scale for the emulator; Chip8 is 64x32 so 10 == 640x320 (optional, default 10)")
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println("./go-chip8 -f ./roms/1-chip8-logo.ch8")
	fmt.Println()
	fmt.Println("Example with optional args:")
	fmt.Println("./go-chip8 -f ./roms/1-chip8-logo.ch8 -d 10 -s 20")
	fmt.Println()
}
