package register

type Register int

const (
	T0 Register = iota
	T1
	T2
	T3
	T4
	T5
	A0 // arg registers
	A1
	A2
	A3
	RV // return value
	RA // return address
	FP // frame pointer
	SP // stack pointer
	BP // break pointer
	PC

	//=== non-user-facing registers
	IR    // instruction register
	DR    // data register
	CB    // code boundary
	DB    // data boundary
	IO    // I/O register
	FLAGS // flags register

)

var registersByName = map[string]Register{
	"T0":    T0,
	"T1":    T1,
	"T2":    T2,
	"T3":    T3,
	"T4":    T4,
	"T5":    T5,
	"A0":    A0,
	"A1":    A1,
	"A2":    A2,
	"A3":    A3,
	"RV":    RV,
	"RA":    RA,
	"FP":    FP,
	"SP":    SP,
	"BP":    BP,
	"PC":    PC,
	"IR":    IR,
	"DR":    DR,
	"CB":    CB,
	"DB":    DB,
	"IO":    IO,
	"FLAGS": FLAGS,
}

var Registers = map[Register]string{
	T0:    "T0",
	T1:    "T1",
	T2:    "T2",
	T3:    "T3",
	T4:    "T4",
	T5:    "T5",
	A0:    "A0",
	A1:    "A1",
	A2:    "A2",
	A3:    "A3",
	RV:    "RV",
	RA:    "RA",
	FP:    "FP",
	SP:    "SP",
	BP:    "BP",
	PC:    "PC",
	IR:    "IR",
	DR:    "DR",
	CB:    "CB",
	DB:    "DB",
	IO:    "IO",
	FLAGS: "FLAGS",
}

func (r Register) IsWritable() bool {
	return 0 <= r && r < 16
}

func (r Register) IsReadable() bool {
	return 0 <= r && r < 16
}

func (r Register) IsTempRegister() bool {
	return 0 <= r && r < 6
}

func IsWritable(r string) bool {
	if v, ok := registersByName[r]; !ok {
		return false
	} else {
		return v.IsWritable()
	}
}

func IsReadable(r string) bool {
	if v, ok := registersByName[r]; !ok {
		return false
	} else {
		return v.IsReadable()
	}
}

func IsTempRegister(r string) bool {
	if v, ok := registersByName[r]; !ok {
		return false
	} else {
		return v.IsTempRegister()
	}

}
