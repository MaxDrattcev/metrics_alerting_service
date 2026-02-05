package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func parseServerFlags() (string, error) {
	var address = flag.String("a", "localhost:8080", "адрес и порт сервера")

	flag.Parse()

	if err := checkUnknownFlags(); err != nil {
		return "", err
	}

	return *address, nil
}

func checkUnknownFlags() error {
	knownFlags := make(map[string]bool)
	flag.VisitAll(func(f *flag.Flag) {
		knownFlags[f.Name] = true
	})

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		flagName := strings.TrimPrefix(arg, "-")
		flagName = strings.TrimPrefix(flagName, "-")

		if idx := strings.Index(flagName, "="); idx != -1 {
			flagName = flagName[:idx]
		}

		if !knownFlags[flagName] {
			return fmt.Errorf("unknown flag: -%s", flagName)
		}
	}

	return nil
}
