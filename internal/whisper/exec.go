package whisper

import (
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"sort"

	"github.com/busser/whisper/internal/environ"
)

// Modified during testing to catch command output.
var (
	execOut io.Writer = os.Stdout
	execErr io.Writer = os.Stderr
)

func Exec(name string, args ...string) (exitCode int, err error) {
	originalVars := environ.ToMap(os.Environ())

	newVars, err := ResolveAll(originalVars)
	if err != nil {
		return 0, err
	}

	var overloaded []string
	for name, original := range originalVars {
		if newVars[name] != original {
			overloaded = append(overloaded, name)
		}
	}

	sort.Strings(overloaded)
	for _, name := range overloaded {
		log.Printf("[whisper] overloading %s", name)
	}

	subCmd := exec.Command(name, args...)
	subCmd.Env = environ.ToSlice(newVars)
	subCmd.Stdin = os.Stdin
	subCmd.Stdout = execOut
	subCmd.Stderr = execErr

	if err := subCmd.Run(); err != nil {
		exitErr := new(exec.ExitError)
		if errors.As(err, &exitErr) {
			return exitErr.ProcessState.ExitCode(), nil
		}
		return 0, err
	}

	return subCmd.ProcessState.ExitCode(), nil
}
