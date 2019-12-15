package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const TestTimeout = 15

type TestOutcome struct {
	name, message string
	ok            bool
}

func worker(id int, jobs <-chan string, results chan<- *TestOutcome, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		outcome := runTest(job)
		if outcome.ok {
			fmt.Printf(".")
		} else {
			fmt.Printf("x")
		}
		results <- outcome
	}
}

func runTest(path string) *TestOutcome {
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeout*time.Second)
	cmd := exec.CommandContext(ctx, "./goboy", path)
	outcome := &TestOutcome{name: path}
	defer cancel()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(errors.Wrap(err, "couldn't open stdout"))
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(errors.Wrap(err, "couldn't start cmd"))
	}

	var output string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		text := scanner.Text()
		output += "."
		output += text
		if strings.Contains(output, "Passed") {
			outcome.ok = true
			return outcome
		} else if strings.Contains(output, "Failed") {
			outcome.message = output
			return outcome
		}
	}
	if err := cmd.Wait(); err != nil {
		outcome.message = "timeout"
		return outcome
	}
	outcome.message = "error"
	return outcome
}

func main() {
	runtime.GOMAXPROCS(4)
	var wg sync.WaitGroup
	paths := make(chan string, 100)
	results := make(chan *TestOutcome, 100)

	for w := 1; w <= 4; w++ {
		wg.Add(1)
		go worker(w, paths, results, &wg)
	}

	_ = filepath.Walk("specs", func(path string, _ os.FileInfo, _ error) error {
		if filepath.Ext(path) == ".gb" {
			paths <- path
		}
		return nil
	})
	close(paths)

	wg.Wait()
	close(results)

	fmt.Printf("\n")
	fmt.Printf("=============\n")
	fmt.Printf("\n")
	for test := range results {
		if !test.ok {
			fmt.Printf("Test %s failed: %s\n", test.name, test.message)
		}
	}

}
