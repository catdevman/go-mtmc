package emulator

import (
	"encoding/binary"
	"log"
	"sync"
	"time"
)

const (
	WordSize   = 2
	MemorySize = 1 << 4 // 4096 bytes (4K)
)

// MonTanaMiniComputer represents the state of the virtual computer.
type MonTanaMiniComputer struct {
	Memory    []byte
	Registers [16]uint16
	Running   bool
	mutex     sync.Mutex
	observers []Observer
}

// Observer is an interface for components that need to be notified of computer state changes.
type Observer interface {
	Update(computer *MonTanaMiniComputer)
}

// New creates a new MTMC instance.
func New() *MonTanaMiniComputer {
	m := &MonTanaMiniComputer{
		Memory: make([]byte, MemorySize),
	}
	// Initialize SP to the top of memory
	m.Registers[SP] = MemorySize - 2
	return m
}

// AddObserver adds an observer to the computer.
func (c *MonTanaMiniComputer) AddObserver(o Observer) {
	c.observers = append(c.observers, o)
}

// notifyObservers notifies all observers of a state change.
func (c *MonTanaMiniComputer) notifyObservers() {
	for _, o := range c.observers {
		o.Update(c)
	}
}

// Run starts the computer's clock and execution cycle.
func (c *MonTanaMiniComputer) Run() {
	ticker := time.NewTicker(time.Second / 1000) // 1kHz clock speed
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		if c.Running {
			c.step()
			c.notifyObservers()
		}
		c.mutex.Unlock()
	}
}

// Step executes a single instruction.
func (c *MonTanaMiniComputer) Step() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.step()
	c.notifyObservers()
}

// step executes a single instruction.
func (c *MonTanaMiniComputer) step() {
	pc := c.Registers[PC]
	if pc >= MemorySize-1 {
		log.Println("PC out of bounds, stopping execution.")
		c.Running = false
		return
	}

	instruction := binary.BigEndian.Uint16(c.Memory[pc:])
	c.Registers[PC] += 2

	// Decode and execute the instruction based on the specification
	opCode := (instruction & 0b1111000000000000) >> 12
	regD := (instruction & 0b0000111100000000) >> 8
	regS := (instruction & 0b0000000011110000) >> 4
	regT := instruction & 0b0000000000001111
	imm := int16(instruction & 0b0000000011111111)

	switch opCode {
	// ALU Instructions
	case 0b0001: // ADD
		c.Registers[regD] = c.Registers[regS] + c.Registers[regT]
	case 0b0010: // SUB
		c.Registers[regD] = c.Registers[regS] - c.Registers[regT]
	case 0b0011: // AND
		c.Registers[regD] = c.Registers[regS] & c.Registers[regT]
	case 0b0100: // OR
		c.Registers[regD] = c.Registers[regS] | c.Registers[regT]
	case 0b0101: // XOR
		c.Registers[regD] = c.Registers[regS] ^ c.Registers[regT]
	case 0b0110: // SLL
		c.Registers[regD] = c.Registers[regS] << c.Registers[regT]
	case 0b0111: // SRL
		c.Registers[regD] = c.Registers[regS] >> c.Registers[regT]

	// Immediate Instructions
	case 0b1001: // ADDI
		c.Registers[regD] = c.Registers[regS] + uint16(imm)
	case 0b1010: // SUBI
		c.Registers[regD] = c.Registers[regS] - uint16(imm)

	// Load/Store
	case 0b1100: // LW
		addr := c.Registers[regS] + uint16(imm)
		c.Registers[regD] = binary.BigEndian.Uint16(c.Memory[addr:])
	case 0b1101: // SW
		addr := c.Registers[regS] + uint16(imm)
		binary.BigEndian.PutUint16(c.Memory[addr:], c.Registers[regD])

	// Branching
	case 0b1110: // BZ
		if c.Registers[regS] == 0 {
			c.Registers[PC] += uint16(imm) * 2 // Branch is relative
		}
	case 0b1111: // HALT
		c.Running = false

	default:
		log.Printf("Unknown instruction: 0x%04X\n", instruction)
		c.Running = false
	}

	// The status register would be updated here based on ALU results
}

// LoadProgram loads a program into memory at a specific address.
func (c *MonTanaMiniComputer) LoadProgram(program []byte, address uint16) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	copy(c.Memory[address:], program)
	c.Registers[PC] = address
}

// GetState returns a snapshot of the computer's state.
func (c *MonTanaMiniComputer) GetState() map[string]interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Create a map for named registers for easier display
	namedRegisters := map[string]uint16{
		"R0": c.Registers[R0], "R1": c.Registers[R1], "R2": c.Registers[R2], "R3": c.Registers[R3],
		"R4": c.Registers[R4], "R5": c.Registers[R5], "R6": c.Registers[R6], "R7": c.Registers[R7],
		"GP": c.Registers[GP], "FP": c.Registers[FP], "SP": c.Registers[SP], "RA": c.Registers[RA],
		"HI": c.Registers[HI], "LO": c.Registers[LO], "PC": c.Registers[PC], "SR": c.Registers[SR],
	}

	return map[string]interface{}{
		"registers":      c.Registers,
		"namedRegisters": namedRegisters,
		"running":        c.Running,
		"memory":         c.Memory[:256], // Send a portion of memory for display
	}
}
