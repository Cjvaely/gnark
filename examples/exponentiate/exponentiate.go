package main

import (
	"fmt"

	"github.com/consensys/gnark/cs"
)

func main() {
	circuit := New()
	circuit.Write("circuit.r1cs")
}

const bitSize = 8 // number of bits of exponent

// New return the circuit implementing
// y == x**e
// only the bitSize least significant bits of e are used
func New() cs.CS {

	// create root constraint system
	circuit := cs.New()

	// declare secret and public inputs
	x := circuit.PUBLIC_INPUT("x")
	e := circuit.SECRET_INPUT("e")
	y := circuit.PUBLIC_INPUT("y")

	// specify constraints
	output := circuit.ALLOCATE(1)
	bits := circuit.TO_BINARY(e, bitSize)

	for i := 0; i < len(bits); i++ {

		bits[i].Tag(fmt.Sprintf("e[%d]", i)) // we can tag a variable for testing and / or debugging purposes, it has no impact on performances

		if i != 0 {
			output = circuit.MUL(output, output)
		}
		multiply := circuit.MUL(output, x)
		output = circuit.SELECT(bits[len(bits)-1-i], multiply, output)

		output.Tag(fmt.Sprintf("output after processing exponent bit %d", len(bits)-1-i))
	}

	circuit.MUSTBE_EQ(y, output)

	return circuit
}
