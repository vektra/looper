package gat

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

type Run struct {
	Tags    string
	failing []string
}

func (run *Run) RunAll() {
	run.goTest("./...")
}

func (run *Run) RunOnChange(file string) {
	if isGoFile(file) {
		// TODO: optimization, skip if no test files exist
		packageDir := "./" + filepath.Dir(file) // watchDir = ./
		run.goTest(packageDir)
	}
}

func (run *Run) runTest(test_files string) bool {
	args := []string{"test"}
	if len(run.Tags) > 0 {
		args = append(args, []string{"-tags", run.Tags}...)
	}
	args = append(args, test_files)

	command := "go"

	if _, err := os.Stat("Godeps/Godeps.json"); err == nil {
		args = append([]string{"go"}, args...)
		command = "godep"
	}

	cmd := exec.Command(command, args...)
	// cmd.Dir watchDir = ./

	PrintCommand(cmd.Args) // includes "go"

	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Println(err)
	}
	PrintCommandOutput(out)

	RedGreen(cmd.ProcessState.Success())
	ShowDuration(cmd.ProcessState.UserTime())

	return cmd.ProcessState.Success()
}

func (run *Run) goTest(test_files string) {
	if run.runTest(test_files) {
		// if test_files was in failing, remove it

		for idx, tf := range run.failing {
			if tf == test_files {
				run.failing = append(run.failing[:idx], run.failing[idx+1:]...)
				break
			}
		}

		for idx, tf := range run.failing {
			PrintRerun(fmt.Sprintf("Retrying failing tests: %s", tf))
			if !run.runTest(tf) {
				run.failing = run.failing[idx:]
				return
			}
		}

		run.failing = nil
	} else {
		for _, tf := range run.failing {
			if tf == test_files {
				return
			}
		}

		run.failing = append(run.failing, test_files)
	}
}

func isGoFile(file string) bool {
	return filepath.Ext(file) == ".go"
}
