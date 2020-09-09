// CODE ADAPTED FROM SOURCE: https://mikeash.com/pyblog/fluid-simulation-for-dummies.html
package fluid

import (
	"math"
)

type FluidCube struct {
	size 	int		// length of cube sides
	dt 	 	float32 // length of the timestep
	diff 	float32 // diffusion (how fast stuff spreads out in the fluid)
	visc 	float32 // viscosity (how thick the fluid is)

	s 		[]float32 // scratch space density array
	density []float32 // density array

	Vx 		[]float32 // velocity array
	Vy 		[]float32 // velocity array

	Vx0 	[]float32 // scratch space velocity array (need old values while computing new ones)
	Vy0 	[]float32 // scratch space velocity array (need old values while computing new ones)
}


//
// FluidCube functions
//

func FluidCubeCreate(size int, diffusion, viscosity, dt float32) *FluidCube {
	cube := &FluidCube{}
	N := size

	cube.size = size
	cube.dt = dt
	cube.diff = diffusion
	cube.visc = viscosity

	cube.s = make([]float32, N*N)	
	cube.density = make([]float32, N*N)

	cube.Vx = make([]float32, N*N)
	cube.Vy = make([]float32, N*N)

	cube.Vx0 = make([]float32, N*N)
	cube.Vy0 = make([]float32, N*N)

	return cube
}

func (cube *FluidCube) Step() {
	N 	 	:= cube.size;
	visc 	:= cube.visc
	diff 	:= cube.diff
	dt 		:= cube.dt
	Vx 		:= cube.Vx
	Vy 		:= cube.Vy
	Vx0 	:= cube.Vx0
	Vy0 	:= cube.Vy0
	s 		:= cube.s
	density := cube.density
    
    diffuse(1, Vx0, Vx, visc, dt, 4, N);
    diffuse(2, Vy0, Vy, visc, dt, 4, N);
    
    project(Vx0, Vy0, Vx, Vy, 4, N);
    
    advect(1, Vx, Vx0, Vx0, Vy0, dt, N);
    advect(2, Vy, Vy0, Vx0, Vy0, dt, N);
    
    project(Vx, Vy, Vx0, Vy0, 4, N);
    
    diffuse(0, s, density, diff, dt, 4, N);
    advect(0, density, s, Vx, Vy, dt, N);
}

func (cube *FluidCube) AddDensity(x, y int, amount float32) {
	N := cube.size
	cube.density[ix(x, y, N)] += amount
}

func (cube *FluidCube) AddVelocity(x, y int, amountX, amountY float32) {
	N := cube.size
	index := ix(x, y, N)

	cube.Vx[index] += amountX
	cube.Vy[index] += amountY
}

func (cube *FluidCube) Density(x, y int) float32 {
	N := cube.size
	return cube.density[ix(x, y, N)]
}

func (cube *FluidCube) Velocity(x, y int) (float32, float32) {
	N := cube.size
	return cube.Vx[ix(x, y, N)], cube.Vy[ix(x, y, N)]
}


//
// Helper functions
//

func ix(x, y, N int) int {
	return x + y * N
}

func set_bnd(b int, x []float32, N int) {
	for j:=1; j<N-1; j++ {
		if b == 2 {
			x[ix(j, 0, N)] = -x[ix(j, 1, N)]
			x[ix(j, N-1, N)] = -x[ix(j, N-2, N)]
		} else {
			x[ix(j, 0, N)] = x[ix(j, 1, N)]
			x[ix(j, N-1, N)] = x[ix(j, N-2, N)]
		}
	}

	for k:=1; k<N-1; k++ {
		if b == 1 {
			x[ix(0, k, N)] = -x[ix(1, k, N)]
			x[ix(N-1, k, N)] = -x[ix(N-2, k, N)]
		} else {
			x[ix(0, k, N)] = x[ix(1, k, N)]
			x[ix(N-1, k, N)] = x[ix(N-2, k, N)]
		}
	}

	x[ix(0, 0, N)] 	   = 0.5 * (x[ix(1, 0, N)] + x[ix(0, 1, N)]);
	x[ix(0, N-1, N)]   = 0.5 * (x[ix(1, N-1, N)] + x[ix(0, N-2, N)]);
	x[ix(N-1, 0, N)]   = 0.5 * (x[ix(N-2, 0, N)] + x[ix(N-1, 1, N)]);
	x[ix(N-1, N-1, N)] = 0.5 * (x[ix(N-2, N-1, N)] + x[ix(N-1, N-2, N)]);
}

func lin_solve(b int, x, x0 []float32, a, c float32, iter, N int) {
	cRecip := 1.0/c
	for k:=0; k<iter; k++ {
		for m:= 1; m<N-1; m++ {
			for j:=1; j<N-1; j++ {
				x[ix(j, m, N)] =
					(x0[ix(j, m, N)] +
						a * (x[ix(j+1, m, N)] +
							 x[ix(j-1, m, N)] +
							 x[ix(j, m+1, N)] +
							 x[ix(j, m-1, N)])) * cRecip
			}
		}
		set_bnd(b, x, N)
	}
}

func diffuse(b int, x, x0 []float32, diff, dt float32, iter, N int) {
	a := dt * diff * float32(N-2) * float32(N-2)
	lin_solve(b, x, x0, a, 1 + 6 * a, iter, N)
}

func advect(b int, d, d0, velocX, velocY []float32, dt float32, N int) {
	var i0, i1, j0, j1 float32

	dtx := dt * float32(N-2)
	dty := dt * float32(N-2)

	var s0, s1, t0, t1 float32
	var tmp1, tmp2, x, y float32

	Nfloat := float32(N)
	var ifloat, jfloat float32
	var i, j int

	for j, jfloat = 1, 1.0; j<N-1; j, jfloat = j+1, jfloat+1 {
		for i, ifloat = 1, 1.0; i<N-1; i, ifloat = i+1, ifloat+1 {
			tmp1 = dtx * velocX[ix(i, j, N)]
			tmp2 = dty * velocY[ix(i, j, N)]
			x = ifloat - tmp1
			y = jfloat - tmp2

			if x < 0.5 { x = 0.5 }
			if x > Nfloat + 0.5 { x = Nfloat + 0.5 }
			i0 = floorf(x)
			i1 = i0 + 1.0
			if y < 0.5 { y = 0.5 }
			if y > Nfloat + 0.5 { y = Nfloat + 0.5 }
			j0 = floorf(y)
			j1 = j0 + 1.0

			s1 = x - i0
			s0 = 1.0 - s1
			t1 = y - j0
			t0 = 1.0 - t1

			i0i := int(i0)
			i1i := int(i1)
			j0i := int(j0)
			j1i := int(j1)

			d[ix(i, j, N)] =
				s0 * (t0 * d0[ix(i0i, j0i, N)] + t1 * d0[ix(i0i, j1i, N)]) +
				s1 * (t0 * d0[ix(i1i, j0i, N)] + t1 * d0[ix(i1i, j1i, N)])
		}
	}

	set_bnd(b, d, N)
}

func project(velocX, velocY, p, div []float32, iter, N int) {
	for j:= 1; j<N-1; j++ {
		for i:=1; i<N-1; i++ {
			div[ix(i, j, N)] = -0.5*(
				velocX[ix(i+1, j, N)] -
				velocX[ix(i-1, j, N)] +
				velocY[ix(i, j+1, N)] -
				velocY[ix(i, j-1, N)])/float32(N)
			p[ix(i, j, N)] = 0
		}
	}

	set_bnd(0, div, N)
	set_bnd(0, p, N)
	lin_solve(0, p, div, 1, 6, iter, N)

	for j:= 1; j<N-1; j++ {
		for i:=1; i<N-1; i++ {
			velocX[ix(i, j, N)] -= 0.5 * (p[ix(i+1, j, N)] - p[ix(i-1, j, N)]) * float32(N)
			velocY[ix(i, j, N)] -= 0.5 * (p[ix(i, j+1, N)] - p[ix(i, j-1, N)]) * float32(N)
		}
	}

	set_bnd(1, velocX, N)
	set_bnd(2, velocY, N)
}

func floorf(x float32) float32 {
	return float32(math.Floor(float64(x)))
}