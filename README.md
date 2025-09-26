# Overview
This is a quick and dirty Chip8 emulator written in Go to further my knowledge on emulation.

<img width="640" alt="Test ROM Screenshot" src="https://github.com/user-attachments/assets/1b4baf20-7b53-4905-9b3b-a588fba8cd50" />

## How to build this application
- You will need to install SDL2 locally. No optional packages are needed for this application. Follow the Requirements section instructions in the [go-sdl2 package README](https://github.com/veandco/go-sdl2?tab=readme-ov-file#requirements).
- Once installed, clone this repo and run `go build`
- If on Windows, you'll need to copy the runtime SDL2.dll into the repo directory as well as indicated in the [go-sdl2 package README](https://github.com/veandco/go-sdl2?tab=readme-ov-file#requirements).

## How to use this application

Flags
- `-f`: Path to a Chip8 ROM file
- `-d`: Specifies the cycle delay to control the emulator cycle and update speeds (optional, default 5)
- `-s`: Specifies the video scale for the emulator; Chip8 is 64x32 so 10 == 640x320 (optional, default 10)

### Example
- Linux: `./go-chip8 -f ./roms/1-chip8-logo.ch8`
- Windows: `.\go-chip8.exe -f .\roms\1-chip8-logo.ch8`

### Example with optional args
- Linux: `./go-chip8 -f ./roms/1-chip8-logo.ch8 -d 10 -s 20`
- Windows: `.\go-chip8.exe -f .\roms\1-chip8-logo.ch8 -d 10 -s 20`

## Special Thanks
- Austin Morlan for his excellent write-up: [article](https://austinmorlan.com/posts/chip8_emulator/)
- Tim Franssen for his collection of test ROMs: [repo](https://github.com/Timendus/chip8-test-suite)
- Zophar's Domain for Pong which is now in the public domain: [website](https://www.zophar.net/pdroms/chip8.html) | [mirror](https://archive.org/details/Chip-8RomsThatAreInThePublicDomain)

## TODO
- Implement sound functionality
- Implement 5x33 properly to handle BCDs - Every test I have tried fails
