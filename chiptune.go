package main

import (
	"os"
	"fmt"
)

func main() {
	if len(os.Args) < 3 {
		panic(fmt.Errorf("usage: %v [in.tune] [out.wave]", os.Args[0]))
	}

	tune, err := ReadTune(os.Args[1])
	if err != nil {
		panic(err)
	}

	for _, v := range(tune.dynamics) {
		fmt.Println(v)
	}
}
