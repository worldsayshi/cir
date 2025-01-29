package main

import (
	"flag"
	"fmt"
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
	versionFlag := flag.Bool("version", false, "print the version and commit id")
	sessionFile := flag.String("session", path.Join(homedir, ".cir/default-session.yaml"), "path to the session file")
	flag.Parse()

	if *versionFlag {
		versionData, err := os.ReadFile("version.txt")
		if err != nil {
			fmt.Printf("Error reading version file: %v\n", err)
			return
		}
		fmt.Printf("Version Information:\n%s\n", string(versionData))
		return
	}

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
