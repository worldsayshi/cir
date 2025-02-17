package main

import (
	"flag"
	"log"
	"os"
	"path"
)

func setupLogging() (f *os.File, err error) {
	f, err = os.OpenFile("cir.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
		return nil, err
	}

	log.SetOutput(f)
	return f, nil
}

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	sessionFile := flag.String("session", path.Join(homedir, ".cir/default-session.yaml"), "path to the session file")
	flag.Parse()

	logfile, err := setupLogging()
	if err != nil {
		panic(err)
	}
	defer logfile.Close()

	cirApp := NewCirApplication(*sessionFile)

	if err := cirApp.Run(); err != nil {
		panic(err)
	}
}
