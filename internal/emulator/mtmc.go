package emulator

import (
	"encoding/binary"
	"log"
	"sync"
	"time"
)

const (
	MemorySize = 4096
)

// MonTanaMiniComputer represents the state of the virtual computer.
type MonTanaMiniComputer struct {
	Memory    []byte
	Registers [16]int16
	PC        uint16 // Program Counter
	SP        uint16 // Stack Pointer
	Z         bool   // Zero flag
	N         bool   // Negative flag
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
	return &MonTanaMiniComputer{
		Memory: make([]byte, MemorySize),
	}
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
	if c.PC >= MemorySize-1 {
		log.Println("PC out of bounds, stopping execution.")
		c.Running = false
		return
	}

	instruction := binary.BigEndian.Uint16(c.Memory[c.PC:])
	c.PC += 2

	// Decode and execute the instruction
	opCode := (instruction & 0xF000) >> 12
	regA := (instruction & 0x0F00) >> 8
	regB := (instruction & 0x00F0) >> 4
	regC := instruction & 0x000F
	imm := int16(instruction & 0x00FF)

	switch opCode {
	case 0x1: // ADD
		c.Registers[regA] = c.Registers[regB] + c.Registers[regC]
	case 0x2: // SUB
		c.Registers[regA] = c.Registers[regB] - c.Registers[regC]
	case 0x3: // ADDI
		c.Registers[regA] = c.Registers[regB] + imm
	case 0x4: // LW
		addr := uint16(c.Registers[regB] + imm)
		if addr < MemorySize-1 {
			c.Registers[regA] = int16(binary.BigEndian.Uint16(c.Memory[addr:]))
		}
	case 0x5: // SW
		addr := uint16(c.Registers[regB] + imm)
		if addr < MemorySize-1 {
			binary.BigEndian.PutUint16(c.Memory[addr:], uint16(c.Registers[regA]))
		}
	case 0x8: // JZ
		if c.Z {
			c.PC = uint16(imm)
		}
	case 0x9: // JMP
		c.PC = uint16(c.Registers[regA])
	case 0xF: // HALT
		c.Running = false
	default:
		log.Printf("Unknown instruction: 0x%04X\n", instruction)
		c.Running = false
	}

	// Update flags (simplified)
	if opCode >= 1 && opCode <= 3 {
		c.Z = c.Registers[regA] == 0
		c.N = c.Registers[regA] < 0
	}
}

// LoadProgram loads a program into memory at a specific address.
func (c *MonTanaMiniComputer) LoadProgram(program []byte, address uint16) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	copy(c.Memory[address:], program)
	c.PC = address
}

// GetState returns a snapshot of the computer's state.
func (c *MonTanaMiniComputer) GetState() map[string]interface{} {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return map[string]interface{}{
		"registers": c.Registers,
		"pc":        c.PC,
		"sp":        c.SP,
		"z":         c.Z,
		"n":         c.N,
		"running":   c.Running,
		"memory":    c.Memory[:256], // Send a portion of memory for display
	}
}
