package perf

import (
	"os"
	"redis-cluster-manager/vars"
	"runtime"
	"runtime/pprof"
)

func StartCpuProfile() *os.File {
	if vars.CPUProfiler {
		f, err := os.Create("cpu.pprof")
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		return f
	}
	return nil
}

func StopCpuProfile(f *os.File) {
	if f != nil {
		pprof.StopCPUProfile()
		f.Close()
		return
	}
}

func MemProfile() {
	if vars.MEMProfiler {
		f, err := os.Create("mem.pprof")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
	}
}
