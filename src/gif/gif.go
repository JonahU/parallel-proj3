package gif

import (
	"image"
	"image/gif"
	"image/color"
	"image/color/palette"
	"os"
)

type GIF struct {
	data    *gif.GIF 		  // underlying data
	bounds  image.Rectangle   // gif bounds (assumes all frames are same size) 
	chunks  []image.Rectangle // image split up into NumCPU() independent chunks
	delay   int				  // default delay between frames, measured in 100ths of a second
	frames  uint 			  // how many frames in the gif
	outPath string 			  // where to save the image
}

// Frame points to fields in the GIF.data struct
type Frame struct {
	image  *image.Paletted  // pointer to GIF.data.Image[index]
	delay  *int				// pointer to GIF.data.Delay[index]
	index  int				// index in GIF.data slices
}

func NewGIF(x, y, delay int, frames uint, outPath string, chunkCount int) *GIF {
	images := make([]*image.Paletted, frames)
	delays := make([]int, frames)
	data := &gif.GIF{Image: images, Delay: delays}
	bounds := image.Rectangle{
		Min: image.Point{X:0, Y:0},
		Max: image.Point{X:x, Y:y}}
	chunks := chunk(bounds, chunkCount)
	return &GIF{data, bounds, chunks, delay, frames, outPath}
}


//
// GIF functions
//

func (g *GIF) NewFrame(index int) *Frame {
	image := image.NewPaletted(g.bounds, palette.Plan9)
	g.data.Image[index] = image
	g.data.Delay[index] = g.delay
	return &Frame{image, &g.data.Delay[index], index}
}

func (g *GIF) GetFrame(index int) *Frame {
	return &Frame{g.data.Image[index], &g.data.Delay[index], index}
}

func (g *GIF) Size() image.Point {
	return g.bounds.Max
}

func (g *GIF) Bounds(i int) image.Rectangle {
	return g.chunks[i]
}

func (g *GIF) Frames() uint {
	return g.frames
}

func (g *GIF) OutPath() string {
	return g.outPath
}

// Save saves the image to the given file
func (g *GIF) Save() error {
	outWriter, err := os.Create(g.outPath); if err != nil {
		return err
	}
	defer outWriter.Close()

	// EncodeAll is a serial bottleneck
	err = gif.EncodeAll(outWriter, g.data); if err != nil {
		return err
	}
	return nil
}


//
// Frame functions
//

func (frame *Frame) Set(x, y int, c color.Color) {
	frame.image.Set(x, y, c)
}

func (frame *Frame) SetDelay(delay int) {
	*frame.delay = delay
}


//
// Helper functions
//

// Splits up the image into chunks either vertically or horizontally (depends on
// the image orientation). All the chunks are of equal length except the
// last one which may be slightly larger.
func chunk(bounds image.Rectangle, chunks int) []image.Rectangle {
	if chunks == 0 || chunks == 1 {
		// sequential version
		minMax := []image.Rectangle{bounds}
		return minMax
	}

	// parallel version
	var splitVertical bool
	if bounds.Max.X > bounds.Max.Y { // image is portrait
		splitVertical = true
	}

	var xIncrement, yIncrement, xFinal, yFinal int
	if splitVertical {
		xIncrement = bounds.Max.X / chunks 
		xFinal = bounds.Max.X - xIncrement * (chunks-1)
	} else {
		yIncrement = bounds.Max.Y / chunks 
		yFinal = bounds.Max.Y - yIncrement * (chunks-1)
	}

	var minMax []image.Rectangle
	var rect image.Rectangle
	var minX, minY int
	for i:=0; i<chunks-1; i++ {
		rect.Min =  image.Point{minX, minY}
		if splitVertical {
			minX += xIncrement
			rect.Max =  image.Point{minX, bounds.Max.Y}
		} else {
			minY += yIncrement
			rect.Max = image.Point{bounds.Max.X, minY}
		}
		minMax = append(minMax, rect)
	}

	rect.Min = image.Point{minX, minY}
	if splitVertical {
		rect.Max = image.Point{minX + xFinal, bounds.Max.Y}
	} else {
		rect.Max = image.Point{bounds.Max.X, minY + yFinal}
	}
	minMax = append(minMax, rect)
	return minMax
}