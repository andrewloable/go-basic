package runtime

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tempFile creates a temporary file in the test's temp directory and returns its path.
func tempFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// ---------------------------------------------------------------------------
// FileManager creation
// ---------------------------------------------------------------------------

func TestNewFileManager(t *testing.T) {
	fm := NewFileManager()
	if fm == nil {
		t.Fatal("NewFileManager() returned nil")
	}
	if len(fm.files) != 0 {
		t.Errorf("new FileManager has %d files, want 0", len(fm.files))
	}
}

// ---------------------------------------------------------------------------
// FileOpen / FileClose basics
// ---------------------------------------------------------------------------

func TestFileOpenClose(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "test.txt")

	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("FileOpen output: %v", err)
	}

	// Opening same number again should fail.
	err = fm.FileOpen(1, path, FileModeOutput, 0)
	if err == nil {
		t.Fatal("expected error opening already-open file number")
	}

	err = fm.FileClose(1)
	if err != nil {
		t.Fatalf("FileClose: %v", err)
	}

	// Closing again should fail.
	err = fm.FileClose(1)
	if err == nil {
		t.Fatal("expected error closing already-closed file")
	}
}

func TestFileCloseAll(t *testing.T) {
	fm := NewFileManager()
	p1 := tempFile(t, "f1.txt")
	p2 := tempFile(t, "f2.txt")

	fm.FileOpen(1, p1, FileModeOutput, 0)
	fm.FileOpen(2, p2, FileModeOutput, 0)

	err := fm.FileCloseAll()
	if err != nil {
		t.Fatalf("FileCloseAll: %v", err)
	}

	if len(fm.files) != 0 {
		t.Errorf("after FileCloseAll, %d files remain", len(fm.files))
	}
}

func TestFileCloseZeroClosesAll(t *testing.T) {
	fm := NewFileManager()
	p1 := tempFile(t, "f1.txt")
	p2 := tempFile(t, "f2.txt")

	fm.FileOpen(1, p1, FileModeOutput, 0)
	fm.FileOpen(2, p2, FileModeOutput, 0)

	err := fm.FileClose(0)
	if err != nil {
		t.Fatalf("FileClose(0): %v", err)
	}

	if len(fm.files) != 0 {
		t.Errorf("after FileClose(0), %d files remain", len(fm.files))
	}
}

// ---------------------------------------------------------------------------
// Sequential write + read (PRINT # / INPUT #)
// ---------------------------------------------------------------------------

func TestSequentialWriteRead(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "seq.txt")

	// Write
	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	err = fm.FileWrite(1, "Hello", 42, 3.14)
	if err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	fm.FileClose(1)

	// Read back
	err = fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("open input: %v", err)
	}

	v1, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput 1: %v", err)
	}
	if v1 != "Hello" {
		t.Errorf("value 1 = %q, want %q", v1, "Hello")
	}

	v2, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput 2: %v", err)
	}
	if strings.TrimSpace(v2) != "42" {
		t.Errorf("value 2 = %q, want %q", v2, "42")
	}

	v3, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput 3: %v", err)
	}
	if strings.TrimSpace(v3) != "3.14" {
		t.Errorf("value 3 = %q, want %q", v3, "3.14")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// FilePrint / FileLineInput
// ---------------------------------------------------------------------------

func TestFilePrintLineInput(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "lines.txt")

	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	fm.FilePrint(1, "Line one")
	fm.FilePrint(1, "Line two")
	fm.FileClose(1)

	err = fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("open input: %v", err)
	}

	l1, err := fm.FileLineInput(1)
	if err != nil {
		t.Fatalf("FileLineInput 1: %v", err)
	}
	if l1 != "Line one" {
		t.Errorf("line 1 = %q, want %q", l1, "Line one")
	}

	l2, err := fm.FileLineInput(1)
	if err != nil {
		t.Fatalf("FileLineInput 2: %v", err)
	}
	if l2 != "Line two" {
		t.Errorf("line 2 = %q, want %q", l2, "Line two")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// Append mode
// ---------------------------------------------------------------------------

func TestAppendMode(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "append.txt")

	// Create with some content.
	fm.FileOpen(1, path, FileModeOutput, 0)
	fm.FilePrint(1, "First")
	fm.FileClose(1)

	// Append more content.
	fm.FileOpen(1, path, FileModeAppend, 0)
	fm.FilePrint(1, "Second")
	fm.FileClose(1)

	// Read back both lines.
	fm.FileOpen(1, path, FileModeInput, 0)

	l1, _ := fm.FileLineInput(1)
	if l1 != "First" {
		t.Errorf("line 1 = %q, want %q", l1, "First")
	}
	l2, _ := fm.FileLineInput(1)
	if l2 != "Second" {
		t.Errorf("line 2 = %q, want %q", l2, "Second")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// FileInput on non-input file
// ---------------------------------------------------------------------------

func TestFileInputWrongMode(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "out.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)
	_, err := fm.FileInput(1)
	if err == nil {
		t.Fatal("expected error reading from output file")
	}
	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// Random Access: FIELD / LSET / RSET / GET / PUT
// ---------------------------------------------------------------------------

func TestRandomAccess(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "random.dat")

	recLen := 30
	err := fm.FileOpen(1, path, FileModeRandom, recLen)
	if err != nil {
		t.Fatalf("open random: %v", err)
	}

	// Define fields: name (20 bytes) + score (10 bytes).
	fields := []FieldDef{
		{Name: "name", Length: 20},
		{Name: "score", Length: 10},
	}
	err = fm.Field(1, fields)
	if err != nil {
		t.Fatalf("Field: %v", err)
	}

	// LSET name, RSET score.
	fm.Lset(1, "name", "Alice")
	fm.Rset(1, "score", "100")

	// PUT record 1.
	err = fm.RandomPut(1, 1)
	if err != nil {
		t.Fatalf("RandomPut: %v", err)
	}

	// Write a second record.
	fm.Lset(1, "name", "Bob")
	fm.Rset(1, "score", "200")
	fm.RandomPut(1, 2)

	// GET record 1 back.
	err = fm.RandomGet(1, 1)
	if err != nil {
		t.Fatalf("RandomGet: %v", err)
	}

	name, err := fm.GetFieldValue(1, "name")
	if err != nil {
		t.Fatalf("GetFieldValue name: %v", err)
	}
	// Should be left-justified, padded with spaces.
	if !strings.HasPrefix(name, "Alice") {
		t.Errorf("name = %q, want prefix \"Alice\"", name)
	}
	if len(name) != 20 {
		t.Errorf("name length = %d, want 20", len(name))
	}

	score, err := fm.GetFieldValue(1, "score")
	if err != nil {
		t.Fatalf("GetFieldValue score: %v", err)
	}
	// Should be right-justified.
	trimmed := strings.TrimSpace(score)
	if trimmed != "100" {
		t.Errorf("score = %q (trimmed %q), want \"100\"", score, trimmed)
	}

	// GET record 2.
	fm.RandomGet(1, 2)
	name, _ = fm.GetFieldValue(1, "name")
	if !strings.HasPrefix(name, "Bob") {
		t.Errorf("record 2 name = %q, want prefix \"Bob\"", name)
	}

	fm.FileClose(1)
}

func TestFieldExceedsRecordLength(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "field_err.dat")

	fm.FileOpen(1, path, FileModeRandom, 10)

	fields := []FieldDef{
		{Name: "a", Length: 6},
		{Name: "b", Length: 6}, // total 12 > 10
	}
	err := fm.Field(1, fields)
	if err == nil {
		t.Fatal("expected error for field total exceeding record length")
	}

	fm.FileClose(1)
}

func TestFieldOnNonRandomFile(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "seq.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)
	err := fm.Field(1, []FieldDef{{Name: "x", Length: 10}})
	if err == nil {
		t.Fatal("expected error for FIELD on non-random file")
	}
	fm.FileClose(1)
}

func TestGetFieldValueUnknownField(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "rf.dat")

	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "x", Length: 10}})
	_, err := fm.GetFieldValue(1, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown field name")
	}
	fm.FileClose(1)
}

func TestLsetRsetUnknownField(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "rf2.dat")

	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "x", Length: 10}})

	err := fm.Lset(1, "nonexistent", "val")
	if err == nil {
		t.Fatal("expected error for Lset unknown field")
	}
	err = fm.Rset(1, "nonexistent", "val")
	if err == nil {
		t.Fatal("expected error for Rset unknown field")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// Binary I/O
// ---------------------------------------------------------------------------

func TestBinaryIO(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "binary.dat")

	err := fm.FileOpen(1, path, FileModeBinary, 0)
	if err != nil {
		t.Fatalf("open binary: %v", err)
	}

	// Write data at position 1 (1-based).
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	err = fm.BinaryPut(1, 1, data)
	if err != nil {
		t.Fatalf("BinaryPut: %v", err)
	}

	// Read it back.
	got, err := fm.BinaryGet(1, 1, 5)
	if err != nil {
		t.Fatalf("BinaryGet: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("BinaryGet returned %d bytes, want 5", len(got))
	}
	for i, b := range data {
		if got[i] != b {
			t.Errorf("byte %d: got %02x, want %02x", i, got[i], b)
		}
	}

	// Write at a different position (1-based position 10).
	err = fm.BinaryPut(1, 10, []byte{0xAA, 0xBB})
	if err != nil {
		t.Fatalf("BinaryPut at offset: %v", err)
	}
	got2, err := fm.BinaryGet(1, 10, 2)
	if err != nil {
		t.Fatalf("BinaryGet at offset: %v", err)
	}
	if got2[0] != 0xAA || got2[1] != 0xBB {
		t.Errorf("data at offset 10: got %x, want [AA BB]", got2)
	}

	fm.FileClose(1)
}

func TestBinaryOnNonBinaryFile(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "seq.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)

	_, err := fm.BinaryGet(1, 1, 5)
	if err == nil {
		t.Fatal("expected error for BinaryGet on non-binary file")
	}
	err = fm.BinaryPut(1, 1, []byte{1})
	if err == nil {
		t.Fatal("expected error for BinaryPut on non-binary file")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// EOF
// ---------------------------------------------------------------------------

func TestEof(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "eof.txt")

	// Write some data.
	fm.FileOpen(1, path, FileModeOutput, 0)
	fm.FilePrint(1, "data")
	fm.FileClose(1)

	// Read and check EOF.
	fm.FileOpen(1, path, FileModeInput, 0)

	eof, err := fm.Eof(1)
	if err != nil {
		t.Fatalf("Eof: %v", err)
	}
	if eof {
		t.Error("expected EOF=false before reading")
	}

	fm.FileLineInput(1)

	eof, err = fm.Eof(1)
	if err != nil {
		t.Fatalf("Eof after read: %v", err)
	}
	if !eof {
		t.Error("expected EOF=true after reading all data")
	}

	fm.FileClose(1)
}

func TestEofBinaryMode(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "eof_bin.dat")

	// Write 5 bytes.
	fm.FileOpen(1, path, FileModeBinary, 0)
	fm.BinaryPut(1, 1, []byte{1, 2, 3, 4, 5})

	// Seek to end.
	fm.FileSeek(1, 6) // past the 5 bytes

	eof, err := fm.Eof(1)
	if err != nil {
		t.Fatalf("Eof binary: %v", err)
	}
	if !eof {
		t.Error("expected EOF at end of binary file")
	}

	// Seek to start.
	fm.FileSeek(1, 1)
	eof, _ = fm.Eof(1)
	if eof {
		t.Error("expected not EOF at start of binary file")
	}

	fm.FileClose(1)
}

func TestEofFileNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Eof(99)
	if err == nil {
		t.Fatal("expected error for EOF on non-open file")
	}
}

// ---------------------------------------------------------------------------
// LOC
// ---------------------------------------------------------------------------

func TestLoc(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "loc.dat")

	fm.FileOpen(1, path, FileModeBinary, 0)
	fm.BinaryPut(1, 1, []byte("Hello"))

	pos, err := fm.Loc(1)
	if err != nil {
		t.Fatalf("Loc: %v", err)
	}
	// After writing 5 bytes, position should be 5.
	if pos != 5 {
		t.Errorf("Loc = %d, want 5", pos)
	}

	fm.FileClose(1)
}

func TestLocRandom(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "loc_rand.dat")

	recLen := 10
	fm.FileOpen(1, path, FileModeRandom, recLen)

	// At start, position / recLen should be 0.
	pos, err := fm.Loc(1)
	if err != nil {
		t.Fatalf("Loc: %v", err)
	}
	if pos != 0 {
		t.Errorf("Loc at start = %d, want 0", pos)
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// LOF
// ---------------------------------------------------------------------------

func TestLof(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "lof.dat")

	fm.FileOpen(1, path, FileModeBinary, 0)
	fm.BinaryPut(1, 1, []byte("1234567890"))

	size, err := fm.Lof(1)
	if err != nil {
		t.Fatalf("Lof: %v", err)
	}
	if size != 10 {
		t.Errorf("Lof = %d, want 10", size)
	}

	fm.FileClose(1)
}

func TestLofEmpty(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "lof_empty.dat")

	fm.FileOpen(1, path, FileModeOutput, 0)

	size, err := fm.Lof(1)
	if err != nil {
		t.Fatalf("Lof: %v", err)
	}
	if size != 0 {
		t.Errorf("Lof empty = %d, want 0", size)
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// FileSeek / FileSeekPos
// ---------------------------------------------------------------------------

func TestFileSeekBinary(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "seek.dat")

	fm.FileOpen(1, path, FileModeBinary, 0)
	fm.BinaryPut(1, 1, []byte("ABCDEFGHIJ"))

	// Seek to position 5 (1-based).
	err := fm.FileSeek(1, 5)
	if err != nil {
		t.Fatalf("FileSeek: %v", err)
	}

	pos, err := fm.FileSeekPos(1)
	if err != nil {
		t.Fatalf("FileSeekPos: %v", err)
	}
	if pos != 5 {
		t.Errorf("FileSeekPos = %d, want 5", pos)
	}

	// Read from that position.
	data, err := fm.BinaryGet(1, 5, 3)
	if err != nil {
		t.Fatalf("BinaryGet: %v", err)
	}
	if string(data) != "EFG" {
		t.Errorf("data at pos 5 = %q, want %q", string(data), "EFG")
	}

	fm.FileClose(1)
}

func TestFileSeekRandom(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "seek_rand.dat")

	recLen := 10
	fm.FileOpen(1, path, FileModeRandom, recLen)

	// Seek to record 3 (1-based).
	err := fm.FileSeek(1, 3)
	if err != nil {
		t.Fatalf("FileSeek: %v", err)
	}

	pos, err := fm.FileSeekPos(1)
	if err != nil {
		t.Fatalf("FileSeekPos: %v", err)
	}
	if pos != 3 {
		t.Errorf("FileSeekPos = %d, want 3", pos)
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// RandomGet / RandomPut on non-random file
// ---------------------------------------------------------------------------

func TestRandomGetPutWrongMode(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "wrong_mode.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)

	err := fm.RandomGet(1, 1)
	if err == nil {
		t.Fatal("expected error for RandomGet on non-random file")
	}
	err = fm.RandomPut(1, 1)
	if err == nil {
		t.Fatal("expected error for RandomPut on non-random file")
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// Default record length
// ---------------------------------------------------------------------------

func TestDefaultRecordLength(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "defreclen.dat")

	err := fm.FileOpen(1, path, FileModeRandom, 0)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}

	bf, _ := fm.getFile(1)
	if bf.RecLen != 128 {
		t.Errorf("default RecLen = %d, want 128", bf.RecLen)
	}

	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// MkiBytes / CviBytes etc. (byte-level helpers from fileio.go)
// ---------------------------------------------------------------------------

func TestMkiBytesCviBytes(t *testing.T) {
	tests := []int16{0, 1, -1, 32767, -32768}
	for _, v := range tests {
		b := MkiBytes(v)
		if len(b) != 2 {
			t.Errorf("MkiBytes(%d) len = %d, want 2", v, len(b))
		}
		got := CviBytes(b)
		if got != v {
			t.Errorf("CviBytes(MkiBytes(%d)) = %d", v, got)
		}
	}
}

func TestCviBytesTooShort(t *testing.T) {
	got := CviBytes([]byte{1})
	if got != 0 {
		t.Errorf("CviBytes short = %d, want 0", got)
	}
}

func TestMklBytesCvlBytes(t *testing.T) {
	tests := []int32{0, 1, -1, 2147483647, -2147483648}
	for _, v := range tests {
		b := MklBytes(v)
		if len(b) != 4 {
			t.Errorf("MklBytes(%d) len = %d, want 4", v, len(b))
		}
		got := CvlBytes(b)
		if got != v {
			t.Errorf("CvlBytes(MklBytes(%d)) = %d", v, got)
		}
	}
}

func TestCvlBytesTooShort(t *testing.T) {
	got := CvlBytes([]byte{1, 2})
	if got != 0 {
		t.Errorf("CvlBytes short = %d, want 0", got)
	}
}

func TestMksBytesCvsBytes(t *testing.T) {
	tests := []float32{0, 1, -1, 3.14, -273.15}
	for _, v := range tests {
		b := MksBytes(v)
		if len(b) != 4 {
			t.Errorf("MksBytes(%g) len = %d, want 4", v, len(b))
		}
		got := CvsBytes(b)
		if got != v {
			t.Errorf("CvsBytes(MksBytes(%g)) = %g", v, got)
		}
	}
}

func TestCvsBytesTooShort(t *testing.T) {
	got := CvsBytes([]byte{1})
	if got != 0 {
		t.Errorf("CvsBytes short = %g, want 0", got)
	}
}

func TestMkdBytesCvdBytes(t *testing.T) {
	tests := []float64{0, 1, -1, math.Pi, -273.15, 1e100}
	for _, v := range tests {
		b := MkdBytes(v)
		if len(b) != 8 {
			t.Errorf("MkdBytes(%g) len = %d, want 8", v, len(b))
		}
		got := CvdBytes(b)
		if got != v {
			t.Errorf("CvdBytes(MkdBytes(%g)) = %g", v, got)
		}
	}
}

func TestCvdBytesTooShort(t *testing.T) {
	got := CvdBytes([]byte{1, 2, 3, 4})
	if got != 0 {
		t.Errorf("CvdBytes short = %g, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// Open non-existent file for input
// ---------------------------------------------------------------------------

func TestOpenNonExistentForInput(t *testing.T) {
	fm := NewFileManager()
	err := fm.FileOpen(1, filepath.Join(t.TempDir(), "no_such_file.txt"), FileModeInput, 0)
	if err == nil {
		t.Fatal("expected error opening non-existent file for input")
	}
}

// ---------------------------------------------------------------------------
// WRITE # with quoted strings containing quotes
// ---------------------------------------------------------------------------

func TestFileWriteQuotedStrings(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "quoted.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)
	err := fm.FileWrite(1, "He said \"hi\"", 42)
	if err != nil {
		t.Fatalf("FileWrite: %v", err)
	}
	fm.FileClose(1)

	// Verify the file content.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	content := string(data)
	// Should contain doubled quotes.
	if !strings.Contains(content, "\"\"") {
		t.Errorf("content = %q, expected doubled quotes", content)
	}
	if !strings.Contains(content, "42") {
		t.Errorf("content = %q, expected 42", content)
	}
}

// ---------------------------------------------------------------------------
// FileLineInput on non-input file
// ---------------------------------------------------------------------------

func TestFileLineInputWrongMode(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "outonly.txt")

	fm.FileOpen(1, path, FileModeOutput, 0)
	_, err := fm.FileLineInput(1)
	if err == nil {
		t.Fatal("expected error for LineInput on output file")
	}
	fm.FileClose(1)
}

// ---------------------------------------------------------------------------
// getFile on non-open file number
// ---------------------------------------------------------------------------

func TestGetFileNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.getFile(99)
	if err == nil {
		t.Fatal("expected error for non-open file")
	}
}

// ---------------------------------------------------------------------------
// getWriter — both branches (Writer nil and Writer already set)
// ---------------------------------------------------------------------------

func TestGetWriterNilBranch(t *testing.T) {
	// FileOpen for Output pre-sets bf.Writer, so to hit the nil branch we use
	// FileModeBinary which does NOT pre-set Writer, then call FileWrite.
	fm := NewFileManager()
	path := tempFile(t, "getwriter_nil.dat")

	err := fm.FileOpen(1, path, FileModeBinary, 0)
	if err != nil {
		t.Fatalf("FileOpen binary: %v", err)
	}
	defer fm.FileClose(1)

	// FileWrite calls getWriter; bf.Writer is nil → creates new writer.
	if err := fm.FileWrite(1, "first"); err != nil {
		t.Fatalf("FileWrite (nil writer): %v", err)
	}
	// Second call: bf.Writer is now non-nil → returns existing writer.
	if err := fm.FileWrite(1, "second"); err != nil {
		t.Fatalf("FileWrite (existing writer): %v", err)
	}
}

func TestGetWriterNonNilBranch(t *testing.T) {
	// For Output mode, FileOpen pre-sets Writer (non-nil path immediately).
	fm := NewFileManager()
	path := tempFile(t, "getwriter_nonnull.txt")

	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("FileOpen output: %v", err)
	}
	defer fm.FileClose(1)

	// Both calls hit the non-nil branch (Writer was set by FileOpen).
	fm.FilePrint(1, "hello")
	fm.FilePrint(1, "world")
}

// ---------------------------------------------------------------------------
// FileInput — quoted strings, CRLF, doubled quotes
// ---------------------------------------------------------------------------

func TestFileInputQuotedString(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "quoted.txt")

	// Write a quoted CSV-style field.
	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("FileOpen output: %v", err)
	}
	fm.FileWrite(1, "hello world")
	fm.FileClose(1)

	// Now read it back via FileInput.
	err = fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("FileOpen input: %v", err)
	}
	defer fm.FileClose(1)

	got, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput quoted: %v", err)
	}
	if got != "hello world" {
		t.Errorf("FileInput quoted = %q, want %q", got, "hello world")
	}
}

func TestFileInputDoubledQuote(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "dquote.txt")

	// Write a string that contains a quote (FileWrite uses doubled-quote escaping).
	err := fm.FileOpen(1, path, FileModeOutput, 0)
	if err != nil {
		t.Fatalf("FileOpen output: %v", err)
	}
	fm.FileWrite(1, `say "hi"`)
	fm.FileClose(1)

	err = fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("FileOpen input: %v", err)
	}
	defer fm.FileClose(1)

	got, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput doubled quote: %v", err)
	}
	if got != `say "hi"` {
		t.Errorf("FileInput doubled quote = %q, want %q", got, `say "hi"`)
	}
}

func TestFileInputCRLF(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "crlf.txt")

	// Write a file with CRLF line endings manually.
	if err := os.WriteFile(path, []byte("line1\r\nline2\r\n"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer fm.FileClose(1)

	got, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput CRLF first: %v", err)
	}
	if got != "line1" {
		t.Errorf("FileInput CRLF = %q, want %q", got, "line1")
	}
}

func TestFileInputCROnly(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "cronly.txt")

	// CR alone as line terminator.
	if err := os.WriteFile(path, []byte("lineA\rlineB\r"), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer fm.FileClose(1)

	got, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput CR-only: %v", err)
	}
	if got != "lineA" {
		t.Errorf("FileInput CR-only = %q, want %q", got, "lineA")
	}
}

func TestFileInputEOFInsideQuote(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "eofinquote.txt")

	// Quoted string that has EOF before closing quote.
	if err := os.WriteFile(path, []byte(`"unterminated`), 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	err := fm.FileOpen(1, path, FileModeInput, 0)
	if err != nil {
		t.Fatalf("FileOpen: %v", err)
	}
	defer fm.FileClose(1)

	got, err := fm.FileInput(1)
	if err != nil {
		t.Fatalf("FileInput EOF in quote: %v", err)
	}
	// Should return whatever was read before EOF.
	if got != "unterminated" {
		t.Errorf("FileInput EOF in quote = %q, want %q", got, "unterminated")
	}
}

// ---------------------------------------------------------------------------
// Lset / Rset — num == 0 (search all files) and value longer than field
// ---------------------------------------------------------------------------

func TestLsetNumZeroFindsField(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "lset0.dat")

	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "name", Length: 10}})

	// num == 0: search all open files.
	err := fm.Lset(0, "name", "Alice")
	if err != nil {
		t.Fatalf("Lset(0,...): %v", err)
	}
	got, _ := fm.GetFieldValue(1, "name")
	// Should be left-justified, padded with spaces.
	if !strings.HasPrefix(got, "Alice") {
		t.Errorf("Lset(0) result = %q, want prefix 'Alice'", got)
	}
	fm.FileClose(1)
}

func TestLsetNumZeroNotFound(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "lset0nf.dat")
	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "x", Length: 10}})

	err := fm.Lset(0, "nonexistent", "val")
	if err == nil {
		t.Fatal("Lset(0,...) with no matching field should return error")
	}
	fm.FileClose(1)
}

func TestRsetNumZeroFindsField(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "rset0.dat")

	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "city", Length: 10}})

	err := fm.Rset(0, "city", "NY")
	if err != nil {
		t.Fatalf("Rset(0,...): %v", err)
	}
	got, _ := fm.GetFieldValue(1, "city")
	if !strings.HasSuffix(strings.TrimRight(got, " "), "NY") && !strings.Contains(got, "NY") {
		t.Errorf("Rset(0) result = %q, want 'NY' right-justified", got)
	}
	fm.FileClose(1)
}

func TestRsetNumZeroNotFound(t *testing.T) {
	fm := NewFileManager()
	path := tempFile(t, "rset0nf.dat")
	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "x", Length: 10}})

	err := fm.Rset(0, "nonexistent", "val")
	if err == nil {
		t.Fatal("Rset(0,...) with no matching field should return error")
	}
	fm.FileClose(1)
}

func TestRsetValueLongerThanField(t *testing.T) {
	// Value longer than field → truncate from end (start < 0 branch).
	fm := NewFileManager()
	path := tempFile(t, "rset_long.dat")

	fm.FileOpen(1, path, FileModeRandom, 20)
	fm.Field(1, []FieldDef{{Name: "code", Length: 3}})

	err := fm.Rset(1, "code", "toolongvalue")
	if err != nil {
		t.Fatalf("Rset long value: %v", err)
	}
	got, _ := fm.GetFieldValue(1, "code")
	if len(got) != 3 {
		t.Errorf("Rset long value: field length = %d, want 3", len(got))
	}
	fm.FileClose(1)
}

func TestLsetNoFile(t *testing.T) {
	fm := NewFileManager()
	err := fm.Lset(1, "field", "value")
	if err == nil {
		t.Error("expected error for Lset on closed file")
	}
}

func TestRsetNoFile(t *testing.T) {
	fm := NewFileManager()
	err := fm.Rset(1, "field", "value")
	if err == nil {
		t.Error("expected error for Rset on closed file")
	}
}

func TestFileWriteNotOpen(t *testing.T) {
	fm := NewFileManager()
	err := fm.FileWrite(99, []interface{}{"test"})
	if err == nil {
		t.Error("expected error writing to unopened file")
	}
}

func TestFilePrintNotOpen(t *testing.T) {
	fm := NewFileManager()
	err := fm.FilePrint(99, " hello")
	if err == nil {
		t.Error("expected error printing to unopened file")
	}
}

func TestEofNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Eof(99)
	if err == nil {
		t.Error("expected error for Eof on closed file")
	}
}

func TestLocNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Loc(99)
	if err == nil {
		t.Error("expected error for Loc on closed file")
	}
}

func TestLofNotOpen(t *testing.T) {
	fm := NewFileManager()
	_, err := fm.Lof(99)
	if err == nil {
		t.Error("expected error for Lof on closed file")
	}
}
