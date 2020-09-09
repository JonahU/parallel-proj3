package fluid

import (
	"sync"
	"image"
	"fmt"
	"proj3/barrier"
)

type Task struct {
	sg *SimulationGIF
	id int
}

type writeTask struct {
	sg	   *SimulationGIF
	bounds image.Rectangle
	cube   densityCube 		// FluidCube (Regular mode) or cacheCube (BSP mode)
	parent int
	id	   int
}

type simTask struct {
	sim	*Simulation
}

var taskId int
func TaskCreate(sg *SimulationGIF) *Task {
	taskId++
	return &Task{sg, taskId}
}

func Worker(tasks <-chan *Task, workerId, threadCount int,  mainWg *sync.WaitGroup, bspMode bool) {
	// create write tasks channel
	writeTasks := make(chan *writeTask, threadCount)
	frameDone := make(chan struct{}, threadCount)

	// create write WaitGroup
	var writeWg sync.WaitGroup

	// spawn writer workers
	for i:=0; i<threadCount; i++ {
		writeWg.Add(1)
		go writerWorker(writeTasks, frameDone, i, &writeWg)
	}

	for {
		task, more := <-tasks; if !more {
			// close the writeTasks channel
			close(writeTasks)

			// wait for writer workers
			writeWg.Wait()

			// signal to main goroutine worker is done
			mainWg.Done()
			return
		}

		if bspMode {
			doWorkBSP(task, threadCount, writeTasks, frameDone, &writeWg)
		} else {
			doWork(task, threadCount, writeTasks, frameDone)
		}


		// save the image
		task.sg.Save()
		fmt.Printf("Saved %s\n", task.sg.GIF.OutPath())
	}
}

func doWork(task *Task, threadCount int, writeTasks chan<- *writeTask, frameDone <-chan struct{}) {
	// cycle through GIF frames
	for i:=0; i<int(task.sg.GIF.Frames()); i++ {

		// initalize the current frame
		task.sg.InitFrame()

		// push image chunks to write to writeTasks channel
		// image chunks will be handled by writerWorkers
		for i:=0; i<threadCount; i++ {
			writeTasks <- &writeTask{task.sg, task.sg.GIF.Bounds(i), task.sg.sim.cube, i, task.id}
		}

		// wait for writers to finish writing current frame
		for i:=0; i<threadCount; i++ {
			<-frameDone
		}

		// update the fluid cube
		task.sg.sim.Update()

		// step forward in the simulation
		task.sg.sim.Step()
	}
}

// does the same as doWork() but handles writing the previous GIF frame + updating the simulation in parallel
// uses a barrier for synchronization
func doWorkBSP(task *Task, threadCount int, writeTasks chan<- *writeTask, frameDone <-chan struct{}, wg *sync.WaitGroup) {
	// initialize simulation worker channels
	simWorkerStart := make(chan bool)
	simWorkerDone := make(chan struct{})

	// add 1 to wait group so main worker doesn't exit before sim worker is done
	wg.Add(1)

	// spawn simulation worker
	go simWorker(task.sg.sim, simWorkerStart, simWorkerDone, wg)

	// create barrier
	bar := barrier.BarrierCreate(threadCount+1, frameDone, simWorkerDone)

	// cycle through GIF frames
	for i:=0; i<int(task.sg.GIF.Frames()); i++ {

		// initalize the current frame
		task.sg.InitFrame()

		// push image chunks to write to writeTasks channel
		// image chunks will be handled by writerWorkers
		for i:=0; i<threadCount; i++ {
			writeTasks <- &writeTask{task.sg, task.sg.GIF.Bounds(i), task.sg.sim.cubePrevState, i, task.id}
		}
		
		// tell sim worker to start work
		simWorkerStart<-true

		// synchronize writers + simulation workers via barrier
		bar.Wait()

		// update prevState density values
		task.sg.sim.UpdatePrevState()

		// increment simulation tick
		task.sg.sim.NextTick()
	}

	// tell sim worker to stop waiting for more work
	close(simWorkerStart)
}

func writerWorker(tasks <-chan *writeTask, done chan<- struct{}, writerWorkerId int, workerWg *sync.WaitGroup) {
	for {
		task, more := <-tasks; if !more {
			workerWg.Done()
			return
		}
		task.sg.writeFrameChunk(task.cube, task.bounds)
		done <- struct{}{}
	}
}

func simWorker(sim *Simulation, start <-chan bool, done chan<- struct{}, workerWg *sync.WaitGroup) {
	for {
		_, more := <-start; if !more {
			workerWg.Done()
			return
		}

		sim.Update() 	 // update the fluid cube
		sim.CubeStep()	 // step forward in the fluid cube

		// tell worker current simulation tick is done computing
		done <- struct{}{}
	}
}

