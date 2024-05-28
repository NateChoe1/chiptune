// chiptune - a chiptune song generator
// Copyright (C) 2024 Nate Choe
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"os"
	"fmt"
	"github.com/go-audio/wav"
	"github.com/go-audio/audio"
)

func main() {
	fmt.Println("chiptune - a chiptune song generator")
	fmt.Println("Copyright (C) 2024 Nate Choe")
	fmt.Println("")
	fmt.Println("This program is free software: you can redistribute it and/or modify")
	fmt.Println("it under the terms of the GNU General Public License as published by")
	fmt.Println("the Free Software Foundation, either version 3 of the License, or")
	fmt.Println("(at your option) any later version.")
	fmt.Println("")
	fmt.Println("This program is distributed in the hope that it will be useful,")
	fmt.Println("but WITHOUT ANY WARRANTY; without even the implied warranty of")
	fmt.Println("MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the")
	fmt.Println("GNU General Public License for more details.")
	fmt.Println("")
	fmt.Println("You should have received a copy of the GNU General Public License")
	fmt.Println("along with this program.  If not, see <https://www.gnu.org/licenses/>.")
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
