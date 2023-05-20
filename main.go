package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func main() {
	fullCert := flag.Bool("fc", false, "checks certificates instead of running the decider")
	shortCert := flag.Bool("sc", false, "checks certificates instead of running the decider")
	stepLimit := flag.Int("n", 10000, "scans with this stepLimit")
	exact := flag.Bool("x", false, "only tests for records at steplimit, for use with filtered input")
	printMode := flag.Int("pm", 0, "what to print: 0 -> solved TMs, 1 -> certificates, 2 -> certificates with indent")
	cores := flag.Int("cores", 0, "maximum number of TMs to work on in parallel")

	flag.Parse()

	if *cores <= 0 {
		*cores = runtime.GOMAXPROCS(0)
	}
	workTokens := make(chan struct{}, *cores)
	for i := 0; i < *cores; i++ {
		workTokens <- struct{}{}
	}
	input := bufio.NewReader(os.Stdin) //a Scanner would be more convenient, but the strings for some full certificates are too long

	switch {
	case *fullCert:
		checkFullCerts(input, workTokens, *printMode)
	case *shortCert:
		checkShortCerts(input, workTokens, *printMode)
	default:
		runScan(input, workTokens, *stepLimit, *exact, *printMode)
	}

	//make sure all the work is finished
	for i := 0; i < *cores; i++ {
		_ = <-workTokens
	}
}

func checkFullCerts(input *bufio.Reader, workTokens chan struct{}, printMode int) {
	var readerErr error
	for readerErr == nil {
		var text string
		text, readerErr = input.ReadString('\n')
		if text == "" {
			continue
		}
		text = strings.TrimSpace(text)
		_ = <-workTokens
		go func() {
			defer func() {
				workTokens <- struct{}{}
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "Panic at %s\n%s\n", text, err)
				}
			}()
			cert := fullCert{}
			err := json.Unmarshal([]byte(text), &cert)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to parse %s\n%s\n", text, err)
				return
			}
			verifyBouncer(cert, printMode)
		}()
	}
	if readerErr != io.EOF {
		fmt.Fprintln(os.Stderr, readerErr)
	}
}

func checkShortCerts(input *bufio.Reader, workTokens chan struct{}, printMode int) {
	var readerErr error
	for readerErr == nil {
		var text string
		text, readerErr = input.ReadString('\n')
		if text == "" {
			continue
		}
		text = strings.TrimSpace(text)
		_ = <-workTokens
		go func() {
			defer func() {
				workTokens <- struct{}{}
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "Panic at %s\n%s\n", text, err)
				}
			}()
			cert := shortCert{}
			err := json.Unmarshal([]byte(text), &cert)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to parse %s\n%s\n", text, err)
				return
			}
			tm := cert.Tm
			if cert.Mirror {
				tm = tm.mirror()
			}
			rules := findRules(tm, cert.Start, cert.CycleSteps)
			if rules == nil {
				return
			}
			verifyBouncer(fullCert{cert.Tm, cert.Mirror, cert.Start, rules}, printMode)
		}()
	}
	if readerErr != io.EOF {
		fmt.Fprintln(os.Stderr, readerErr)
	}
}

func runScan(input *bufio.Reader, workTokens chan struct{}, stepLimit int, exact bool, printMode int) {
	var readerErr error
	for readerErr == nil {
		var text string
		text, readerErr = input.ReadString('\n')
		if text == "" {
			continue
		}
		text = strings.TrimSpace(text)
		_ = <-workTokens
		go func() {
			defer func() {
				workTokens <- struct{}{}
				if err := recover(); err != nil {
					fmt.Fprintf(os.Stderr, "Panic at %s\n%s\n", text, err)
				}
			}()
			tm := parseTM(text)
			if tm.numStates == 0 {
				fmt.Fprintf(os.Stderr, "Unable to parse %s\n", text)
				return
			}
			if !exact {
				for n := 100; n < stepLimit; n *= 10 {
					if decideBouncers(tm, n, printMode) {
						return
					}
				}
			}
			decideBouncers(tm, stepLimit, printMode)
		}()
	}
	if readerErr != io.EOF {
		fmt.Fprintln(os.Stderr, readerErr)
	}
}
