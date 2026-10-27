package rhsm

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mizdebsk/radii/internal/log"
)

func parseRepoFile(path string) (map[string]map[string]string, error) {
	f, err := os.Open(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warnf("failed to close file %s: %v", path, err)
		}
	}()

	sections := make(map[string]map[string]string)
	section := ""

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			if _, exists := sections[section]; !exists {
				sections[section] = make(map[string]string)
			}
			continue
		}
		if section == "" {
			continue
		}

		key, val, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		k := strings.ToLower(strings.TrimSpace(key))
		v := strings.TrimSpace(val)
		if idx := strings.Index(v, "#"); idx >= 0 {
			v = strings.TrimSpace(v[:idx])
		}
		sections[section][k] = v
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return sections, nil
}

func readRepoStates(path string) (map[string]bool, error) {
	sections, err := parseRepoFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read repository definitions from %s: %w", path, err)
	}
	if sections == nil {
		return nil, nil
	}
	states := make(map[string]bool)
	for name, kv := range sections {
		val := strings.ToLower(kv["enabled"])
		states[name] = val == "1" || val == "true" || val == "yes" || val == "on"
	}
	return states, nil
}
