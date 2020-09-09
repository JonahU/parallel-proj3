package main

import (
	"proj3/fluid"
	"fmt"
	"os"
	"flag"
	"bufio"
	"runtime"
	"sync"
	"encoding/json"
)

const DEFAULT_DELAY     int     = 1
const DEFAULT_DIFFUSION float32 = 100
const DEFAULT_VISCOSITY float32 = 1

type settings struct {
	Size 	  int     `json:"size"`
    Frames 	  uint    `json:"frames"`
    SimType   string  `json:"simType"`
	OutPath   string  `json:"outPath"`
    Diffusion float32 `json:"diffusion"` // Optional
    Viscosity float32 `json:"viscosity"` // Optional
	Delay 	  int     `json:"delay"`	 // Optional
	Repeat	  int	  `json:"repeat"`    // Optional
	FadeOut	  bool	  `json:"fadeOut"`	 // Optional
}

func initializeSettings(s *settings) {
	if s.Delay == 0 { s.Delay = DEFAULT_DELAY }
	if s.Diffusion == 0 { s.Diffusion = DEFAULT_DIFFUSION }
	if s.Viscosity == 0 { s.Viscosity = DEFAULT_VISCOSITY }
}

func sequential() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// read json input from stdin
		var input settings
		err := json.Unmarshal(scanner.Bytes(), &input); if err != nil { panic(err) }

		// initialize uninitialized settings
		initializeSettings(&input)

		// create simulation
		fsGIF := fluid.FluidSimulationGIFCreate(
			input.Size,
			input.Frames,
			input.Delay,
			input.SimType,
			input.Diffusion,
			input.Viscosity,
			input.Repeat,
			input.FadeOut,
			input.OutPath,
			0,
			false,
		)

		// run simulation
		fsGIF.Run()

		// save simulation
		fsGIF.Save()
		fmt.Printf("Saved %s\n", input.OutPath)
	}
}

func main() {
	defer fmt.Println("Done!")

	// setup + read command line flags
	threadCount := flag.Int("p", 0, "how many threads")
	bspMode := flag.Bool("bsp", false, "bulk synchronous parallel mode")
	flag.Parse()

	if *threadCount == 0 {
		// SEQUENTIAL VERSION
		sequential()
		return
	}

	// PARALLEL VERSION
	if *threadCount < 0 {
		// if -p flag is negative, default to number of logical cores
		*threadCount = runtime.NumCPU()
	}

	if *bspMode {
		runtime.GOMAXPROCS(*threadCount + 1) // extra thread for updating the fluid sim
	} else {
		runtime.GOMAXPROCS(*threadCount)
	}

	// create tasks channel
	tasks := make(chan *fluid.Task, *threadCount)

	// create waitgroup
	var wg sync.WaitGroup

	// spawn workers
	for i:=0; i<*threadCount; i++ {
		wg.Add(1)
		go fluid.Worker(tasks, i, *threadCount, &wg, *bspMode)
	}

	// read input and create tasks
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// read json input from stdin
		var input settings
		err := json.Unmarshal(scanner.Bytes(), &input); if err != nil { panic(err) }

		// initialize uninitialized settings
		initializeSettings(&input)

		// create simulation
		fsGIF := fluid.FluidSimulationGIFCreate(
			input.Size,
			input.Frames,
			input.Delay,
			input.SimType,
			input.Diffusion,
			input.Viscosity,
			input.Repeat,
			input.FadeOut,
			input.OutPath,
			*threadCount,
			*bspMode,
		)
		task := fluid.TaskCreate(fsGIF)
		tasks <- task
	}

	// close tasks channel and wait for workers to finish doing work
	close(tasks)
	wg.Wait()	
}