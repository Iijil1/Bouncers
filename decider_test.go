package main

import (
	"testing"
)

func TestDecider(t *testing.T) {
	// tm := parseTM("1LB---_0LC1LD_0RD1LC_1RE0LA_1LA0RE")
	// tm := parseTM("1RB1LC_0LA0RB_1RD1LE_0RB1RC_---0LB")
	tm := parseTM("1RB1RD_1LC1LE_1RA0LB_0RA---_0RC0RB")
	if !decideBouncers(tm, 1700, 4) {
		t.Fail()
	}
}
