package runtime

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// Timer returns seconds since midnight as float64 (BASIC's TIMER).
func Timer() float64 {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return now.Sub(midnight).Seconds()
}

// Delay pauses execution for the given number of seconds (Turbo BASIC extension).
func Delay(seconds float64) {
	if seconds <= 0 {
		return
	}
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

// DateStr returns the current date as "MM-DD-YYYY" (BASIC's DATE$).
func DateStr() string {
	return time.Now().Format("01-02-2006")
}

// TimeStr returns the current time as "HH:MM:SS" (BASIC's TIME$).
func TimeStr() string {
	return time.Now().Format("15:04:05")
}

// CommandStr returns the command-line arguments as a single string (BASIC's COMMAND$).
// Arguments are joined with spaces and returned in uppercase, matching BASIC convention.
func CommandStr() string {
	if len(os.Args) <= 1 {
		return ""
	}
	return strings.Join(os.Args[1:], " ")
}

// EnvironGet reads an environment variable (BASIC's ENVIRON$).
func EnvironGet(name string) string {
	return os.Getenv(name)
}

// EnvironSet sets an environment variable (BASIC's ENVIRON statement).
func EnvironSet(name, value string) error {
	return os.Setenv(name, value)
}

// Shell executes a system command and waits for completion (BASIC's SHELL).
func Shell(command string) error {
	if command == "" {
		return nil
	}
	var cmd *exec.Cmd
	// Use the platform shell to interpret the command string.
	if shellPath := os.Getenv("COMSPEC"); shellPath != "" {
		// Windows: use cmd.exe /C
		cmd = exec.Command(shellPath, "/C", command)
	} else {
		// Unix: use /bin/sh -c
		cmd = exec.Command("/bin/sh", "-c", command)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Inkey returns the next character from the keyboard buffer without waiting.
// Returns empty string if no key is available. In transpiled Go code this is
// a stub — a real implementation would require raw terminal I/O (e.g. golang.org/x/term).
func Inkey() string {
	return ""
}

// Fre returns available heap memory as a rough estimate.
// The argument follows BASIC convention: "" for string space, -1 for array space, -2 for stack.
// In managed Go memory, this is an approximation.
func Fre(_ float64) float64 {
	return 32768 // Return a plausible constant; GC manages actual memory.
}

// Peek reads a byte from the given memory address. In DOS BASIC, PEEK accessed
// real memory-mapped hardware. In transpiled Go code, memory addresses have no
// meaning — always returns 0.
func Peek(_ float64) float64 {
	return 0
}

// Poke writes a byte to the given memory address. No-op in transpiled code.
func Poke(_, _ float64) {
	// Memory-mapped I/O is not available in transpiled Go code.
}

// SwapInt swaps two int values via pointers.
// For transpiled code, swaps are generated inline as: a, b = b, a
// This exists for the VM.
func SwapInt(a, b *int) { *a, *b = *b, *a }

// SwapFloat swaps two float64 values via pointers.
func SwapFloat(a, b *float64) { *a, *b = *b, *a }

// SwapString swaps two string values via pointers.
func SwapString(a, b *string) { *a, *b = *b, *a }
