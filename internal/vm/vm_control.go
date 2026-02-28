package vm

// ---------------------------------------------------------------------------
// Variable / array opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execLoad(inst Instruction) error {
	name := vm.chunk.Constants[inst.Operand].Str
	v, ok := vm.globals[name]
	if !ok {
		// Uninitialized variable — return zero value.
		v = Value{Type: ValInt, Int: 0}
	}
	vm.push(v)
	return nil
}

func (vm *VM) execStore(inst Instruction) {
	name := vm.chunk.Constants[inst.Operand].Str
	vm.globals[name] = vm.pop()
}

func (vm *VM) execLoadArray(inst Instruction) error {
	name := vm.chunk.Constants[inst.Operand].Str
	arr, ok := vm.arrays[name]
	if !ok {
		return vm.runtimeError("array %q not dimensioned", name)
	}
	idx, err := vm.computeArrayIndex(arr)
	if err != nil {
		return err
	}
	vm.push(arr.data[idx])
	return nil
}

func (vm *VM) execStoreArray(inst Instruction) error {
	name := vm.chunk.Constants[inst.Operand].Str
	arr, ok := vm.arrays[name]
	if !ok {
		return vm.runtimeError("array %q not dimensioned", name)
	}
	val := vm.pop()
	idx, err := vm.computeArrayIndex(arr)
	if err != nil {
		return err
	}
	arr.data[idx] = val
	return nil
}

// ---------------------------------------------------------------------------
// Control flow opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execCall(inst Instruction) {
	vm.callStack = append(vm.callStack, CallFrame{
		ReturnAddr: vm.ip,
		LocalBase:  0,
	})
	vm.ip = int(inst.Operand)
}

func (vm *VM) execRet() error {
	if len(vm.callStack) == 0 {
		return vm.runtimeError("RETURN without CALL")
	}
	frame := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]
	vm.ip = frame.ReturnAddr
	return nil
}

func (vm *VM) execGosub(inst Instruction) {
	vm.callStack = append(vm.callStack, CallFrame{
		ReturnAddr: vm.ip,
		LocalBase:  0,
	})
	vm.ip = int(inst.Operand)
}

func (vm *VM) execReturn() error {
	if len(vm.callStack) == 0 {
		return vm.runtimeError("RETURN without GOSUB")
	}
	frame := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]
	vm.ip = frame.ReturnAddr
	return nil
}

// ---------------------------------------------------------------------------
// Type conversion opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execToInt() {
	v := vm.pop()
	vm.push(IntVal(v.asInt()))
}

func (vm *VM) execToLong() {
	v := vm.pop()
	vm.push(IntVal(v.asInt()))
}

func (vm *VM) execToSingle() {
	v := vm.pop()
	vm.push(FloatVal(float64(float32(v.asFloat()))))
}

func (vm *VM) execToDouble() {
	v := vm.pop()
	vm.push(FloatVal(v.asFloat()))
}

func (vm *VM) execToString() {
	v := vm.pop()
	vm.push(StringVal(v.asString()))
}

// ---------------------------------------------------------------------------
// Array opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execDimArray(inst Instruction) {
	numDims := int(inst.Operand)
	nameVal := vm.pop()
	name := nameVal.Str

	dims := make([]int, numDims)
	totalSize := 1
	for i := numDims - 1; i >= 0; i-- {
		sizeVal := vm.pop()
		size := int(sizeVal.asInt()) + 1 // BASIC arrays are 0..N, so N+1 elements
		dims[i] = size
		totalSize *= size
	}

	vm.arrays[name] = &arrayValue{
		dims: dims,
		data: make([]Value, totalSize),
	}
}

// ---------------------------------------------------------------------------
// Data opcode handlers
// ---------------------------------------------------------------------------

func (vm *VM) execRead(inst Instruction) error {
	if vm.dataPtr >= len(vm.dataPool) {
		return vm.runtimeError("out of DATA")
	}
	val := vm.dataPool[vm.dataPtr]
	vm.dataPtr++
	name := vm.chunk.Constants[inst.Operand].Str
	vm.globals[name] = val
	return nil
}

func (vm *VM) execRestore(inst Instruction) {
	target := int(inst.Operand)
	if target < 0 {
		vm.dataPtr = 0
	} else {
		vm.dataPtr = target
	}
}

// ---------------------------------------------------------------------------
// Array index computation
// ---------------------------------------------------------------------------

func (vm *VM) computeArrayIndex(arr *arrayValue) (int, error) {
	numDims := len(arr.dims)
	indices := make([]int, numDims)
	for i := numDims - 1; i >= 0; i-- {
		indices[i] = int(vm.pop().asInt())
	}

	// Compute flat index (row-major order).
	flat := 0
	multiplier := 1
	for i := numDims - 1; i >= 0; i-- {
		idx := indices[i]
		if idx < 0 || idx >= arr.dims[i] {
			return 0, vm.runtimeError("array index out of bounds: dimension %d, index %d, size %d",
				i, idx, arr.dims[i])
		}
		flat += idx * multiplier
		multiplier *= arr.dims[i]
	}
	return flat, nil
}
