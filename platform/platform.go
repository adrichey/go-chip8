package platform

import "fmt"

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
