// fileio.go — File I/O built-ins for the BASIC runtime.
//
// # Compiler Design Note: BASIC's File I/O Model
//
// Turbo BASIC supports three distinct file access styles, each opened with a
// different mode keyword:
//
//  1. Sequential (INPUT / OUTPUT / APPEND)
//     Files are read or written as a stream of comma-delimited or newline-
//     delimited text values. Think CSV. PRINT # writes, INPUT # reads.
//
//  2. Random access (RANDOM)
//     Files are divided into fixed-length records. FIELD maps variable names
//     to byte ranges within a record buffer. GET # reads a record by number;
//     PUT # writes one. This was BASIC's way of implementing flat-file databases.
//
//  3. Binary (BINARY)
//     Raw byte access at arbitrary positions. GET # / PUT # work on byte arrays.
//
// # File Handle Table (FileManager)
//
// BASIC identifies files by small integers called file numbers or handles:
//
//	OPEN "data.txt" FOR INPUT AS #1
//	LINE INPUT #1, a$
//	CLOSE #1
//
// The FileManager struct maintains a map[int]*BasicFile that plays the same
// role as the OS file descriptor table — it maps BASIC's numeric handles to
// the underlying Go *os.File objects plus associated buffering state.
//
// This is the same abstraction used in the C runtime (FILE*) and Python's
// io module: the language runtime owns the handle table so that the user
// program cannot forge or corrupt file state.
//
// # Generated Code Integration
//
// The code generator declares a *FileManager global in each generated program:
//
//	var _fm = rt.NewFileManager()
//
// OPEN, CLOSE, GET, PUT, PRINT #, INPUT # are all translated to method calls
// on _fm, e.g. _fm.FileOpen(1, "data.txt", rt.FileModeInput, 0).

package runtime

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// FileMode represents the mode a file was opened in.
type FileMode int

const (
	FileModeInput  FileMode = iota // sequential read
	FileModeOutput                 // sequential write (truncate)
	FileModeAppend                 // sequential write (append)
	FileModeRandom                 // random access (fixed-length records)
	FileModeBinary                 // raw binary access
)

// BasicFile represents an open file in the BASIC runtime.
type BasicFile struct {
	Handle   *os.File
	Mode     FileMode
	RecLen   int    // record length for RANDOM mode (default 128)
	FieldBuf []byte // field buffer for RANDOM access
	Fields   []FieldDef
	Reader   *bufio.Reader
	Writer   *bufio.Writer
	Number   int
}

// FieldDef describes a FIELD variable mapping within a record buffer.
type FieldDef struct {
	Name   string
	Offset int
	Length int
}

// FileManager manages all open BASIC files indexed by their BASIC file numbers.
//
// The map key is the BASIC file number (e.g. 1 for OPEN … AS #1). Using a map
// rather than a slice means file numbers can be arbitrary integers and there is
// no fixed upper limit. CLOSE #0 (or CLOSE with no argument) closes all files
// via FileCloseAll, mirroring DOS BASIC behaviour on program exit.
type FileManager struct {
	files map[int]*BasicFile
}

// NewFileManager creates a new FileManager.
func NewFileManager() *FileManager {
	return &FileManager{
		files: make(map[int]*BasicFile),
	}
}

// FileOpen opens a file, registers it under the given BASIC file number, and
// sets up buffered readers/writers appropriate for the access mode.
//
// BASIC: OPEN name FOR mode AS #num [LEN = recLen]
//
// The recLen parameter is only meaningful for FileModeRandom; for other modes
// it is ignored and defaults to 128 if zero (DOS BASIC's default record size).
// Opening a number that is already in use is an error (File already open).
func (fm *FileManager) FileOpen(num int, name string, mode FileMode, recLen int) error {
	if _, exists := fm.files[num]; exists {
		return fmt.Errorf("file #%d already open", num)
	}
	if recLen <= 0 {
		recLen = 128
	}

	var f *os.File
	var err error

	switch mode {
	case FileModeInput:
		f, err = os.Open(name)
	case FileModeOutput:
		f, err = os.Create(name)
	case FileModeAppend:
		f, err = os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	case FileModeRandom:
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	case FileModeBinary:
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE, 0644)
	default:
		return fmt.Errorf("invalid file mode %d", mode)
	}

	if err != nil {
		return err
	}

	bf := &BasicFile{
		Handle: f,
		Mode:   mode,
		RecLen: recLen,
		Number: num,
	}

	if mode == FileModeRandom {
		bf.FieldBuf = make([]byte, recLen)
	}

	if mode == FileModeInput {
		bf.Reader = bufio.NewReader(f)
	}
	if mode == FileModeOutput || mode == FileModeAppend {
		bf.Writer = bufio.NewWriter(f)
	}

	fm.files[num] = bf
	return nil
}

// FileClose closes the file with the given BASIC file number.
// BASIC: CLOSE #num — flushes any buffered output before closing.
// If num is 0, all open files are closed (CLOSE with no argument).
func (fm *FileManager) FileClose(num int) error {
	if num == 0 {
		return fm.FileCloseAll()
	}

	bf, ok := fm.files[num]
	if !ok {
		return fmt.Errorf("file #%d not open", num)
	}

	if bf.Writer != nil {
		bf.Writer.Flush()
	}
	err := bf.Handle.Close()
	delete(fm.files, num)
	return err
}

// FileCloseAll closes all open files.
func (fm *FileManager) FileCloseAll() error {
	var firstErr error
	for num, bf := range fm.files {
		if bf.Writer != nil {
			bf.Writer.Flush()
		}
		if err := bf.Handle.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(fm.files, num)
	}
	return firstErr
}

// getFile returns the open file for the given number, or an error.
func (fm *FileManager) getFile(num int) (*BasicFile, error) {
	bf, ok := fm.files[num]
	if !ok {
		return nil, fmt.Errorf("file #%d not open", num)
	}
	return bf, nil
}

// --- Sequential I/O ---

// FileInput reads the next comma- or newline-delimited token from a sequential file.
// BASIC: INPUT #num, var — reads one value per call; the caller (generated code)
// converts the raw string to the target variable's type (float, int, string).
//
// The parser here handles quoted strings (with doubled-quote escaping) so that
// strings written with WRITE # (which quote-wraps strings) can be read back
// faithfully with INPUT #.
func (fm *FileManager) FileInput(num int) (string, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return "", err
	}
	if bf.Reader == nil {
		return "", fmt.Errorf("file #%d not open for input", num)
	}

	var sb strings.Builder
	quoted := false

	for {
		ch, err := bf.Reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				return sb.String(), nil
			}
			return "", err
		}

		if !quoted {
			if ch == ',' || ch == '\n' {
				break
			}
			if ch == '\r' {
				// peek for LF
				next, err := bf.Reader.ReadByte()
				if err == nil && next != '\n' {
					bf.Reader.UnreadByte()
				}
				break
			}
			if ch == '"' && sb.Len() == 0 {
				quoted = true
				continue
			}
		} else {
			if ch == '"' {
				// check for doubled quote
				next, err := bf.Reader.ReadByte()
				if err != nil || next != '"' {
					if err == nil {
						// consume trailing comma or newline
						if next == ',' || next == '\n' {
							// done
						} else if next == '\r' {
							peek, err := bf.Reader.ReadByte()
							if err == nil && peek != '\n' {
								bf.Reader.UnreadByte()
							}
						} else {
							bf.Reader.UnreadByte()
						}
					}
					break
				}
				// doubled quote -> literal quote
				sb.WriteByte('"')
				continue
			}
		}
		sb.WriteByte(ch)
	}

	return sb.String(), nil
}

// FileLineInput reads a full line from a sequential file.
func (fm *FileManager) FileLineInput(num int) (string, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return "", err
	}
	if bf.Reader == nil {
		return "", fmt.Errorf("file #%d not open for input", num)
	}

	line, err := bf.Reader.ReadString('\n')
	line = strings.TrimRight(line, "\r\n")
	if err != nil && err != io.EOF {
		return line, err
	}
	return line, nil
}

// FilePrint writes space-separated string values followed by CRLF to a sequential file.
// BASIC: PRINT #num, expr [; expr …] — mirrors console PRINT but to a file.
// Values are already formatted strings (the code generator pre-formats numbers).
func (fm *FileManager) FilePrint(num int, values ...string) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	w := fm.getWriter(bf)

	for i, v := range values {
		if i > 0 {
			w.WriteString(" ")
		}
		w.WriteString(v)
	}
	w.WriteString("\r\n")
	return w.Flush()
}

// FileWrite writes values in WRITE # format: strings are double-quoted and values
// are comma-separated, terminated with CRLF.
// BASIC: WRITE #num, expr [, expr …] — produces machine-readable output suitable
// for reading back with INPUT #. Strings are quoted; doubled quotes inside strings
// are escaped as "". Numbers are written without leading spaces.
func (fm *FileManager) FileWrite(num int, values ...interface{}) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	w := fm.getWriter(bf)

	for i, v := range values {
		if i > 0 {
			w.WriteString(",")
		}
		switch val := v.(type) {
		case string:
			w.WriteString("\"")
			w.WriteString(strings.ReplaceAll(val, "\"", "\"\""))
			w.WriteString("\"")
		case int:
			w.WriteString(strconv.Itoa(val))
		case int64:
			w.WriteString(strconv.FormatInt(val, 10))
		case float64:
			w.WriteString(strconv.FormatFloat(val, 'G', -1, 64))
		default:
			w.WriteString(fmt.Sprintf("%v", val))
		}
	}
	w.WriteString("\r\n")
	return w.Flush()
}

func (fm *FileManager) getWriter(bf *BasicFile) *bufio.Writer {
	if bf.Writer != nil {
		return bf.Writer
	}
	bf.Writer = bufio.NewWriter(bf.Handle)
	return bf.Writer
}

// --- Random Access I/O ---

// Field sets up field variable mappings in the record buffer.
func (fm *FileManager) Field(num int, fields []FieldDef) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	if bf.Mode != FileModeRandom {
		return fmt.Errorf("file #%d not open for random access", num)
	}

	// Validate total field length doesn't exceed record length
	total := 0
	for i := range fields {
		fields[i].Offset = total
		total += fields[i].Length
	}
	if total > bf.RecLen {
		return fmt.Errorf("FIELD total %d exceeds record length %d", total, bf.RecLen)
	}

	bf.Fields = fields
	return nil
}

// RandomGet reads a record from a random-access file into the field buffer.
// Record numbers are 1-based.
func (fm *FileManager) RandomGet(num int, rec int) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	if bf.Mode != FileModeRandom {
		return fmt.Errorf("file #%d not open for random access", num)
	}

	offset := int64(rec-1) * int64(bf.RecLen)
	_, err = bf.Handle.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// Clear the buffer
	for i := range bf.FieldBuf {
		bf.FieldBuf[i] = ' '
	}

	n, err := bf.Handle.Read(bf.FieldBuf)
	if err != nil && err != io.EOF {
		return err
	}
	// Pad remainder with spaces if short read
	for i := n; i < bf.RecLen; i++ {
		bf.FieldBuf[i] = ' '
	}
	return nil
}

// RandomPut writes the field buffer to a record in a random-access file.
// Record numbers are 1-based.
func (fm *FileManager) RandomPut(num int, rec int) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	if bf.Mode != FileModeRandom {
		return fmt.Errorf("file #%d not open for random access", num)
	}

	offset := int64(rec-1) * int64(bf.RecLen)
	_, err = bf.Handle.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = bf.Handle.Write(bf.FieldBuf)
	return err
}

// GetFieldValue reads a field variable value from the current record buffer.
func (fm *FileManager) GetFieldValue(num int, fieldName string) (string, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return "", err
	}

	for _, f := range bf.Fields {
		if strings.EqualFold(f.Name, fieldName) {
			end := f.Offset + f.Length
			if end > len(bf.FieldBuf) {
				end = len(bf.FieldBuf)
			}
			return string(bf.FieldBuf[f.Offset:end]), nil
		}
	}
	return "", fmt.Errorf("field %q not defined for file #%d", fieldName, num)
}

// Lset left-justifies a string into a field variable, padding with spaces.
// If num is 0, searches all open files for the named field.
func (fm *FileManager) Lset(num int, fieldName string, value string) error {
	if num == 0 {
		for n := range fm.files {
			if err := fm.lsetFile(n, fieldName, value); err == nil {
				return nil
			}
		}
		return fmt.Errorf("field %q not defined in any open file", fieldName)
	}
	return fm.lsetFile(num, fieldName, value)
}

func (fm *FileManager) lsetFile(num int, fieldName string, value string) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}

	for _, f := range bf.Fields {
		if strings.EqualFold(f.Name, fieldName) {
			// Clear field
			for i := f.Offset; i < f.Offset+f.Length && i < len(bf.FieldBuf); i++ {
				bf.FieldBuf[i] = ' '
			}
			// Copy value left-justified
			n := f.Length
			if len(value) < n {
				n = len(value)
			}
			copy(bf.FieldBuf[f.Offset:f.Offset+n], value[:n])
			return nil
		}
	}
	return fmt.Errorf("field %q not defined for file #%d", fieldName, num)
}

// Rset right-justifies a string into a field variable, padding with spaces.
// If num is 0, searches all open files for the named field.
func (fm *FileManager) Rset(num int, fieldName string, value string) error {
	if num == 0 {
		for n := range fm.files {
			if err := fm.rsetFile(n, fieldName, value); err == nil {
				return nil
			}
		}
		return fmt.Errorf("field %q not defined in any open file", fieldName)
	}
	return fm.rsetFile(num, fieldName, value)
}

func (fm *FileManager) rsetFile(num int, fieldName string, value string) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}

	for _, f := range bf.Fields {
		if strings.EqualFold(f.Name, fieldName) {
			// Clear field
			for i := f.Offset; i < f.Offset+f.Length && i < len(bf.FieldBuf); i++ {
				bf.FieldBuf[i] = ' '
			}
			// Copy value right-justified
			start := f.Length - len(value)
			if start < 0 {
				// Value longer than field, truncate from end
				copy(bf.FieldBuf[f.Offset:f.Offset+f.Length], value[len(value)-f.Length:])
			} else {
				copy(bf.FieldBuf[f.Offset+start:f.Offset+f.Length], value)
			}
			return nil
		}
	}
	return fmt.Errorf("field %q not defined for file #%d", fieldName, num)
}

// --- Binary I/O ---

// BinaryGet reads raw bytes from a binary file at the given position (1-based).
func (fm *FileManager) BinaryGet(num int, pos int64, size int) ([]byte, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return nil, err
	}
	if bf.Mode != FileModeBinary {
		return nil, fmt.Errorf("file #%d not open for binary access", num)
	}

	_, err = bf.Handle.Seek(pos-1, io.SeekStart)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, size)
	n, err := io.ReadFull(bf.Handle, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, err
	}
	return buf[:n], nil
}

// BinaryPut writes raw bytes to a binary file at the given position (1-based).
func (fm *FileManager) BinaryPut(num int, pos int64, data []byte) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}
	if bf.Mode != FileModeBinary {
		return fmt.Errorf("file #%d not open for binary access", num)
	}

	_, err = bf.Handle.Seek(pos-1, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = bf.Handle.Write(data)
	return err
}

// --- File Status Functions ---

// Eof returns true if end-of-file has been reached for the given file number.
func (fm *FileManager) Eof(num int) (bool, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return false, err
	}

	if bf.Reader != nil {
		_, err := bf.Reader.Peek(1)
		return err == io.EOF, nil
	}

	// For random/binary, check current position vs file size
	cur, err := bf.Handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	end, err := bf.Handle.Seek(0, io.SeekEnd)
	if err != nil {
		return false, err
	}
	// Seek back to where we were
	bf.Handle.Seek(cur, io.SeekStart)
	return cur >= end, nil
}

// Loc returns the current record/byte position for the given file number.
// For random files, returns the last record number read/written.
// For sequential/binary files, returns the byte position.
func (fm *FileManager) Loc(num int) (int64, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return 0, err
	}

	pos, err := bf.Handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	if bf.Mode == FileModeRandom {
		return pos / int64(bf.RecLen), nil
	}
	return pos, nil
}

// Lof returns the length of the file in bytes.
func (fm *FileManager) Lof(num int) (int64, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return 0, err
	}

	info, err := bf.Handle.Stat()
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FileSeek sets the file position for binary/random files.
// Position is 1-based for BASIC compatibility.
func (fm *FileManager) FileSeek(num int, pos int64) error {
	bf, err := fm.getFile(num)
	if err != nil {
		return err
	}

	if bf.Mode == FileModeRandom {
		_, err = bf.Handle.Seek((pos-1)*int64(bf.RecLen), io.SeekStart)
	} else {
		_, err = bf.Handle.Seek(pos-1, io.SeekStart)
	}
	return err
}

// FileSeekPos returns the current seek position (1-based).
func (fm *FileManager) FileSeekPos(num int) (int64, error) {
	bf, err := fm.getFile(num)
	if err != nil {
		return 0, err
	}

	pos, err := bf.Handle.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	if bf.Mode == FileModeRandom {
		return pos/int64(bf.RecLen) + 1, nil
	}
	return pos + 1, nil
}

// --- Helper functions for binary encoding (used with MKI$/MKL$/MKS$/MKD$ and CVI/CVL/CVS/CVD) ---

// MkiBytes encodes a 16-bit integer to 2 bytes (little-endian).
func MkiBytes(v int16) []byte {
	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, uint16(v))
	return buf
}

// MklBytes encodes a 32-bit integer to 4 bytes (little-endian).
func MklBytes(v int32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(v))
	return buf
}

// MksBytes encodes a 32-bit float to 4 bytes (little-endian).
func MksBytes(v float32) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, math.Float32bits(v))
	return buf
}

// MkdBytes encodes a 64-bit float to 8 bytes (little-endian).
func MkdBytes(v float64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
	return buf
}

// CviBytes decodes 2 bytes to a 16-bit integer (little-endian).
func CviBytes(b []byte) int16 {
	if len(b) < 2 {
		return 0
	}
	return int16(binary.LittleEndian.Uint16(b))
}

// CvlBytes decodes 4 bytes to a 32-bit integer (little-endian).
func CvlBytes(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(binary.LittleEndian.Uint32(b))
}

// CvsBytes decodes 4 bytes to a 32-bit float (little-endian).
func CvsBytes(b []byte) float32 {
	if len(b) < 4 {
		return 0
	}
	return math.Float32frombits(binary.LittleEndian.Uint32(b))
}

// CvdBytes decodes 8 bytes to a 64-bit float (little-endian).
func CvdBytes(b []byte) float64 {
	if len(b) < 8 {
		return 0
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(b))
}
