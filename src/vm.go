package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

// Mem size
const M = 32768

// Op codes
const (
	HALT uint16 = iota
	SET
	PUSH
	POP
	EQ
	GT
	JMP
	JT
	JF
	ADD
	MULT
	MOD
	AND
	OR
	NOT
	RMEM
	WMEM
	CALL
	RET
	OUT
	IN
	NOOP
)

// vm type
type vm struct {
	register [8]uint16
	stack    []uint16
	memory   []uint16
}

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please give the input binary as parameter: %v challenge.bin", os.Args[0])
		os.Exit(2)
	}

	// Read file
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	// Initialize VM
	vm := &vm{memory: parse(string(b))}

	// Execute
	vm.exec()
}

// exec executes the binary code
func (vm *vm) exec() {
	// Our cursor that points to the actual position in the memory
	cursor := uint16(0)

	// Reader for opcode IN
	reader := bufio.NewReader(os.Stdin)

	for {
		// Retrieve the operation
		op := vm.memory[cursor]
		switch op {
		case HALT: // Code 0
			fmt.Print("Halt op code !")
			os.Exit(0)

		case SET: // Code 1
			vm.set(cursor+1, vm.get(cursor+2))
			cursor += 3

		case PUSH: // Code 2
			vm.stack = append(vm.stack, vm.get(cursor+1))
			cursor += 2

		case POP: // Code 3
			popped, err := vm.pop()
			if err != nil {
				panic(err)
			}
			vm.set(cursor+1, popped)
			cursor += 2

		case EQ: // Code 4
			if vm.get(cursor+2) == vm.get(cursor+3) {
				vm.set(cursor+1, 1)
			} else {
				vm.set(cursor+1, 0)
			}
			cursor += 4

		case GT: // Code 5
			if vm.get(cursor+2) > vm.get(cursor+3) {
				vm.set(cursor+1, 1)
			} else {
				vm.set(cursor+1, 0)
			}
			cursor += 4

		case JMP: // Code 6
			cursor = vm.get(cursor + 1)

		case JT: // Code 7
			if vm.get(cursor+1) != 0 {
				cursor = vm.get(cursor + 2)
			} else {
				cursor += 3
			}

		case JF: // Code 8
			if vm.get(cursor+1) == 0 {
				cursor = vm.get(cursor + 2)
			} else {
				cursor += 3
			}

		case ADD: // Code 9
			vm.set(cursor+1, (vm.get(cursor+2)+vm.get(cursor+3))%M)
			cursor += 4

		case MULT: // Code 10
			vm.set(cursor+1, (vm.get(cursor+2)*vm.get(cursor+3))%M)
			cursor += 4

		case MOD: // Code 11
			vm.set(cursor+1, vm.get(cursor+2)%vm.get(cursor+3))
			cursor += 4

		case AND: // Code 12
			vm.set(cursor+1, vm.get(cursor+2)&vm.get(cursor+3))
			cursor += 4

		case OR: // Code 13
			vm.set(cursor+1, vm.get(cursor+2)|vm.get(cursor+3))
			cursor += 4

		case NOT: // Code 14
			vm.set(cursor+1, 0x7fff&^vm.get(cursor+2))
			cursor += 3

		case RMEM: // Code 15
			vm.set(cursor+1, vm.get(vm.get(cursor+2)))
			cursor += 3

		case WMEM: // Code 16
			vm.memory[vm.get(cursor+1)] = vm.get(cursor + 2)
			cursor += 3

		case CALL: // Code 17
			vm.push(cursor + 2)
			cursor = vm.get(cursor + 1)

		case RET: // Code 18
			popped, err := vm.pop()
			if err != nil {
				// Halt
				fmt.Print("RET operation resulted in halt !")
				os.Exit(0)
			}
			cursor = popped

		case OUT: // Code 19
			fmt.Print(string(vm.get(cursor + 1)))
			cursor += 2

		case IN: // Code 20
			b, _ := reader.ReadByte()
			vm.set(cursor+1, uint16(b))
			cursor += 2

		case NOOP: // Code 21
			cursor++

		default:
			panic(fmt.Errorf("Unrecognized opcode %v", op))
		}
	}
}

// parse Parses the binary as a string and return the list of 16-bits values respecting little-endian convention
func parse(input string) []uint16 {
	mem := []uint16{}

	for i := 0; i < len(input)-1; i += 2 {
		v, err := strconv.ParseUint(tob(input[i+1])+tob(input[i]), 2, 16)
		if err != nil {
			panic(err)
		}

		mem = append(mem, uint16(v))
	}
	return mem
}

// tob Converts to byte representation of size 8
func tob(c uint8) string {
	res := fmt.Sprintf("%b", c)
	s := len(res)
	for i := 0; i < 8-s; i++ {
		res = "0" + res
	}
	return res
}

// get Retrieves a value by checking the register
func (vm vm) get(addr uint16) uint16 {
	m := vm.memory[addr]
	if m > M+7 {
		panic(fmt.Errorf("Get operation: Invalid address %v", m))
	}

	// Register
	if m >= M {
		return vm.register[m-M]
	}

	return m
}

// set Modify a value in the memory
func (vm *vm) set(addr, value uint16) {
	m := vm.memory[addr]
	if m > M+8 {
		panic(fmt.Errorf("Set operation: Invalid address %v", m))
	}

	// Set in register
	vm.register[m-M] = value
}

// Push to stack
func (vm *vm) push(value uint16) {
	vm.stack = append(vm.stack, value)
}

// Pop from stack
func (vm *vm) pop() (uint16, error) {
	if len(vm.stack) > 0 {
		res := vm.stack[len(vm.stack)-1]
		vm.stack = vm.stack[:len(vm.stack)-1]
		return res, nil
	}
	return 0, fmt.Errorf("empty stack ")
}