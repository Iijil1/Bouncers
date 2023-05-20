package main

import (
	"errors"
	"fmt"
	"strings"
)

type baseSymbol int

func (bs baseSymbol) base() baseSymbol {
	return bs
}

func (baseSymbol) history() historySlice {
	return nil
}

func (baseSymbol) col() color {
	return 0
}

type symbol interface {
	base() baseSymbol
	history() historySlice
	col() color
}

type historySymbol struct {
	historySlice
	baseSymbol
}

func (cs historySymbol) history() historySlice {
	return cs.historySlice
}

type historySlice []headConfig

type colorSymbol struct {
	color
	baseSymbol
}

func (cs colorSymbol) col() color {
	return cs.color
}

type color int

type halfTapeCell struct {
	value symbol
	next  *halfTapeCell
}

func (htc halfTapeCell) tapeString(reverse bool) string {
	if htc.next == nil {
		return fmt.Sprint(htc.value.base())
	}
	restString := htc.next.tapeString(reverse)
	if reverse {
		return fmt.Sprintf("%s%v", restString, htc.value.base())
	}
	return fmt.Sprintf("%v%s", htc.value.base(), restString)
}

type halfTape struct {
	first *halfTapeCell
	len   int
}

func (ht halfTape) tapeString(reverse bool) string {
	if ht.first == nil {
		return ""
	}
	return ht.first.tapeString(reverse)
}

func (ht halfTape) String() string {
	return ht.tapeString(false)
}

func (ht *halfTape) push(sy symbol) {
	htc := halfTapeCell{
		value: sy,
		next:  ht.first,
	}
	ht.first = &htc
	ht.len += 1
}

func (ht *halfTape) pop() symbol {
	sy := ht.first
	if sy == nil {
		return nil
	}
	ht.first = sy.next
	ht.len -= 1
	return sy.value
}

type record struct {
	state tmState
	steps int
	tape  halfTape
}

func (rec record) String() string {
	return fmt.Sprintf("<%v%v", rec.state, rec.tape)
}

type tmState int

const A tmState = 0
const B tmState = 1
const C tmState = 2
const D tmState = 3
const E tmState = 4
const F tmState = 5

func (tms tmState) String() string {
	return string(rune('A' + tms))
}

func (tms tmState) MarshalText() ([]byte, error) {
	return []byte(tms.String()), nil
}

func (tms *tmState) UnmarshalText(text []byte) error {
	if len(text) != 1 {
		return errors.New("Unable to parse TM state " + string(text))
	}
	*tms = tmState(text[0] - 'A')
	return nil
}

type headConfig struct {
	state  tmState
	symbol baseSymbol
}

func (hc headConfig) String() string {
	return fmt.Sprintf("%v%v", hc.state, hc.symbol)
}

type direction bool

const L direction = true
const R direction = false

func (d direction) String() string {
	if d {
		return "L"
	}
	return "R"
}

func (d direction) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *direction) UnmarshalText(text []byte) error {
	switch string(text) {
	case "L":
		*d = L
	case "R":
		*d = R
	default:
		return errors.New("Unable to parse direction: " + string(text))
	}
	return nil
}

type transition struct {
	symbol    baseSymbol
	direction direction
	state     tmState
}

func (tr transition) String() string {
	return fmt.Sprintf("%v%v%v", tr.symbol, tr.direction, tr.state)
}

type turingMachine struct {
	numStates   int
	numSymbols  int
	transitions map[headConfig]transition
}

func (tm turingMachine) String() string {
	if tm.numStates <= 0 || tm.numSymbols <= 0 {
		return ""
	}
	standardFormat := ""
	for i := 0; i < tm.numStates; i++ {
		standardFormat += "_"
		for j := 0; j < tm.numSymbols; j++ {
			hc := headConfig{tmState(i), baseSymbol(j)}
			if tr, ok := tm.transitions[hc]; ok {
				standardFormat += tr.String()
			} else {
				standardFormat += "---"
			}
		}
	}
	return standardFormat[1:]
}

//standard text format
func parseTM(s string) turingMachine {
	defer func() {
		recover()
	}()
	stateStrings := strings.Split(s, "_")
	if len(stateStrings[0])%3 != 0 {
		panic("")
	}
	tm := turingMachine{
		numStates:   len(stateStrings),
		numSymbols:  len(stateStrings[0]) / 3,
		transitions: map[headConfig]transition{},
	}
	if tm.numStates < 2 {
		panic("")
	}
	for i, stateString := range stateStrings {
		if len(stateString) != tm.numSymbols*3 {
			panic("")
		}
		for j := 0; len(stateString) >= 3; j++ {
			symbolString := stateString[:3]
			stateString = stateString[3:]
			newTMState := tmState(symbolString[2] - 'A')
			if int(newTMState) < 0 || int(newTMState) >= tm.numStates {
				continue
			}
			newSymbol := baseSymbol(symbolString[0] - '0')
			if newSymbol < 0 || int(newSymbol) >= tm.numSymbols {
				panic("")
			}
			newDirection := L
			if symbolString[1] == 'R' {
				newDirection = R
			}
			tm.transitions[headConfig{tmState(i), baseSymbol(j)}] = transition{newSymbol, newDirection, newTMState}
		}
	}
	return tm
}

func (tm turingMachine) MarshalText() ([]byte, error) {
	return []byte(tm.String()), nil
}

func (tm *turingMachine) UnmarshalText(text []byte) error {
	*tm = parseTM(string(text))
	if tm.numStates == 0 {
		return errors.New("Unable to parse TM string: " + string(text))
	}
	return nil
}

func (tm turingMachine) mirror() turingMachine {
	newTm := turingMachine{
		numStates:   tm.numStates,
		numSymbols:  tm.numSymbols,
		transitions: map[headConfig]transition{},
	}
	for hc, tr := range tm.transitions {
		tr.direction = !tr.direction
		newTm.transitions[hc] = tr
	}
	return newTm
}

// A0
// --steps-->
// word0 buffer S> word1 word2 ... wordN
type initialConditions struct {
	Steps  int
	Words  []word
	State  tmState
	Buffer word
}

//  buffer1 S1> word1 || word1 <S1 buffer1
//  --steps-->
//  word2 buffer2 S2> stub  || stub <S2 buffer2 word2
type transitionRule struct {
	StartWord   word
	StartDir    direction
	StartState  tmState
	StartBuffer word
	Steps       int
	Growing     bool
	EndWord     word
	EndDir      direction
	EndState    tmState
	EndBuffer   word
	Stub        word
}

type fullCert struct {
	Tm     turingMachine
	Mirror bool
	Start  initialConditions
	Rules  []transitionRule
}

type shortCert struct {
	Tm         turingMachine
	Mirror     bool
	Start      initialConditions
	CycleSteps int
}

type word []baseSymbol

func (w word) MarshalText() ([]byte, error) {
	str := ""
	for _, sy := range w {
		str += fmt.Sprint(sy)
	}
	return []byte(str), nil
}

func (w *word) UnmarshalText(text []byte) error {
	*w = make(word, len(text))
	for i, sy := range text {
		(*w)[i] = baseSymbol(sy - '0')
	}
	return nil
}
