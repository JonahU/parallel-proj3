// this package is for internal testing, it doesn't work with go test

package simpletest

import (
	"proj3/gif"
	"proj3/fluid"
	"image/color"
)


const DEFAULT_DIFFUSION float32 = 100
const DEFAULT_VISCOSITY float32 = 1


var RED color.RGBA = color.RGBA{ 255, 0, 0, 255 }
var GREEN color.RGBA = color.RGBA{ 0, 255, 0, 255 }
var BLUE color.RGBA = color.RGBA{ 0, 0, 255, 255 }
var COLORS [3]color.RGBA = [3]color.RGBA{ RED, GREEN, BLUE }


// Simple function to test proj3/gif
// Function to test outputing a simple proj3/gif GIF
// Outputs a rectangle that infinitely cycles through red, green and blue
func RGBRectangle() {
	g := gif.NewGIF(200, 400, 100, 3, "RGBRectangle.gif", 0)
	for i:=0; i<3; i++ {
		f := g.NewFrame(i)
		for x:=0; x<200; x++ {
			for y:=0; y<400; y++ {
				f.Set(x, y, COLORS[i])
			}
		}
	}

	err := g.Save(); if err != nil {
		panic(err)
	}
}

// Simple function to test proj3/fluid
func FluidSim() {
	sg := fluid.FluidSimulationGIFCreate(64, 200, 2, "random", DEFAULT_DIFFUSION, DEFAULT_VISCOSITY, 1, true, "Fluid.gif", 0, false)
	sg.Run()
	sg.Save()
}
