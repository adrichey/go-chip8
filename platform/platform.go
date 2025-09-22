package platform

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const VIDEO_HEIGHT = 32
const VIDEO_WIDTH = 64

type Screen struct {
	Window [VIDEO_HEIGHT][VIDEO_WIDTH]uint32
}

func (s *Screen) Reset() {
	for k := range s.Window {
		for i := range s.Window[k] {
			s.Window[k][i] = 0x00000000
		}
	}
}

func (s *Screen) Draw() {
	// TODO - Draw the screen via SDL
	for k := range s.Window {
		fmt.Println(s.Window[k])
	}
}

type Platform struct {
	Keypad [16]byte
	Screen Screen
}

func (p *Platform) ProcessInput() bool {
	quit := false

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := ebent.(type) {
		case *sdl.QuitEvent:
			quit = true
		case sdl.KeyboardEvent:
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
				p.Keypad[0] = s
			case sdl.K_1:
				p.Keypad[1] = s
			case sdl.K_2:
				p.Keypad[2] = s
			case sdl.K_3:
				p.Keypad[3] = s
			case sdl.K_q:
				p.Keypad[4] = s
			case sdl.K_w:
				p.Keypad[5] = s
			case sdl.K_e:
				p.Keypad[6] = s
			case sdl.K_a:
				p.Keypad[7] = s
			case sdl.K_s:
				p.Keypad[8] = s
			case sdl.K_d:
				p.Keypad[9] = s
			case sdl.K_z:
				p.Keypad[0xA] = s
			case sdl.K_c:
				p.Keypad[0xB] = s
			case sdl.K_4:
				p.Keypad[0xC] = s
			case sdl.K_r:
				p.Keypad[0xD] = s
			case sdl.K_f:
				p.Keypad[0xE] = s
			case sdl.K_v:
				p.Keypad[0xF] = s
			}
		}
	}

	return quit
}
