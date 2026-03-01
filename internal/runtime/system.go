// system.go — System-level built-ins for the BASIC runtime.
//
// # Compiler Design Note: Mapping OS Calls to a New Platform
//
// Turbo BASIC ran on MS-DOS and interacted with the OS through DOS interrupts
// and BIOS calls. Many of these operations have direct equivalents in Go's
// standard library, but some (PEEK/POKE, memory segment access) have no
// meaningful translation to a modern OS.
//
// The pattern used here is:
//
//   - For functions with a clear modern equivalent (TIMER, DATE$, TIME$,
//     SHELL, ENVIRON$), wrap the Go standard library call directly.
//   - For functions that accessed DOS-specific hardware memory (PEEK, POKE),
//     provide a stub that returns a safe neutral value and is documented as
//     a no-op. This keeps the generated program compilable even if the BASIC
//     source used PEEK/POKE for hardware tricks that cannot run on modern OS.
//   - For functions with no equivalent (FRE — heap space query), return a
//     plausible constant so programs that check available memory do not fail.
//
// This "stub with documentation" strategy is a common approach in emulation and
// cross-compilation: preserve the API surface so the generated code compiles,
// but acknowledge that the semantics cannot be faithfully reproduced.

package runtime

import (
	"fmt"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"
)

// Timer returns the number of seconds elapsed since midnight as a float64.
// BASIC: TIMER — on DOS this read the BIOS tick counter (18.2 ticks/second)
// and divided by 18.2. Here we use Go's time package for exact sub-second
// resolution. Programs that poll TIMER for timing loops or random seeds work
// correctly with this implementation.
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

// LoopYield inserts a 1ms pause in tight loops (DO/LOOP, WHILE/WEND).
// Compiled Go runs orders of magnitude faster than interpreted BASIC, so
// busy-wait loops that calibrate timing (like GORILLA.BAS's CalcDelay)
// would produce absurdly large values without this throttle. The 1ms
// sleep also prevents tight loops from consuming 100% CPU.
func LoopYield() {
	time.Sleep(time.Millisecond)
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

// Fre returns an estimate of available memory (BASIC's FRE function).
// BASIC: FRE(0) or FRE("") — on DOS this returned the number of free bytes in the
// BASIC heap or string space. Go's garbage collector manages memory automatically,
// so we report approximate free heap memory from runtime.MemStats.
func Fre(_ float64) float64 {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)
	// Approximate free memory: total system memory minus heap in use.
	free := m.Sys - m.HeapInuse
	// Cap to a reasonable range for BASIC programs (max ~2GB).
	const maxFree = 2 * 1024 * 1024 * 1024 // 2 GB
	if free > maxFree {
		free = maxFree
	}
	return float64(free)
}

// Peek reads a byte from a DOS memory address — stub that always returns 0.
//
// BASIC: PEEK(address) — in DOS BASIC, memory was a flat 20-bit address space
// shared between the OS, the BASIC interpreter, and hardware registers. Common
// uses included reading the keyboard buffer (segment 0x40), checking video
// memory, or probing BIOS data areas.
//
// In a modern OS process, the address space is virtual and protected. There is
// no meaningful way to access arbitrary addresses, so this function returns 0.
// BASIC programs that use PEEK only for timing or display detection will still
// compile; programs that critically depend on specific hardware memory layouts
// will need manual porting.
func Peek(addr float64) float64 {
	peekPokeWarnOnce()
	return 0
}

// Poke writes a byte to a DOS memory address — stub that does nothing.
//
// BASIC: POKE address, value — the write-side companion to PEEK. Classic uses
// included directly modifying video memory for fast screen output, setting
// keyboard flags, or altering interrupt vectors.
//
// For the same reasons as Peek, this cannot be implemented meaningfully in a
// modern OS process and is a documented no-op.
func Poke(addr, val float64) {
	peekPokeWarnOnce()
}

var peekPokeWarned bool

// peekPokeWarnOnce prints a one-time warning to stderr when GOBASIC_PEEK_WARN=1.
func peekPokeWarnOnce() {
	if peekPokeWarned {
		return
	}
	peekPokeWarned = true
	if os.Getenv("GOBASIC_PEEK_WARN") == "1" {
		fmt.Fprintln(os.Stderr, "Warning: PEEK/POKE — DOS memory access is not available in transpiled code. These calls are no-ops.")
	}
}

// SwapInt swaps two int values via pointers.
// For transpiled code, swaps are generated inline as: a, b = b, a
// This exists for the VM.
func SwapInt(a, b *int) { *a, *b = *b, *a }

// SwapFloat swaps two float64 values via pointers.
func SwapFloat(a, b *float64) { *a, *b = *b, *a }

// SwapString swaps two string values via pointers.
func SwapString(a, b *string) { *a, *b = *b, *a }
