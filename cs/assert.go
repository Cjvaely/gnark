/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cs

import (
	"errors"
	"testing"

	"github.com/consensys/gnark/cs/internal/curve"
	"github.com/stretchr/testify/require"
)

// Assert is a helper to test circuits
type Assert struct {
	t *testing.T
	*require.Assertions
}

// NewAssert returns an helper to test Constraint Systems
// this helper embeds a stretch/testify Assert object for convenience
func NewAssert(t *testing.T) *Assert {
	return &Assert{t, require.New(t)}
}

// NotSolved check that a solution does NOT solve a circuit
// error may be missing inputs or unsatisfied constraints
func (assert *Assert) NotSolved(circuit CS, solution Assignments) {
	// sanity check that no assignement return an error if we need inputs
	assert.errInputNotSet(circuit)

	{
		r1cs := NewR1CS(&circuit)

		// solving with missing assignments should return a ErrInputNotSet
		nbInputs := r1cs.NbPrivateInputs() + r1cs.NbPublicInputs()
		if len(solution) < nbInputs {
			_, _, _, err := r1cs.Solve(solution)
			assert.Error(err, "solving R1CS with bad solution should return an error")
			assert.True(errors.Is(err, ErrInputNotSet), "expected ErrInputNotSet")
			return
		}

		if len(r1cs.Constraints) == 0 {
			assert.t.Log("circuit has no constraints, any input will solve it")
			return
		}
	}

	{
		r1cs := NewR1CS(&circuit)
		_, _, _, err := r1cs.Solve(solution)
		assert.Error(err, "solving R1CS with bad solution should return an error")
		assert.True(errors.Is(err, ErrUnsatisfiedConstraint) || errors.Is(err, ErrInputVisiblity), "expected ErrUnsatisfiedConstraint or ErrInputVisiblity")
	}
}

// Solved check that a solution solves a circuit
// for each expectedValues, this helper compares the output from r1cs.Inspect() after Solving.
// this helper also ensure the result vectors a*b=c
func (assert *Assert) Solved(circuit CS, solution Assignments, expectedValues map[string]interface{}) {
	// sanity check that no assignement return an error if we need inputs
	assert.errInputNotSet(circuit)

	{
		r1cs := NewR1CS(&circuit)
		a, b, c, err := r1cs.Solve(solution)
		assert.Nil(err, "solving R1CS with good solution shouldn't return an error")
		assert.Equal(len(a), len(b), "R1CS solution a,b and c vectors should be the same size")
		assert.Equal(len(b), len(c), "R1CS solution a,b and c vectors should be the same size")

		var tmp curve.Element
		for i := 0; i < len(a); i++ {
			assert.True(tmp.Mul(&a[i], &b[i]).Equal(&c[i]), "R1CS solution should be valid a * b = c rank 1 constriant")
		}

		values, err := r1cs.Inspect()
		assert.Nil(err, "inspecting values from R1CS after solving shouldn't return an error")

		for k, i := range expectedValues {
			got, ok := values[k]
			assert.True(ok, "expectedValues must be found in returned values from r1Inspect()")
			v := elementFromInterface(i)
			assert.True(v.Equal(&got), "at tag "+k+" expected "+v.String()+" got "+got.String())
		}

	}
}

// -------------------------------------------------------------------------------------------------
// internal

type expectedCS struct {
	nbWires, nbMOConstraints, nbNOConstraints int
	nbConstraints                             uint64
}

type expectedR1CS struct {
	nbWires, nbComputationalConstraints, nbConstraints, privateInputStartIndex, publicInputStartIndex int
}

func (assert *Assert) csIsCorrect(circuit CS, expectedCS expectedCS) {
	assert.Equal(expectedCS.nbWires, circuit.countWires(), "cs nbWires")
	assert.Equal(expectedCS.nbConstraints, circuit.nbConstraints, "cs nbConstraints")
	assert.Equal(expectedCS.nbMOConstraints, len(circuit.MOConstraints), "cs nb MOConstraints")
	assert.Equal(expectedCS.nbNOConstraints, len(circuit.NOConstraints), "cs nb NOConstraints")
}

func (assert *Assert) r1csIsCorrect(circuit CS, expectedR1CS expectedR1CS) {
	r1cs := NewR1CS(&circuit)
	assert.Equal(expectedR1CS.nbWires, len(r1cs.WireTracker), "r1cs nbWires")
	assert.Equal(expectedR1CS.nbComputationalConstraints, len(r1cs.ComputationalGraph), "r1cs nbComputationalConstraints")
	assert.Equal(expectedR1CS.nbConstraints, len(r1cs.Constraints), "r1cs nbConstraints")
	assert.Equal(expectedR1CS.publicInputStartIndex, r1cs.PublicInputsStartIndex, "r1cs public input start index")
	assert.Equal(expectedR1CS.privateInputStartIndex, r1cs.PrivateInputsStartIndex, "r1cs private input start index")
}

func (assert *Assert) errInputNotSet(circuit CS) {
	r1cs := NewR1CS(&circuit)

	nbInputs := r1cs.NbPrivateInputs() + r1cs.NbPublicInputs()
	if nbInputs > 0 {
		_, _, _, err := r1cs.Solve(NewAssignment())
		assert.Error(err, "solving R1CS without assignments should return an error")
		assert.True(errors.Is(err, ErrInputNotSet), "expected ErrInputNotSet")
	}
}
