package main

import (
	"log"
	"os"

	"github.com/goyek/goyek"
	shellwords "github.com/mattn/go-shellwords"
)

const buildDir = "build"

func main() {
	if err := os.Chdir(".."); err != nil {
		log.Fatalln(err)
	}
	flow().Main()
}

func flow() *goyek.Taskflow {
	flow := &goyek.Taskflow{}

	test := flow.Register(taskTest())
	lint := flow.Register(taskLint())
	misspell := flow.Register(taskMisspell())
	coverage := flow.Register(taskGenerateCoverage(goyek.Deps{test}))
	all := flow.Register(taskAll(goyek.Deps{
		test, lint, misspell, coverage,
	}))

	flow.DefaultTask = all
	return flow
}

func taskTest() goyek.Task {
	return goyek.Task{
		Name:  "test",
		Usage: "go test with code coverage",
		Command: func(tf *goyek.TF) {
			Exec(tf, "", "go test -covermode=atomic -coverprofile=coverage.out ./...")
		},
	}
}

func taskLint() goyek.Task {
	return goyek.Task{
		Name:  "lint",
		Usage: "golangci-lint",
		Command: func(tf *goyek.TF) {
			Exec(tf, buildDir, "go install github.com/golangci/golangci-lint/cmd/golangci-lint")
			Exec(tf, "", "golangci-lint run")
		},
	}
}

func taskMisspell() goyek.Task {
	return goyek.Task{
		Name:  "misspell",
		Usage: "misspell",
		Command: func(tf *goyek.TF) {
			Exec(tf, buildDir, "go install github.com/client9/misspell/cmd/misspell")
			Exec(tf, "", "misspell -error -locale=US *.md")
		},
	}
}

func taskAll(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "all",
		Usage: "build pipeline",
		Deps:  deps,
	}
}

// Exec runs the provided command line.
// It fails the task in case of any problems.
func Exec(tf *goyek.TF, workDir string, cmdLine string) {
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		tf.Fatalf("parse command line: %v", err)
	}
	cmd := tf.Cmd(args[0], args[1:]...)
	cmd.Dir = workDir
	if err := cmd.Run(); err != nil {
		tf.Fatalf("run command: %v", err)
	}
}

func taskGenerateCoverage(deps goyek.Deps) goyek.Task {
	return goyek.Task{
		Name:  "coverage",
		Usage: "generate coverage metrics by running test and generating markdown badge updates",
		Deps:  deps,
		Command: func(tf *goyek.TF) {
			_ = os.Mkdir("artifacts", 0700)
			Exec(tf, "", "go install github.com/jpoles1/gopherbadger@master")
			Exec(tf, "", "go test ./backlinker/ -coverprofile ./artifacts/cover.out")
			Exec(tf, "", "go tool cover -html=./artifacts/cover.out -o ./artifacts/coverage.html")
			Exec(tf, "", "gopherbadger -md=\"README.md,coverage.md\" -tags 'unit'")
		},
	}
}
