package main

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func runTm(tm turingMachine, startState tmState, startPos int, startTape []baseSymbol, stepLimit int, growth map[direction]bool) (finalState tmState, finalPos int, finalTape []baseSymbol, steps int) {
	finalTape = make([]baseSymbol, len(startTape))
	copy(finalTape, startTape)
	finalPos = startPos
	finalState = startState
	if finalPos < 0 || finalPos >= len(finalTape) {
		return
	}
	for steps = 1; steps <= stepLimit; steps++ {

		tr, ok := tm.transitions[headConfig{finalState, finalTape[finalPos]}]
		if !ok {
			return
		}
		finalTape[finalPos] = tr.symbol
		finalState = tr.state
		if tr.direction == L {
			finalPos -= 1
		} else {
			finalPos += 1
		}
		switch finalPos {
		case -1:
			//leaving tape to the left
			if !growth[L] {
				return
			}
			finalTape = append([]baseSymbol{0}, finalTape...)
			finalPos = 0

		case len(finalTape):
			//leaving tape to the right
			if !growth[R] {
				return
			}
			finalTape = append(finalTape, baseSymbol(0))
		}
	}
	steps--
	return
}

func verifyBouncer(cert fullCert, printMode int) bool {
	tm := cert.Tm
	if cert.Mirror {
		tm = tm.mirror()
	}
	result := checkInitialConditions(tm, cert.Start) &&
		checkRules(tm, cert.Rules) &&
		checkApplication(cert.Start, cert.Rules)

	if result && printMode >= 0 {
		printCert(cert, printMode)
	}
	return result
}

func checkInitialConditions(tm turingMachine, start initialConditions) bool {

	if len(start.Words) < 3 || len(start.Words)%2 != 1 {
		return false
	}

	startState := tmState(0)
	startPos := 0
	startTape := []baseSymbol{0}
	stepLimit := start.Steps
	growth := map[direction]bool{
		L: true,
		R: true,
	}

	claimedState := start.State
	claimedPos := len(start.Words[0]) + len(start.Buffer)
	claimedTape := make([]baseSymbol, claimedPos)
	copy(claimedTape, append(start.Words[0], start.Buffer...))
	for i := 2; i < len(start.Words); i += 2 {
		claimedTape = append(claimedTape, start.Words[i]...)
	}
	claimedSteps := start.Steps

	actualState, actualPos, actualTape, actualSteps := runTm(tm, startState, startPos, startTape, stepLimit, growth)

	return claimedState == actualState &&
		claimedPos == actualPos &&
		claimedSteps == actualSteps &&
		reflect.DeepEqual(claimedTape, actualTape)
}

func checkRules(tm turingMachine, rules []transitionRule) bool {
	if len(rules) < 2 || len(rules)%2 != 0 {
		return false
	}
	for i, rule := range rules {
		if !checkRule(tm, rule) {
			return false
		}
		if i%2 == 0 && !checkChainRule(rule) {
			return false
		}
	}
	return true
}

func checkRule(tm turingMachine, rule transitionRule) bool {
	if len(rule.StartBuffer) != len(rule.EndBuffer) {
		return false
	}

	startState := rule.StartState
	startPos := 0
	startTape := make([]baseSymbol, len(rule.StartBuffer)+len(rule.StartWord))
	switch rule.StartDir {
	case L:
		copy(startTape, append(rule.StartWord, rule.StartBuffer...))
		startPos = len(rule.StartWord) - 1
	case R:
		copy(startTape, append(rule.StartBuffer, rule.StartWord...))
		startPos = len(rule.StartBuffer)
	}
	stepLimit := rule.Steps
	growth := map[direction]bool{
		rule.StartDir:  rule.Growing,
		!rule.StartDir: false,
	}

	claimedState := rule.EndState
	claimedPos := 0
	claimedTape := make([]baseSymbol, len(rule.EndBuffer)+len(rule.EndWord)+len(rule.Stub))
	switch rule.EndDir {
	case L:
		copy(claimedTape, append(rule.Stub, append(rule.EndBuffer, rule.EndWord...)...))
		claimedPos = len(rule.Stub) - 1
	case R:
		copy(claimedTape, append(rule.EndWord, append(rule.EndBuffer, rule.Stub...)...))
		claimedPos = len(rule.EndWord) + len(rule.EndBuffer)
	}
	claimedSteps := rule.Steps

	actualState, actualPos, actualTape, actualSteps := runTm(tm, startState, startPos, startTape, stepLimit, growth)

	return claimedState == actualState &&
		claimedPos == actualPos &&
		claimedSteps == actualSteps &&
		reflect.DeepEqual(claimedTape, actualTape)
}

func checkChainRule(rule transitionRule) bool {
	return rule.StartState == rule.EndState &&
		rule.StartDir == rule.EndDir &&
		!rule.Growing &&
		len(rule.StartWord) > 0 &&
		len(rule.Stub) == 0 &&
		reflect.DeepEqual(rule.StartBuffer, rule.EndBuffer)
}

func checkApplication(start initialConditions, rules []transitionRule) bool {
	actualState := start.State
	actualDir := R
	actualPos := 1
	actualBuffer := start.Buffer
	actualWords := make([]word, len(start.Words))
	copy(actualWords, start.Words)
	for i, rule := range rules {
		if !checkRuleContext(actualState, actualDir, actualPos, actualBuffer, actualWords, rule, i == len(rules)-1) {
			return false
		}
		actualState = rule.EndState
		actualBuffer = rule.EndBuffer
		actualWords[actualPos] = rule.EndWord
		actualDir = rule.EndDir
		switch actualDir {
		case L:
			actualPos -= 1
		case R:
			actualPos += 1
		}
		if actualPos < 0 || actualPos >= len(actualWords) {
			return false
		}
	}
	actualStub := rules[len(rules)-1].Stub

	return checkInduction(actualState, actualDir, actualPos, actualBuffer, actualWords, actualStub, start)
}

func checkRuleContext(curState tmState, curDir direction, curPos int, curBuffer word, curWords []word, rule transitionRule, lastRule bool) bool {
	return rule.StartState == curState &&
		rule.StartDir == curDir &&
		rule.Growing == (curPos == 0 || curPos == len(curWords)-1) &&
		(len(rule.Stub) == 0 || lastRule) &&
		reflect.DeepEqual(rule.StartBuffer, curBuffer) &&
		reflect.DeepEqual(rule.StartWord, curWords[curPos])
}

func checkInduction(actualState tmState, actualDir direction, actualPos int, actualBuffer word, actualWords []word, actualStub word, start initialConditions) bool {
	if actualState != start.State ||
		actualDir != R ||
		actualPos != 1 ||
		!reflect.DeepEqual(actualBuffer, start.Buffer) ||
		!reflect.DeepEqual(actualWords[0], start.Words[0]) {
		return false
	}
	actualRightWords := make([]word, len(actualWords))
	copy(actualRightWords, actualWords)
	actualRightWords[0] = actualStub

	claimedRightWords := make([]word, len(start.Words))
	copy(claimedRightWords, start.Words)
	claimedRightWords[0] = word{}
	for i := 1; i < len(claimedRightWords); i += 2 {
		claimedRightWords[i-1] = append(claimedRightWords[i-1], claimedRightWords[i]...)
	}

	return checkWords(actualRightWords, claimedRightWords)
}

func checkWords(actualRightWords []word, claimedRightWords []word) bool {
	rightAlign(actualRightWords)
	rightAlign(claimedRightWords)
	return reflect.DeepEqual(actualRightWords, claimedRightWords)
}

func rightAlign(words []word) {
	for i := len(words) - 1; i > 1; i -= 2 {
		for len(words[i]) > 0 && words[i][0] == words[i-1][0] {
			sy := words[i][0]
			words[i] = words[i][1:]
			words[i-1] = append(words[i-1][1:], sy)
			words[i-2] = append(words[i-2], sy)
		}
	}
}

func printCert(cert fullCert, printMode int) {
	switch printMode {
	case 0:
		fmt.Println(cert.Tm)
	case 1:
		sCert := shortCert{
			Tm:         cert.Tm,
			Mirror:     cert.Mirror,
			Start:      cert.Start,
			CycleSteps: 0,
		}
		for _, rule := range cert.Rules {
			sCert.CycleSteps += rule.Steps
		}
		b, err := json.Marshal(sCert)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	case 2:
		b, err := json.Marshal(cert)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	case 3:
		sCert := shortCert{
			Tm:         cert.Tm,
			Mirror:     cert.Mirror,
			Start:      cert.Start,
			CycleSteps: 0,
		}
		for _, rule := range cert.Rules {
			sCert.CycleSteps += rule.Steps
		}
		b, err := json.MarshalIndent(sCert, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	case 4:
		b, err := json.MarshalIndent(cert, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(b))
	}
}
