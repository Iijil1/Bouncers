package main

import (
	"fmt"
	"reflect"
)

func decideBouncers(tm turingMachine, stepLimit int, printMode int) bool {
	return decideLeftBouncers(tm, false, stepLimit, printMode) ||
		decideLeftBouncers(tm.mirror(), true, stepLimit, printMode)
}

func decideLeftBouncers(tm turingMachine, mirrored bool, stepLimit int, printMode int) bool {
	records := findRecords(tm, stepLimit)
	numRecords := len(records)
	for i := 1; i*3 < numRecords; i++ {
		if checkRecords(tm, mirrored, [4]record{records[numRecords-1-3*i], records[numRecords-1-2*i], records[numRecords-1-i], records[numRecords-1]}, printMode) {
			return true
		}
	}
	return false
}

func findRecords(tm turingMachine, stepLimit int) []record {
	records := []record{}
	halfTapes := map[direction]*halfTape{L: {}, R: {}}
	headCon := headConfig{}
	for steps := 1; steps <= stepLimit; steps++ {
		tr, ok := tm.transitions[headCon]
		if !ok {
			break
		}
		halfTapes[!tr.direction].push(historySymbol{historySlice{headCon}, tr.symbol})
		headCon.state = tr.state
		newSy := halfTapes[tr.direction].pop()
		if newSy != nil {
			headCon.symbol = newSy.base()
		} else {
			headCon.symbol = baseSymbol(0)
			if tr.direction == L {
				records = append(records, record{headCon.state, steps, *halfTapes[R]})
			}
		}
	}
	return records
}

func checkRecords(tm turingMachine, mirrored bool, records [4]record, printMode int) bool {
	if !sameStates(records) {
		return false
	}
	if !quadraticProgression(records) {
		return false
	}
	dirSequence1, historyTape1 := findContext(tm, records[0], records[1].steps-records[0].steps)
	dirSequence2, historyTape2 := findContext(tm, records[1], records[2].steps-records[1].steps)

	bufSize := findBufferSize(dirSequence1, dirSequence2)

	growth := historyTape2.len - historyTape1.len
	colorTape1 := findColors(historyTape1, growth)
	colorTape2 := findColors(historyTape2, growth)

	words := findRepeaters(colorTape1, colorTape2, bufSize)
	if words == nil {
		return false
	}

	//records[i] has buffer + repeater^(i-1) + walls
	start := findStart(tm, records[1], bufSize, words, records[2].steps)
	rules := findRules(tm, start, records[3].steps-records[2].steps)
	if rules == nil {
		return false
	}
	if mirrored {
		tm = tm.mirror()
	}
	return verifyBouncer(fullCert{tm, mirrored, start, rules}, printMode)
}

func sameStates(records [4]record) bool {
	return records[0].state == records[1].state && records[0].state == records[2].state && records[0].state == records[3].state
}

func quadraticProgression(records [4]record) bool {
	diff := [3]int{records[1].steps - records[0].steps, records[2].steps - records[1].steps, records[3].steps - records[2].steps}
	diffdiff := [2]int{diff[1] - diff[0], diff[2] - diff[1]}
	return diffdiff[0] > 0 && diffdiff[0] == diffdiff[1]
}

func findContext(tm turingMachine, startRecord record, stepLimit int) ([]int, halfTape) {
	directions := []int{0}
	halfTapes := map[direction]*halfTape{L: {}, R: &startRecord.tape}
	headCon := headConfig{startRecord.state, baseSymbol(0)}
	lastDir := L
	var lastCol historySlice = nil
	for steps := 1; steps <= stepLimit; steps++ {
		tr, ok := tm.transitions[headCon]
		if !ok {
			break
		}
		halfTapes[!tr.direction].push(historySymbol{append(lastCol, headCon), tr.symbol})
		headCon.state = tr.state
		newSy := halfTapes[tr.direction].pop()
		if newSy != nil {
			headCon.symbol = newSy.base()
			lastCol = newSy.history()
		} else {
			headCon.symbol = baseSymbol(0)
			lastCol = nil
		}
		if lastDir == tr.direction {
			directions[len(directions)-1] += 1
		} else {
			directions = append(directions, 1)
		}
		lastDir = tr.direction
	}
	return directions, *halfTapes[R]
}

func findBufferSize(sequence1 []int, sequence2 []int) int {
	bufSize := 0
	for !checkDirectionMatch(sequence1, sequence2) {
		bufSize += 1
		sequence1 = smoothenSequence(sequence1, bufSize)
		sequence2 = smoothenSequence(sequence2, bufSize)
	}
	return bufSize
}

func checkDirectionMatch(seq1 []int, seq2 []int) bool {
	if len(seq1) != len(seq2) {
		return false
	}
	for i := 1; i < len(seq1)-1; i++ {
		if seq1[i] == seq2[i] {
			return false
		}
	}
	return true
}

func smoothenSequence(seq []int, buf int) []int {
	//move indices and build a new slice for O(n) speed.
	//deleting elements of the slice in place and copying
	//the rest closer leads to O(n^2). That was bad.
	i, j, k := 0, 1, 2
	res := []int{}
	for k < len(seq) {
		if seq[j] <= buf {
			seq[i] = seq[i] - seq[j] + seq[k]
			j, k = k+1, k+2
		} else {
			res = append(res, seq[i])
			i, j, k = j, k, k+1
		}
	}
	res = append(res, seq[i])
	if j < len(seq) {
		res = append(res, seq[j])
	}
	return res
}

func findColors(tape halfTape, n int) halfTape {
	fullHistory := make([]historySlice, tape.len+2*n)
	storage := halfTape{}
	pos := n
	for elem := tape.pop(); elem != nil; elem = tape.pop() {
		fullHistory[pos] = elem.history()
		pos += 1
		storage.push(elem.base())
	}

	preColorMap := map[string]color{}
	lastPreColor := color(0)
	fullPreColor := make([]color, len(fullHistory))
	for i, history := range fullHistory {
		historyIndex := fmt.Sprint(history)
		preCol, ok := preColorMap[historyIndex]
		if !ok {
			preCol = lastPreColor + 1
			preColorMap[historyIndex] = preCol
			lastPreColor = preCol
		}
		fullPreColor[i] = preCol
	}
	colorMap := map[string]color{}
	lastColor := color(0)
	for elem := storage.pop(); elem != nil; elem = storage.pop() {
		pos -= 1
		curColorSlice := fullPreColor[pos-n : pos+n]
		colorIndex := fmt.Sprint(curColorSlice)
		col, ok := colorMap[colorIndex]
		if !ok {
			col = lastColor + 1
			colorMap[colorIndex] = col
			lastColor = col
		}
		tape.push(colorSymbol{col, elem.base()})
	}
	return tape
}

func findRepeaters(tape1 halfTape, tape2 halfTape, bufSize int) []word {
	for i := 0; i < bufSize; i++ {
		tape1.pop()
		tape2.pop()
	}

	words := []word{}
	symbol1 := tape1.pop()
	symbol2 := tape2.pop()
	curWord := word{}
	for symbol2 != nil {
		if len(words)%2 == 0 {
			if reflect.DeepEqual(symbol1, symbol2) {
				curWord = append(curWord, symbol2.base())
				symbol1 = tape1.pop()
				symbol2 = tape2.pop()
			} else {
				words = append(words, curWord)
				curWord = word{symbol2.base()}
				symbol2 = tape2.pop()
			}
		} else {
			if reflect.DeepEqual(symbol1, symbol2) {
				words = append(words, curWord)
				curWord = word{symbol2.base()}
				symbol1 = tape1.pop()
				symbol2 = tape2.pop()
			} else {
				curWord = append(curWord, symbol2.base())
				symbol2 = tape2.pop()
			}
		}
	}
	if symbol1 != nil {
		return nil
	}
	words = append(words, curWord)
	if len(words)%2 == 0 {
		words = append(words, word{})
	}
	return words
}

func findStart(tm turingMachine, record record, bufSize int, words []word, stepLimit int) initialConditions {
	startState := record.state
	startPos := 0
	startTape := make([]baseSymbol, bufSize+len(words[0])+1)
	for i := 1; i < len(startTape); i++ {
		sy := record.tape.pop()
		if sy != nil {
			startTape[i] = sy.base()
		}
	}
	growth := map[direction]bool{
		L: true,
		R: false,
	}
	actualState, _, actualTape, actualSteps := runTm(tm, startState, startPos, startTape, stepLimit, growth)
	buffer := make([]baseSymbol, bufSize)
	copy(buffer, actualTape[len(actualTape)-bufSize:])
	words[0] = actualTape[:len(actualTape)-bufSize]
	start := initialConditions{
		Steps:  actualSteps + record.steps,
		Words:  words,
		State:  actualState,
		Buffer: buffer,
	}
	return start
}

func findRules(tm turingMachine, start initialConditions, stepLimit int) []transitionRule {
	if len(start.Words) < 3 {
		return nil
	}
	curState := start.State
	curDir := R
	curGlobalPos := 1
	curBuffer := start.Buffer
	curWords := make([]word, len(start.Words))
	copy(curWords, start.Words)
	rules := []transitionRule{}
	for stepLimit > 0 {
		startState := curState
		startInnerPos := 0
		startTape := make([]baseSymbol, len(curBuffer)+len(curWords[curGlobalPos]))
		switch curDir {
		case L:
			copy(startTape, append(curWords[curGlobalPos], curBuffer...))
			startInnerPos = len(curWords[curGlobalPos]) - 1
		case R:
			copy(startTape, append(curBuffer, curWords[curGlobalPos]...))
			startInnerPos = len(curBuffer)
		}
		growth := map[direction]bool{
			L: curGlobalPos == 0,
			R: curGlobalPos == len(curWords)-1,
		}

		endState, endPos, endTape, steps := runTm(tm, startState, startInnerPos, startTape, stepLimit, growth)
		endDir := R
		endBuffer := word{}
		endWord := word{}
		endStub := word{}
		bufSize := len(start.Buffer)
		switch endPos {
		case -1:
			endDir = L
			endBuffer = append(endBuffer, endTape[:bufSize]...)
			endWord = append(endWord, endTape[bufSize:]...)
		default:
			if endPos-bufSize < 0 {
				return nil
			}
			endWord = append(endWord, endTape[:endPos-bufSize]...)
			endBuffer = append(endBuffer, endTape[endPos-bufSize:endPos]...)
			endStub = append(endStub, endTape[endPos:]...)
		}

		rule := transitionRule{
			StartWord:   curWords[curGlobalPos],
			StartDir:    curDir,
			StartState:  curState,
			StartBuffer: curBuffer,
			Steps:       steps,
			Growing:     curGlobalPos == 0 || curGlobalPos == len(curWords)-1,
			EndWord:     endWord,
			EndDir:      endDir,
			EndState:    endState,
			EndBuffer:   endBuffer,
			Stub:        endStub,
		}
		if len(rules)%2 == 0 && !checkChainRule(rule) {
			return nil
		}
		rules = append(rules, rule)
		stepLimit -= steps

		curState = rule.EndState
		curBuffer = rule.EndBuffer
		curWords[curGlobalPos] = rule.EndWord
		curDir = rule.EndDir
		switch curDir {
		case L:
			curGlobalPos -= 1
		case R:
			curGlobalPos += 1
		}
		if curGlobalPos < 0 || curGlobalPos >= len(curWords) {
			return nil
		}
	}
	actualStub := rules[len(rules)-1].Stub
	if !checkInduction(curState, curDir, curGlobalPos, curBuffer, curWords, actualStub, start) {
		return nil
	}
	return rules
}
