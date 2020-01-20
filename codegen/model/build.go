package model

type Build struct {
	// path to the main.go file
	MainFile string

	// push image after  build
	Push bool

	Image
}
