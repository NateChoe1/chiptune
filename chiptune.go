package main

import (
	"os"
	"fmt"
	"github.com/go-audio/wav"
	"github.com/go-audio/audio"
)

func main() {
	if len(os.Args) < 3 {
		panic(fmt.Errorf("usage: %v [in.tune] [out.wave]", os.Args[0]))
	}

	tune, err := ReadTune(os.Args[1])
	if err != nil {
		panic(err)
	}

	outFile, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	encoder := wav.NewEncoder(outFile, SAMPLE_RATE, 16, 1, 1)
	defer encoder.Close()

	currLen := 0

	for beat, v := range(tune.tuneData) {
		newData := make([]int, tune.samplesPerBeat)
		for _, note := range(v) {
			line := note.line
			if 'A' <= line && line <= 'Z' {
				line += 'a' - 'A'
				note.instrument.articulate(currLen)
			}
			dynamic := tune.dynamics[beat][line]
			for i := 0; i < len(newData); i += 1 {
				baseVol :=
				note.instrument.getSample(SAMPLE_RATE, currLen + i)
				thisNote := int(baseVol * float64(dynamic.multiplier))
				newData[i] += thisNote
			}
		}
		currLen += tune.samplesPerBeat
		wavSegment := audio.IntBuffer {
			Data: newData,
			Format: audio.FormatMono44100,
			SourceBitDepth: 16,
		}
		err = encoder.Write(&wavSegment)
		if err != nil {
			panic(err)
		}
	}
}
