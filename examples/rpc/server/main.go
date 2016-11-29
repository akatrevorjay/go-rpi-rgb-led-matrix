package main

import (
	"flag"

	"github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/mcuadros/go-rpi-rgb-led-matrix/rpc"
)

var (
	rows       = flag.Int("led-rows", 32, "number of rows supported")
	chain      = flag.Int("led-chain", 6, "number of displays daisy-chained")
	parallel   = flag.Int("led-parallel", 1, "number of daisy-chained panels")
	brightness = flag.Int("brightness", 10, "brightness (0-100)")
)

func main() {
	config := &rgbmatrix.DefaultConfig
	config.Rows = *rows
	config.ChainLength = *chain
	config.Brightness = *brightness
	config.Parallel = *parallel

	m, err := rgbmatrix.NewRGBLedMatrix(config)
	fatal(err)

	rpc.Serve(m)
}

func init() {
	flag.Parse()
}

func fatal(err error) {
	if err != nil {
		panic(err)
	}
}
