package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// loadEnv reads from '.env' in current path and sets the key as env variable.
func loadEnv() error {
	f, err := os.Open(".env")
	if err != nil {
		return fmt.Errorf("load '.env': %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for lineNo := 0; scanner.Scan(); lineNo++ {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			return fmt.Errorf("invalid 'key=value' pair at %d", lineNo)
		}
		os.Setenv(key, value)
	}

	return scanner.Err()
}

// loadConsumerMapping loads 'consumer-mapping.json' file from disk.
func loadConsumerMapping() (map[string]string, error) {
	f, err := os.Open(".consumer-mapping.json")
	if err != nil {
		return nil, fmt.Errorf("load '.consumer-mapping.json': %w", err)
	}
	defer f.Close()

	mapping := make(map[string]string)
	dec := json.NewDecoder(f)
	if err = dec.Decode(&mapping); err != nil {
		return nil, fmt.Errorf("parse '.consumer-mapping.json': %w", err)
	}

	return mapping, nil
}
