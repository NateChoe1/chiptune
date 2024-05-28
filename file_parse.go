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
	"io"
	"os"
	"fmt"
	"math"
	"bufio"
	"strconv"
	"strings"
	"math/rand"
)

type GenericPeriodic func(cycleLen, depth int) float64

func ReadTune(file string) (*Tune, error) {
	linenum := 0
	ret := Tune{}

	f, err := os.Open(os.Args[1])
	if err != nil {
		return nil, err
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	// get the tempo
	bpmString, err := ReadLine(reader, &linenum)
	if err != nil {
		return nil, err
	}
	bpm, err := strconv.Atoi(strings.TrimSpace(string(bpmString)))
	if err != nil {
		return nil, err
	}
	if bpm <= 0 {
		return nil, fmt.Errorf("invalid bpm: %v", bpm)
	}
	samplesPerBeat := 60 * SAMPLE_RATE / bpm
	ret.tempo = bpm
	ret.samplesPerBeat = samplesPerBeat

	// get the lines
	linesString, err := ReadLine(reader, &linenum)
	if err != nil {
		return nil, err
	}
	for _, v := range(linesString) {
		// only use lowercase letters as lines
		if v < 'a' || v > 'z' {
			continue
		}
		ret.lines = append(ret.lines, v)
	}

	// read the instrument specification
	instrumentString, err := ReadLine(reader, &linenum)
	if err != nil {
		return nil, err
	}
	instrumentStrings := strings.Split(string(instrumentString), ",")
	var instrumentList []Instrument
	for _, v := range(instrumentStrings) {
		var err error
		instrumentList, err = ParseInstrument(instrumentList, strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}
	}

	// read the rest of the data
	for {
		line, err := ReadLine(reader, &linenum)
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		err = AddBeat(&ret, instrumentList, line, linenum)
		if err != nil {
			return nil, err
		}
	}

	InterpolateDynamics(&ret)

	return &ret, nil
}

func ReadLine(scanner *bufio.Reader, linenum *int) ([]byte, error) {
	line, err := scanner.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	for i := range(line) {
		if line[i] == ';' || line[i] == '\n' {
			line = line[0:i]
			break
		}
	}
	*linenum += 1
	return line, nil
}

func ParseInstrument(instrumentList []Instrument, instrument string) ([]Instrument, error) {
	parts := strings.Split(instrument, "|")
	switch parts[0] {
	case "noise":
		switch parts[1] {
		case "white":
			return append(instrumentList, MakeWhiteNoise()), nil
		}
		break
	case "tri":
		return MakeGenericPeriodic(instrumentList, parts[1:], GenericTriangle)
	case "saw":
		return MakeGenericPeriodic(instrumentList, parts[1:], GenericSawtooth)
	case "square":
		return MakeGenericPeriodic(instrumentList, parts[1:], GenericSquare)
	}
	return nil, fmt.Errorf("invalid instrument: %v", instrument)
}

func MakeWhiteNoise() Instrument {
	lastArticulation := -ARTICULATION_LEN
	articulate := func(time int) {
		lastArticulation = time
	}
	getSample := func(sampleRate int, time int) float64 {
		if time - lastArticulation < ARTICULATION_LEN {
			return 0.5
		}
		return rand.Float64()
	}
	return Instrument {
		articulate: articulate,
		getSample: getSample,
	}
}

func GenericTriangle(cycleLen, depth int) float64 {
	if depth > cycleLen / 2 {
		depth = cycleLen - depth
	}
	return float64(depth) / float64(cycleLen) * 2
}

func GenericSawtooth(cycleLen, depth int) float64 {
	return float64(depth) / float64(cycleLen)
}

func GenericSquare(cycleLen, depth int) float64 {
	if depth < cycleLen / 2 {
		return 0
	}
	return 1
}

func MakeGenericPeriodic(instrumentList []Instrument,
		spec []string, timbre GenericPeriodic) ([]Instrument, error) {
	if len(spec) < 2{
		return nil, fmt.Errorf("invalid periodic spec: %v", spec)
	}
	startNote, err := ParseNote(spec[1])
	if err != nil {
		return nil, err
	}

	var scale []int
	switch (spec[0]) {
	case "one":
		instrumentList = append(instrumentList, MakePeriodicNote(startNote, timbre))
		return instrumentList, nil
	case "minor":
		scale = []int{0, 2, 3, 5, 7, 8, 10}
		break
	case "major":
		scale = []int{0, 2, 4, 5, 7, 9, 11}
		break
	case "chrom":
		scale = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
		break
	default:
		return nil, fmt.Errorf("invalid scale: %v", spec[0])
	}
	noteCount, err := strconv.Atoi(spec[2])
	if err != nil {
		return nil, err
	}
	for i := 0; i < noteCount; i += 1 {
		newNote := startNote + scale[i%len(scale)] + (i/len(scale)*12)
		instrumentList = append(instrumentList, MakePeriodicNote(newNote, timbre))
	}
	return instrumentList, nil
}

func MakePeriodicNote(note int, timbre GenericPeriodic) Instrument {
	frequency := NoteToFrequency(note)
	periodLen := int(1 / frequency * SAMPLE_RATE)
	lastArticulation := -ARTICULATION_LEN
	articulate := func(time int) {
		lastArticulation = time
	}
	getSample := func(sampleRate, time int) float64 {
		depth := time % periodLen
		if time - lastArticulation < ARTICULATION_LEN {
			return 0.5
		}
		return timbre(periodLen, depth)
	}
	return Instrument {
		articulate: articulate,
		getSample: getSample,
	}
}

func NoteToFrequency(note int) float64 {
	shiftedNote := note - 9 // a4 = 440hz = 0
	baseNote := (shiftedNote % 12 + 12) % 12

	// taken from https://en.wikipedia.org/wiki/Scientific_pitch_notation
	baseValue := []float64{
		261.6256, 277.1826, 293.6648, 311.1270, 329.6276, 349.2282,
		369.9944, 391.9954, 415.3047, 440.0000, 466.1638, 493.8833,
	}[baseNote]

	var octaveShift int
	if shiftedNote >= 0 {
		octaveShift = shiftedNote / 12
	} else {
		octaveShift = -((-shiftedNote + 11) / 12)
	}

	return baseValue * math.Pow(2, float64(octaveShift))
}

// c4 = 0
func ParseNote(note string) (int, error) {
	// get base note (not adjusted by octaves)
	invalidNote := fmt.Errorf("invalid note: %v", note)
	if len(note) < 2 || note[0] < 'a' || note[0] > 'g' {
		return 0, invalidNote
	}
	ret := []int{9, 11, 0, 2, 4, 5, 7}[note[0] - 'a']
	remaining := note[1:]
	switch remaining[0] {
	case 'b':
		ret -= 1
		remaining = remaining[1:]
		break
	case '#':
		ret += 1
		remaining = remaining[1:]
		break
	}
	ret += 12
	ret %= 12

	octave, err := strconv.Atoi(remaining)
	if err != nil {
		return 0, invalidNote
	}
	ret += (octave-4) * 12

	return ret, nil
}

func AddBeat(tune *Tune, instruments []Instrument, line []byte, linenum int) error {
	var noteData []byte
	var dynamicData []byte
	for i := 0; i < len(line); i += 1 {
		if line[i] == '|' {
			noteData = line[0:i]
			dynamicData = line[i+1:]
			goto foundSplit
		}
	}
	// we ignore empty lines
	return nil

foundSplit:
	if len(noteData) != len(instruments) {
		return fmt.Errorf("line %v: invalid instrument count", linenum)
	}
	var lineTuneData []Note
	for i, v := range(noteData) {
		if v == ' ' {
			continue;
		}
		newNote := Note {
			line: v,
			instrument: instruments[i],
		}
		lineTuneData = append(lineTuneData, newNote)
	}
	tune.tuneData = append(tune.tuneData, lineTuneData)

	lineDynamics := make(map[byte]Dynamic)
	dynamicParts := strings.Split(string(dynamicData), ",")

	var lastDynamics map[byte]Dynamic
	if len(tune.dynamics) == 0 {
		lastDynamics = nil
	} else {
		lastDynamics = tune.dynamics[len(tune.dynamics) - 1]
	}

	for _, v := range(dynamicParts) {
		dynamic := strings.TrimSpace(v)
		if dynamic == "" {
			continue
		}
		if len(dynamic) < 2  || dynamic[1] != '='{
			return fmt.Errorf("line %v: invalid dynamic %v", linenum, dynamic)
		}
		line := dynamic[0]
		info := dynamic[2:]
		if !IsValidLine(tune, line) {
			return fmt.Errorf("line %v: invalid line %v", line)
		}
		var prevDynamic Dynamic
		if lastDynamics == nil {
			prevDynamic = Dynamic {
				shouldChange: false,
				level: 0,
			}
		} else {
			prevDynamic = lastDynamics[line]
		}
		var err error
		lineDynamics[line], err = ParseDynamic(info, prevDynamic)
		if err != nil {
			return fmt.Errorf("line %v: %v", linenum, err)
		}
	}
	for _, v := range(tune.lines) {
		if _, ok := lineDynamics[v] ; ok {
			continue
		}
		prevDynamic := lastDynamics[v]
		prevDynamic.handWritten = false
		lineDynamics[v] = prevDynamic
	}
	tune.dynamics = append(tune.dynamics, lineDynamics)

	return nil
}

func ParseDynamic(info string, prevDynamic Dynamic) (Dynamic, error) {
	ret := Dynamic {
		handWritten: true,
		shouldChange: prevDynamic.shouldChange,
		level: prevDynamic.level,
	}
	parts := strings.Split(info, "&")
	explicitlyShouldChange := false
	for _, v := range(parts) {
		part := strings.TrimSpace(v)
		switch part {
		// crescendos and decrescendos are implicitly inferred from the
		// start and stop dynamics
		case "cresc": case "dec":
			ret.shouldChange = true
			explicitlyShouldChange = true
			continue
		case "hold":
			ret.shouldChange = true
			continue
		}
		// dynamics
		ret.shouldChange = explicitlyShouldChange
		var err error
		ret.level, err = DynamicStringToLevel(part)
		if err != nil {
			return Dynamic{}, err
		}
	}
	return ret, nil
}

func DynamicStringToLevel(dynamic string) (int, error) {
	if dynamic == "mp" {
		return -1, nil
	}
	if dynamic == "mf" {
		return 0, nil
	}
	if dynamic[0] != 'f' && dynamic[0] != 'p' {
		// these aren't really dynamics, they're line specs
		return 0, fmt.Errorf("invalid line spec %v", dynamic)
	}
	for i := 1; i < len(dynamic); i += 1 {
		if dynamic[i] != dynamic[i-1] {
			return 0, fmt.Errorf("invalid line spec %v", dynamic)
		}
	}
	if dynamic[0] == 'p' {
		return -1 - len(dynamic), nil
	}
	return len(dynamic), nil
}

func IsValidLine(tune *Tune, line byte) bool {
	for _, v := range(tune.lines) {
		if v == line {
			return true
		}
	}
	return false
}

func InterpolateDynamics(tune *Tune) {
	for _, v := range(tune.lines) {
		InterpolateDynamicLine(tune, v)
	}
}

func InterpolateDynamicLine(tune *Tune, line byte) {
	changingDynamics := false
	var changeStartIndex int
	var changeStartMultiplier int
	var changeEndIndex int
	var changeEndMultiplier int

	changeStartMultiplier = BASE_DYNAMIC_MULT
	for i := 0; i < len(tune.dynamics); i += 1 {
		thisDynamic := tune.dynamics[i][line]
		if !tune.dynamics[i][line].handWritten {
			if changingDynamics {
				progress := float64(i - changeStartIndex) / float64(changeEndIndex - changeStartIndex)
				newDynamic := float64(changeStartMultiplier) +
					float64(changeEndMultiplier - changeStartMultiplier) * progress
				thisDynamic.multiplier = int(newDynamic)
			} else {
				thisDynamic.multiplier = changeStartMultiplier
			}
			tune.dynamics[i][line] = thisDynamic
			continue
		}
		thisDynamic.multiplier = DynamicToMultiplier(thisDynamic.level)
		changeStartIndex = i
		changeStartMultiplier = thisDynamic.multiplier
		tune.dynamics[i][line] = thisDynamic
		changingDynamics = tune.dynamics[i][line].shouldChange
		if !changingDynamics {
			continue
		}

		// figure out where we're going with the crescendo/decrescendo
		nextHandWritten := -1
		for j := i+1; j < len(tune.dynamics); j += 1 {
			if tune.dynamics[j][line].handWritten {
				nextHandWritten = j
				break
			}
		}
		// there is no "next dynamic", so there is no dynamic change
		if nextHandWritten == -1 {
			changingDynamics = false
			continue
		}
		changeEndIndex = nextHandWritten
		changeEndMultiplier = DynamicToMultiplier(tune.dynamics[nextHandWritten][line].level)
	}
}

func DynamicToMultiplier(dynamic int) int {
	return int(BASE_DYNAMIC_MULT * math.Pow(DYNAMIC_STEP, float64(dynamic)))
}
