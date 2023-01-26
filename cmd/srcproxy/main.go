package main

import (
	"fmt"
	"github.com/titanous/json5"
	"log"
	"os"
)

func showUsage() {
	_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [client|server] config.json\n", os.Args[0])
}

func main() {
	if len(os.Args) != 3 {
		showUsage()
		os.Exit(22)
	}
	mode := os.Args[1]
	configPath := os.Args[2]

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("[fatal] failed to open config file: %v\n", err)
		return
	}

	switch mode {
	case "server":
		config := &ServerConfig{}
		err = json5.NewDecoder(configFile).Decode(config)
		configFile.Close()
		if err != nil {
			log.Fatalf("[fatal] failed to parse config file: %v\n", err)
			return
		}
		log.Fatal(runServer(config))
	case "client":
		config := &ClientConfig{}
		err = json5.NewDecoder(configFile).Decode(config)
		configFile.Close()
		if err != nil {
			log.Fatalf("[fatal] failed to parse config file: %v\n", err)
			return
		}
		log.Fatal(runClient(config))
	default:
		configFile.Close()
		showUsage()
		os.Exit(22)
	}
}
