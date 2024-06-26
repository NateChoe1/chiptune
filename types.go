package main

const SAMPLE_RATE = 44100
const ARTICULATION_LEN = 1000

const BASE_DYNAMIC_MULT = 2000
const DYNAMIC_STEP = 1.5

// an instrument can only play 1 note
type Instrument struct {
	articulate func(time int)

	// NOTE: for notes of around 440hz, the number should be in the range of
	// [-1,1].

	// this function MUST be called with a monotonically increasing location
	getSample func(sampleRate, time int) float64 // returns a number centered at 0
}

type Note struct {
	line byte // capitalized if this note is articulated
	instrument Instrument
}

type Dynamic struct {
	handWritten bool // was this in the original file, or was it interpolated?
	shouldChange bool // should i hold this dynamic, or crescendo/decrescendo
	level int // ..., -4 = ppp, -3 = pp, -2 = p, -1 = mp, 0 = mf, 1 = f, 2 = ff, 3 = fff, ...
	          // this value extends infinitely in both directions
	multiplier int // how should we scale this value in the final mix (precalculated)
}

type Tune struct {
	tempo int // tempo in bpm
	samplesPerBeat int // samples per beat (precalculated)
	lines []byte // lines are a single lowercase letter
	tuneData [][]Note
	dynamics []map[byte]Dynamic
}
