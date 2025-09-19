package emulator

import "fmt"

const VIDEO_HEIGHT = 32
const VIDEO_WIDTH = 64

type screen struct {
	window [VIDEO_HEIGHT][VIDEO_WIDTH]uint32
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
