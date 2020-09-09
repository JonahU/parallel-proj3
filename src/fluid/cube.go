package fluid

type densityCube interface {
	Density(x int, y int) float32
}

type cacheCube struct {
	size int
	density []float32
}

func cacheCubeCreate(size int) *cacheCube {
	density := make([]float32, size*size)
	return &cacheCube{size, density}
}

func (cache *cacheCube) SaveState(cube densityCube) {
	for y:=0; y<cache.size; y++ {
		for x:=0; x<cache.size; x++ {
			index := ix(x, y, cache.size)
			cache.density[index] = cube.Density(x, y) 
		}
	}
}

func (cache *cacheCube) Density(x, y int) float32 {
	return cache.density[ix(x, y, cache.size)]
}