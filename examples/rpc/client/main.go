package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/bradfitz/iter"
	"github.com/disintegration/imaging"
	"github.com/mcuadros/go-rpi-rgb-led-matrix"
	"github.com/mcuadros/go-rpi-rgb-led-matrix/rpc"
)

var (
	img          = flag.String("image", "play.gif", "image path")
	panel_rows   = flag.Int("panel-rows", 32, "stacked are this many rows")
	panel_cols   = flag.Int("panel-cols", 32, "stacked are this many colums")
	stacked_wide = flag.Int("stacked-wide", 2, "Transform to this many stacked wide")
	stacked_high = flag.Int("stacked-high", 3, "Transform to this many stacked high")
	duration     = flag.String("duration", "30s", "Play for this many duration")
)

type Viewport struct {
	image.Rectangle
}

func (v *Viewport) PixelCount() int {
	return v.Dx() * v.Dy()
}

type Panel struct {
	Viewport
}

type Stack struct {
	Viewport
	stacked_wide, stacked_high int
	panel_cols, panel_rows     int
}

func (s *Stack) PanelCount() int {
	return s.stacked_wide * s.stacked_high
}

func SumInt(args ...int) int {
	sum := 0
	for _, v := range args {
		sum = sum + v
	}

	return sum
}

func main() {
	f, err := os.Open(*img)
	assert(err)

	m, err := rpc.NewClient("tcp", "ledpi:1234")
	assert(err)
	defer m.Close()

	width, height := m.Geometry()
	server_viewport := Viewport{image.Rect(0, 0, width, height)}
	log.Printf("Server geometry: %dx%d\n", width, height)

	tk := rgbmatrix.NewToolKit(m)

	panel_rows, panel_cols := *panel_rows, *panel_cols
	stacked_wide, stacked_high := *stacked_wide, *stacked_high

	stack := Stack{
		Viewport{image.Rect(
			0, 0,
			panel_cols*stacked_wide, panel_rows*stacked_high,
		)},
		stacked_wide, stacked_high,
		panel_cols, panel_rows,
	}

	if server_viewport.PixelCount() != stack.PixelCount() {
		log.Fatalf(
			"Panel/Stack size seem to be off as the number of computed pixels does not match"+
				" from %d=%d != to %d=%d @%d:%d",
			server_viewport.Size(), server_viewport.PixelCount(),
			stack.Size(), stack.PixelCount(),
			stacked_wide, stacked_high,
		)
	}

	log.Printf(
		"Transform: %s => %s @%d:%d\n",
		server_viewport.Size(),
		stack.Size(),
		stacked_wide, stacked_high,
	)

	tk.Transform = func(img image.Image) *image.NRGBA {
		// resize to match our geometry
		src := imaging.Resize(img, stack.Dx(), stack.Dy(), imaging.Lanczos)
		dst := imaging.New(server_viewport.Dx(), server_viewport.Dy(), color.Black)
		fmt.Printf("\n--- %s --- %s => %s \n", stack.Size(), src.Rect.Size(), dst.Rect.Size())

		dst_stacked_wide := dst.Rect.Max.X / panel_cols

		var slice image.Image
		var start_pos, end_pos image.Point
		var dst_start_pos, dst_end_pos image.Point
		var rect, dst_rect image.Rectangle

		x, y := 0, 0
		dx, dy := 0, 0

		for idx := range iter.N(stack.PanelCount()) {
			start_pos = image.Pt(x*panel_cols, y*panel_rows)
			dst_start_pos = image.Pt(dx*panel_cols, dy*panel_rows)

			end_pos = image.Pt(start_pos.X+panel_cols, start_pos.Y+panel_rows)
			dst_end_pos = image.Pt(dst_start_pos.X+panel_cols, dst_start_pos.Y+panel_rows)

			rect = image.Rectangle{start_pos, end_pos}
			dst_rect = image.Rectangle{dst_start_pos, dst_end_pos}

			fmt.Printf(" idx=%d (%d,%d) %s \t=> %s\n", idx, x, y, rect, dst_rect)

			slice = src.SubImage(rect)

			dst = imaging.Paste(dst, slice, dst_start_pos)

			x++
			if x == stacked_wide {
				y++
				x = 0
			}

			dx++
			if dx == dst_stacked_wide {
				dy++
				dx = 0
			}
		}

		err = imaging.Save(dst, "omg.png")
		return dst
	}

	close, err := tk.PlayGIF(f)
	assert(err)

	duration, err := time.ParseDuration(*duration)
	assert(err)

	time.Sleep(duration)
	close <- true
}

func init() {
	flag.Parse()
}

func assert(err error) {
	if err != nil {
		panic(err)
	}
}
