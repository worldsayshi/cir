package main

import (
	"log"
	"os"
	"os/exec"
)

func setupLogging() (f *os.File, err error) {
	f, err = os.OpenFile("go-coder.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return nil, err
	}

	log.SetOutput(f)
	return f, nil
}

func addContextFile() {
	cmd := "find . -type f -not -path '*/.*' | fzf-tmux -h"
	out, err := exec.Command(
		"bash", "-c", cmd,
		// "find", ".", "-type", "f", "-not", "-path", "'*/.*'", "|", "fzf-tmux", "-h"
	).CombinedOutput() // "50%", "--preview", "'bat --color=always {}'")
	//out, err := cmd.CombinedOutput()
	log.Printf("Output1: %s", out)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Output: %s", out)
}

func main() {

	logfile, err := setupLogging()
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	cirApp := NewCirApplication()

	if err := cirApp.Run(); err != nil {
		panic(err)
	}
}
