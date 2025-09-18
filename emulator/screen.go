package emulator

import "fmt"

type screen struct {
	window [32][64]uint32
}

func (s *screen) reset() {
	for k := range s.window {
		for i := range s.window[k] {
			s.window[k][i] = 0x00000000
		}
	}
}

func (s *screen) draw() {
	// TODO - Draw the screen via SDL
	for k := range s.window {
		fmt.Println(s.window[k])
	}
}
