package fluid

import (
	"image"
	"proj3/gif"
	"math/rand"
	"image/color"
)


const FLOAT32_MIN float32 = 0.0000001


type Simulation struct {
	cube 			*FluidCube		  // fluidCube that handles the actual simulation
	cubePrevState 	*cacheCube		  // prev tick of fluidCube, enables simulaneous writing of prev tick and updating current tick (only used in BSP Mode)
	length  		int				  // how long to simulate for
	simType 		string 			  // simulation type: "random"
	update  		func(*Simulation) // update function that is run on every tick
	fadeOut 		bool   			  // don't add dye for the last 50 ticks
	tick			int				  // current tick
	repeat			int				  // how many times to run the update function every tick
}

type SimulationGIF struct {
	GIF 	*gif.GIF
	sim		*Simulation
}


//
// Simulation functions
//

func FluidSimulationCreate(size, length int, diffusion, viscosity float32, simType string, repeat int, fadeOut bool, bspMode bool) *Simulation {
	f := FluidCubeCreate(size, diffusion, viscosity, FLOAT32_MIN)

	var update func(*Simulation)
	switch simType {
	case "random":
		update = random
	default:
		panic("Unknown simulation type: " + simType)
	}

	if repeat <= 0 {
		repeat = 1
	}

	var prev *cacheCube
	if bspMode {
		prev = cacheCubeCreate(size)
	}

	return &Simulation{f, prev, length, simType, update, fadeOut, 0, repeat}
}

func (sim *Simulation) Run() {
	for i:=0; i<sim.length; i++ {
		// update the fluid cube
		sim.Update()

		// step forward in the simulation
		sim.Step()
	}
}

func (sim *Simulation) CubeStep() {
	sim.cube.Step()
}

func (sim *Simulation) NextTick() {
	sim.tick++
}

func (sim *Simulation) Step() {
	sim.CubeStep()
	sim.NextTick()
}

func (sim *Simulation) Update() {
	// If fadeOut option selected, don't add any dye for the last 50 ticks
	if !(sim.fadeOut && sim.tick > sim.length-50) {

		// scale number of times update is called by repeat amount
		for i:=0; i<sim.repeat; i++ {
			sim.update(sim)
		}
	}
}

// copy FluidCube's density slice values to prevState's density slice
func (sim *Simulation) UpdatePrevState() {
	sim.cubePrevState.SaveState(sim.cube)
}

func random(sim *Simulation) {
	// Generate random coordinates + random density
	randX := int(rand.Int31n(int32(sim.cube.size)-1))
	randY := int(rand.Int31n(int32(sim.cube.size)-1))
	randD := rand.Float32()*200

	// Repeat 4x so effect is more noticeable
	for j:=0; j<4; j++ {
		randXVelocity := rand.Float32()*negative()*2
		randYVelocity := rand.Float32()*negative()*2

		// Add some dye + velocity to a random area of the fluid cube
		sim.cube.AddDensity(randX, randY, randD)
		sim.cube.AddVelocity(randX, randY, randXVelocity, randYVelocity)
				
		// Add some velocity to the center of the fluid cube
		sim.cube.AddVelocity(sim.cube.size/2, sim.cube.size/2, randXVelocity, randYVelocity)
	}
}


//
// SimulationGIF functions
//

func FluidSimulationGIFCreate(size int, frames uint, delay int, simType string, diffusion, viscosity float32, repeat int, fadeOut bool, outPath string, threadCount int, bspMode bool) *SimulationGIF {
	g := gif.NewGIF(size, size, delay, frames, outPath, threadCount)
	s := FluidSimulationCreate(size, int(frames), diffusion, viscosity, simType, repeat, fadeOut, bspMode)
	return &SimulationGIF{g, s}
}

// Lazy initialization of GIF frames for improved performance (in theory)
func (sg *SimulationGIF) InitFrame() *gif.Frame {
	return sg.GIF.NewFrame(sg.sim.tick)
}

func (sg *SimulationGIF) CurrentFrame() *gif.Frame {
	return sg.GIF.GetFrame(sg.sim.tick)
}

func (sg *SimulationGIF) WriteFrame() {
	sg.InitFrame()
	minBounds := image.Point{X:0, Y:0}
	maxBounds := sg.GIF.Size()
	rect := image.Rectangle{ minBounds, maxBounds}
	sg.writeFrameChunk(sg.sim.cube, rect)
}

func (sg *SimulationGIF) writeFrameChunk(cube densityCube, chunk image.Rectangle) {
	for x:=chunk.Min.X; x<chunk.Max.X; x++ {
		for y:=chunk.Min.Y; y<chunk.Max.Y; y++ {
			density := cube.Density(x, y)
			sg.CurrentFrame().Set(x, y, brightness(density))
		}
	}
}

func (sg *SimulationGIF) Run() {
	for i:=0; i<sg.sim.length; i++ {
		// write gif frame
		sg.WriteFrame()

		// update the fluid cube
		sg.sim.Update()

		// step forward in the simulation
		sg.sim.Step()
	}
}

func (sg *SimulationGIF) Save() error {
	return sg.GIF.Save()
}


//
// Helper functions
//

func scale(f float32) uint16 {
	f = f*65535
	if f < 0 {
		return 0
	} else if f > 65535 {
		return 65535
	} else {
		return uint16(f)
	}
}

func brightness(amt float32) color.RGBA64 {
	x := scale(amt)
	return color.RGBA64{x, x, x, x}
}

func negative() float32 {
	rint := rand.Int31n(2) 
	if rint == 0 {
		return -1
	} else {
		return 1
	}
}