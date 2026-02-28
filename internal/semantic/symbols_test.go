package semantic

import (
	"strings"
	"testing"
)

// ===========================================================================
// SymbolTable unit tests (symbols.go)
// ===========================================================================

func TestNewSymbolTable(t *testing.T) {
	st := NewSymbolTable()
	if st.GlobalScope == nil {
		t.Fatal("GlobalScope should not be nil")
	}
	if st.CurrentScope != st.GlobalScope {
		t.Fatal("CurrentScope should start as GlobalScope")
	}
	if st.OptionBase != 0 {
		t.Fatalf("expected OptionBase 0, got %d", st.OptionBase)
	}
}

func TestEnterExitScope(t *testing.T) {
	st := NewSymbolTable()

	st.EnterScope("mySub")
	if st.CurrentScope.Name != "MYSUB" {
		t.Fatalf("expected scope name MYSUB, got %s", st.CurrentScope.Name)
	}
	if st.CurrentScope.Parent != st.GlobalScope {
		t.Fatal("parent of mySub scope should be GlobalScope")
	}
	if _, ok := st.Scopes["MYSUB"]; !ok {
		t.Fatal("scope MYSUB should be registered in Scopes map")
	}

	st.ExitScope()
	if st.CurrentScope != st.GlobalScope {
		t.Fatal("after ExitScope, CurrentScope should be GlobalScope")
	}
}

func TestExitScopeAtGlobal(t *testing.T) {
	st := NewSymbolTable()
	// Exiting from global scope should be a no-op.
	st.ExitScope()
	if st.CurrentScope != st.GlobalScope {
		t.Fatal("ExitScope at global should stay at global")
	}
}

func TestDefineAndLookup(t *testing.T) {
	st := NewSymbolTable()
	sym := &Symbol{Name: "X%", Type: SymVariable, DataType: TypeInteger}
	if err := st.Define("X%", sym); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := st.Lookup("x%") // case-insensitive
	if got != sym {
		t.Fatal("Lookup should find the symbol regardless of case")
	}
}

func TestDefineDuplicate(t *testing.T) {
	st := NewSymbolTable()
	sym := &Symbol{Name: "A", Type: SymVariable}
	_ = st.Define("A", sym)
	err := st.Define("a", &Symbol{Name: "A", Type: SymVariable})
	if err == nil {
		t.Fatal("expected error on duplicate define")
	}
	if !strings.Contains(err.Error(), "already defined") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestLookupParentScope(t *testing.T) {
	st := NewSymbolTable()
	globalSym := &Symbol{Name: "G", Type: SymVariable, DataType: TypeSingle}
	_ = st.Define("G", globalSym)

	st.EnterScope("inner")
	localSym := &Symbol{Name: "L", Type: SymVariable, DataType: TypeInteger}
	_ = st.Define("L", localSym)

	// Should find local symbol.
	if st.Lookup("L") != localSym {
		t.Fatal("should find local symbol L")
	}
	// Should find global symbol through parent chain.
	if st.Lookup("G") != globalSym {
		t.Fatal("should find global symbol G via parent scope")
	}
	// LookupLocal should NOT find the global symbol.
	if st.LookupLocal("G") != nil {
		t.Fatal("LookupLocal should not find G in inner scope")
	}
}

func TestLookupNotFound(t *testing.T) {
	st := NewSymbolTable()
	if st.Lookup("NONEXISTENT") != nil {
		t.Fatal("expected nil for missing symbol")
	}
}

// ===========================================================================
// ResolveType tests
// ===========================================================================

func TestResolveTypeSuffixes(t *testing.T) {
	st := NewSymbolTable()
	tests := []struct {
		name     string
		expected DataType
	}{
		{"count%", TypeInteger},
		{"total&", TypeLong},
		{"ratio!", TypeSingle},
		{"pi#", TypeDouble},
		{"name$", TypeString},
	}
	for _, tc := range tests {
		got := st.ResolveType(tc.name)
		if got != tc.expected {
			t.Errorf("ResolveType(%q) = %v, want %v", tc.name, got, tc.expected)
		}
	}
}

func TestResolveTypeDefault(t *testing.T) {
	st := NewSymbolTable()
	// No suffix, no DEFtype → default is Single.
	got := st.ResolveType("myVar")
	if got != TypeSingle {
		t.Errorf("expected TypeSingle, got %v", got)
	}
}

func TestResolveTypeWithDefType(t *testing.T) {
	st := NewSymbolTable()
	st.SetDefType('I', 'N', TypeInteger) // DEFINT I-N

	got := st.ResolveType("index")
	if got != TypeInteger {
		t.Errorf("expected TypeInteger for 'index', got %v", got)
	}
	got = st.ResolveType("name")
	if got != TypeInteger {
		t.Errorf("expected TypeInteger for 'name' (N is in range), got %v", got)
	}
	got = st.ResolveType("Xval")
	if got != TypeSingle {
		t.Errorf("expected TypeSingle for 'Xval' (not in DEFINT range), got %v", got)
	}
}

func TestSetDefTypeSingleLetter(t *testing.T) {
	st := NewSymbolTable()
	st.SetDefType('A', 'A', TypeDouble)
	if st.ResolveType("alpha") != TypeDouble {
		t.Error("expected TypeDouble for variable starting with A")
	}
	if st.ResolveType("beta") != TypeSingle {
		t.Error("expected default TypeSingle for variable starting with B")
	}
}

func TestResolveTypeSuffixOverridesDefType(t *testing.T) {
	st := NewSymbolTable()
	st.SetDefType('A', 'Z', TypeInteger) // DEFINT A-Z
	// An explicit suffix should override the DEFtype.
	if st.ResolveType("total#") != TypeDouble {
		t.Error("explicit suffix # should override DEFINT")
	}
}

// ===========================================================================
// normaliseUpper
// ===========================================================================

func TestNormaliseUpperLowercase(t *testing.T) {
	// normaliseUpper should convert lowercase to uppercase.
	got := normaliseUpper('a')
	if got != 'A' {
		t.Errorf("normaliseUpper('a') = %c, want 'A'", got)
	}
	got2 := normaliseUpper('z')
	if got2 != 'Z' {
		t.Errorf("normaliseUpper('z') = %c, want 'Z'", got2)
	}
	// Uppercase should be returned unchanged.
	got3 := normaliseUpper('A')
	if got3 != 'A' {
		t.Errorf("normaliseUpper('A') = %c, want 'A'", got3)
	}
}

// SetDefType uses normaliseUpper — test with lowercase range.
func TestSetDefTypeLowercaseRange(t *testing.T) {
	st := NewSymbolTable()
	// Lowercase 'a'-'z' range should be treated as uppercase.
	st.SetDefType('a', 'z', TypeString)
	dt := st.ResolveType("name") // 'n' should be in range A-Z
	if dt != TypeString {
		t.Errorf("SetDefType lowercase: ResolveType('name') = %v, want TypeString", dt)
	}
}

// ===========================================================================
// SymbolType.String() and DataType.String() (defined in symbols.go)
// ===========================================================================

func TestSymbolTypeString(t *testing.T) {
	tests := []struct {
		st   SymbolType
		want string
	}{
		{SymVariable, "Variable"},
		{SymArray, "Array"},
		{SymFunction, "Function"},
		{SymSub, "Sub"},
		{SymLabel, "Label"},
		{SymConst, "Const"},
		{SymDefFn, "DefFn"},
	}
	for _, tc := range tests {
		if got := tc.st.String(); got != tc.want {
			t.Errorf("SymbolType(%d).String() = %q, want %q", tc.st, got, tc.want)
		}
	}
}

func TestDataTypeString(t *testing.T) {
	tests := []struct {
		dt   DataType
		want string
	}{
		{TypeInteger, "Integer(%)"},
		{TypeLong, "Long(&)"},
		{TypeSingle, "Single(!)"},
		{TypeDouble, "Double(#)"},
		{TypeString, "String($)"},
		{TypeUnknown, "Unknown"},
	}
	for _, tc := range tests {
		if got := tc.dt.String(); got != tc.want {
			t.Errorf("DataType(%d).String() = %q, want %q", tc.dt, got, tc.want)
		}
	}
}
