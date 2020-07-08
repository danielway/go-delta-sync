package test

import (
	"github.com/JosephNaberhaus/go-delta-sync/agnostic"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/types"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/value"
)

// A collection of tests
type Suite = []Case

// A function that takes the given body implementation and the method that the
// test will car
type GenerateBodyFunc func(body agnostic.BodyImplementation)

// A method should be created
type Case struct {
	Name        string           // Name of the test case (must be unique)
	Description string           // Describes what the test case is for
	Parameters  []agnostic.Field // Parameters that the generated test method will take in
	Returns     types.Any        // The return type of the method or nil if it returns nothing
	Generator   GenerateBodyFunc // Function that generates the method that the test will target
	Facts       []Fact           // Facts about the Test
}

// A change that happens to the model as a result of a method call
type SideEffect struct {
	FieldName     string    // Name of the field
	ExpectedValue value.Any // The expected value of the field
}

func NewSideEffect(fieldName string, expectedValue value.Any) SideEffect {
	return SideEffect{
		FieldName:     fieldName,
		ExpectedValue: expectedValue,
	}
}

// An assertion that calling the method with the given inputs will result in
// the given side effects and output
type Fact struct {
	Name        string       // A descriptive name for the fact (must be unique in the context of a case)
	Inputs      []value.Any  // Values to pass in a parameters or nil
	SideEffects []SideEffect // Side effects of the method call or nil
	Output      value.Any    // The output of the method or nil
}