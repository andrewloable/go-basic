package vm

import (
	"testing"
)

// ===========================================================================
// Value method tests (String, asFloat, asInt, isTruthy, asString)
// ===========================================================================

func TestValueString(t *testing.T) {
	cases := []struct {
		v    Value
		want string
	}{
		{IntVal(42), "Int(42)"},
		{FloatVal(3.14), "Float(3.14)"},
		{StringVal("hi"), `String("hi")`},
		{Value{Type: 99}, "Value(?)"},
	}
	for _, tc := range cases {
		got := tc.v.String()
		if got != tc.want {
			t.Errorf("Value.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestValueAsFloat(t *testing.T) {
	if IntVal(5).asFloat() != 5.0 {
		t.Error("IntVal.asFloat() should be 5.0")
	}
	if FloatVal(2.5).asFloat() != 2.5 {
		t.Error("FloatVal.asFloat() should be 2.5")
	}
	if StringVal("x").asFloat() != 0 {
		t.Error("StringVal.asFloat() should be 0")
	}
}

func TestValueAsInt(t *testing.T) {
	if IntVal(7).asInt() != 7 {
		t.Error("IntVal.asInt() should be 7")
	}
	if FloatVal(3.9).asInt() != 3 {
		t.Error("FloatVal(3.9).asInt() should be 3 (truncated)")
	}
	if StringVal("x").asInt() != 0 {
		t.Error("StringVal.asInt() should be 0")
	}
}

func TestValueIsTruthy(t *testing.T) {
	if !IntVal(-1).isTruthy() {
		t.Error("IntVal(-1).isTruthy() should be true")
	}
	if IntVal(0).isTruthy() {
		t.Error("IntVal(0).isTruthy() should be false")
	}
	if !FloatVal(0.1).isTruthy() {
		t.Error("FloatVal(0.1).isTruthy() should be true")
	}
	if FloatVal(0).isTruthy() {
		t.Error("FloatVal(0).isTruthy() should be false")
	}
	if !StringVal("x").isTruthy() {
		t.Error("StringVal(non-empty).isTruthy() should be true")
	}
	if StringVal("").isTruthy() {
		t.Error("StringVal('').isTruthy() should be false")
	}
	if (Value{Type: 99}).isTruthy() {
		t.Error("unknown type isTruthy() should be false")
	}
}

func TestValueAsString(t *testing.T) {
	if IntVal(5).asString() != " 5 " {
		t.Errorf("IntVal(5).asString() = %q, want ' 5 '", IntVal(5).asString())
	}
	if IntVal(-3).asString() != "-3 " {
		t.Errorf("IntVal(-3).asString() = %q, want '-3 '", IntVal(-3).asString())
	}
	if FloatVal(1.5).asString() != " 1.5 " {
		t.Errorf("FloatVal(1.5).asString() = %q, want ' 1.5 '", FloatVal(1.5).asString())
	}
	if FloatVal(-1.5).asString() != "-1.5 " {
		t.Errorf("FloatVal(-1.5).asString() = %q, want '-1.5 '", FloatVal(-1.5).asString())
	}
	if StringVal("hello").asString() != "hello" {
		t.Error("StringVal.asString() should return the string as-is")
	}
	if (Value{Type: 99}).asString() != "" {
		t.Error("unknown type asString() should be ''")
	}
}

func TestValueHelpers(t *testing.T) {
	iv := IntVal(10)
	if iv.asFloat() != 10.0 {
		t.Fatal("IntVal.asFloat failed")
	}
	if iv.asInt() != 10 {
		t.Fatal("IntVal.asInt failed")
	}
	if !iv.isTruthy() {
		t.Fatal("IntVal(10) should be truthy")
	}

	fv := FloatVal(0.0)
	if fv.isTruthy() {
		t.Fatal("FloatVal(0.0) should not be truthy")
	}

	sv := StringVal("")
	if sv.isTruthy() {
		t.Fatal("StringVal(\"\") should not be truthy")
	}
	sv2 := StringVal("x")
	if !sv2.isTruthy() {
		t.Fatal("StringVal(\"x\") should be truthy")
	}
}
