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

func readRepoStates(path string) (map[string]bool, error) {
	f, err := os.Open(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read repository definitions from %s: %w", path, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warnf("failed to close file %s: %v", path, err)
		}
	}()

	sc := bufio.NewScanner(f)
	states := make(map[string]bool)
	section := ""

	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.TrimSpace(line[1 : len(line)-1])
			states[section] = false
			continue
		}
		if section == "" {
			continue
		}

		key, val, found := strings.Cut(line, "=")
		if found && strings.EqualFold(strings.TrimSpace(key), "enabled") {
			val = strings.SplitN(val, "#", 2)[0]
			val = strings.ToLower(strings.TrimSpace(val))
			states[section] = val == "1" || val == "true" || val == "yes" || val == "on"
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to read repository definitions from %s: %w", path, err)
	}
	return states, nil
}
