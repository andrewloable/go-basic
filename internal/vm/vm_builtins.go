package vm

import (
	"github.com/loabletech/go-basic/internal/runtime"
)

// ---------------------------------------------------------------------------
// Built-in function dispatch
// ---------------------------------------------------------------------------

func (vm *VM) execBuiltin(id BuiltinID) error {
	switch id {

	// --- Math (1-argument, numeric) ---
	case BuiltinAbs:
		v := vm.pop()
		vm.push(FloatVal(runtime.Abs(v.asFloat())))

	case BuiltinSgn:
		v := vm.pop()
		vm.push(IntVal(int64(runtime.Sgn(v.asFloat()))))

	case BuiltinInt:
		v := vm.pop()
		vm.push(FloatVal(runtime.IntFloor(v.asFloat())))

	case BuiltinFix:
		v := vm.pop()
		vm.push(FloatVal(runtime.Fix(v.asFloat())))

	case BuiltinCeil:
		v := vm.pop()
		vm.push(FloatVal(runtime.Ceil(v.asFloat())))

	case BuiltinSqr:
		v := vm.pop()
		result, err := runtime.Sqr(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinExp:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp(v.asFloat())))

	case BuiltinExp2:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp2(v.asFloat())))

	case BuiltinExp10:
		v := vm.pop()
		vm.push(FloatVal(runtime.Exp10(v.asFloat())))

	case BuiltinLog:
		v := vm.pop()
		result, err := runtime.Log(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinLog2:
		v := vm.pop()
		result, err := runtime.Log2(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinLog10:
		v := vm.pop()
		result, err := runtime.Log10(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	case BuiltinSin:
		v := vm.pop()
		vm.push(FloatVal(runtime.Sin(v.asFloat())))

	case BuiltinCos:
		v := vm.pop()
		vm.push(FloatVal(runtime.Cos(v.asFloat())))

	case BuiltinTan:
		v := vm.pop()
		vm.push(FloatVal(runtime.Tan(v.asFloat())))

	case BuiltinAtn:
		v := vm.pop()
		vm.push(FloatVal(runtime.Atn(v.asFloat())))

	case BuiltinCint:
		v := vm.pop()
		result, err := runtime.Cint(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinClng:
		v := vm.pop()
		result, err := runtime.Clng(v.asFloat())
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCsng:
		v := vm.pop()
		vm.push(FloatVal(float64(runtime.Csng(v.asFloat()))))

	case BuiltinCdbl:
		v := vm.pop()
		vm.push(FloatVal(runtime.Cdbl(v.asFloat())))

	case BuiltinRnd:
		v := vm.pop()
		vm.push(FloatVal(vm.rng.Rnd(v.asFloat())))

	// --- String functions ---
	case BuiltinLeft:
		n := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Left(s.Str, int(n.asInt()))))

	case BuiltinRight:
		n := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Right(s.Str, int(n.asInt()))))

	case BuiltinMid:
		length := vm.pop()
		start := vm.pop()
		s := vm.pop()
		vm.push(StringVal(runtime.Mid(s.Str, int(start.asInt()), int(length.asInt()))))

	case BuiltinLen:
		s := vm.pop()
		vm.push(IntVal(int64(runtime.Len(s.Str))))

	case BuiltinAsc:
		s := vm.pop()
		result, err := runtime.Asc(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinChr:
		n := vm.pop()
		result, err := runtime.Chr(int(n.asInt()))
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(StringVal(result))

	case BuiltinStr:
		n := vm.pop()
		vm.push(StringVal(runtime.Str(n.asFloat())))

	case BuiltinVal:
		s := vm.pop()
		vm.push(FloatVal(runtime.Val(s.Str)))

	case BuiltinInstr:
		find := vm.pop()
		s := vm.pop()
		start := vm.pop()
		vm.push(IntVal(int64(runtime.Instr(int(start.asInt()), s.Str, find.Str))))

	case BuiltinUCase:
		s := vm.pop()
		vm.push(StringVal(runtime.UCase(s.Str)))

	case BuiltinLCase:
		s := vm.pop()
		vm.push(StringVal(runtime.LCase(s.Str)))

	case BuiltinLTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.LTrim(s.Str)))

	case BuiltinRTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.RTrim(s.Str)))

	case BuiltinTrim:
		s := vm.pop()
		vm.push(StringVal(runtime.Trim(s.Str)))

	case BuiltinSpace:
		n := vm.pop()
		vm.push(StringVal(runtime.Space(int(n.asInt()))))

	case BuiltinString:
		char := vm.pop()
		n := vm.pop()
		vm.push(StringVal(runtime.StringRepeat(int(n.asInt()), byte(char.asInt()))))

	case BuiltinHex:
		n := vm.pop()
		vm.push(StringVal(runtime.Hex(int(n.asInt()))))

	case BuiltinOct:
		n := vm.pop()
		vm.push(StringVal(runtime.Oct(int(n.asInt()))))

	case BuiltinBin:
		n := vm.pop()
		vm.push(StringVal(runtime.Bin(int(n.asInt()))))

	// --- Binary conversion functions ---
	case BuiltinMki:
		n := vm.pop()
		vm.push(StringVal(runtime.Mki(int16(n.asInt()))))

	case BuiltinMkl:
		n := vm.pop()
		vm.push(StringVal(runtime.Mkl(int32(n.asInt()))))

	case BuiltinMks:
		n := vm.pop()
		vm.push(StringVal(runtime.Mks(float32(n.asFloat()))))

	case BuiltinMkd:
		n := vm.pop()
		vm.push(StringVal(runtime.Mkd(n.asFloat())))

	case BuiltinCvi:
		s := vm.pop()
		result, err := runtime.Cvi(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCvl:
		s := vm.pop()
		result, err := runtime.Cvl(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(IntVal(int64(result)))

	case BuiltinCvs:
		s := vm.pop()
		result, err := runtime.Cvs(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(float64(result)))

	case BuiltinCvd:
		s := vm.pop()
		result, err := runtime.Cvd(s.Str)
		if err != nil {
			return vm.runtimeError("%s", err)
		}
		vm.push(FloatVal(result))

	// --- I/O helpers ---
	case BuiltinTab:
		n := vm.pop()
		vm.push(StringVal(runtime.Spc(int(n.asInt()))))

	case BuiltinSpc:
		n := vm.pop()
		vm.push(StringVal(runtime.Spc(int(n.asInt()))))

	default:
		return vm.runtimeError("unknown built-in function ID %d", id)
	}

	return nil
}
