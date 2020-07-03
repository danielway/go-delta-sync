package blocks

import (
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/types"
)

type ModelName string

type Field struct {
	Name            string
	TypeDescription types.TypeDescription
}

type Implementation interface {
	Write(fileName string)
	Model(ModelName, ...Field)
	Method(modelName, methodName string, parameters ...Field) BodyImplementation
}

type BodyImplementation interface {
	// Assigns the value at assigned to assignee
	// Go Code: `<assignee> = <assigned>`
	Assign(assignee, assigned Value)
	// Declares a new variable with the given value
	// Go Code: `<declared> := <value>`
	Declare(declared VariableStruct, value Value)

	// Appends a value to the end of an array and ensures that the array value
	// points to the result. This comes with no guarantees that a different
	// reference to the array will not also be modified
	// Go Code: '<array> = append(<array>, <value>)`
	AppendValue(array, value Value)
	// Appends an array to the end of another array and ensures that the array
	// value points to the result. This comes with no guarantee that a
	// different reference to the array that was appended to will not also be
	// modified. However, the value array will not be altered
	// Go Code: `<array> = append(<array>, <valueArray>...)`
	AppendArray(array, valueArray Value)
	// Remove the value at index from the array. The order of the array must
	// not be altered by this operation and it must leave no gap from where the
	// element was removed.
	// Go Code: `<array> = append<array[:<index>], <array>[<index>+1:]...)`
	RemoveValue(array, index Value)

	// Sets key to value in the map, overriding an existing value or creating a
	// new entry a necessary
	// Go Code: `<mapValue>[<key>] = <value>`
	MapPut(mapValue, key, value Value)
	// Deleted the given value from the map. However this is performed the key
	// must be considered to no longer exist on the map
	// Go Code: `delete(<mapValue>, <key>)`
	MapDelete(mapValue, key Value)

	// Iterates through every value of the given array. Index name and value
	// are to be variables containing the equivalent of a zero based index and
	// the value at that index. An empty string for a name will  to indicate to
	// the implementation that the value is not used
	// Go Code: `for <indexName>, <valueName> := range <array> { <body> }
	ForEach(array Value, indexName, valueName string) BodyImplementation

	// Performs a comparison operation on the two values and executed the body
	// if the result is true
	// Go Code: `if <value1> <operator> <value2> { <body> }
	If(value1 Value, operator ComparisonOperator, value2 Value) BodyImplementation
	// Performs a comparison operation on the two values and executes the true
	// body if the result is true and the false body otherwise
	// Go Code: `if <value1> <operator> <value2> { <true block> } else { <false block> }
	IfElse(value1 Value, operator ComparisonOperator, value2 Value) (TrueBody, FalseBody BodyImplementation)
	// Executes the body if the value is true
	// Go Code: `if <value> { <body> }
	IfBool(value Value) BodyImplementation
	// Execute the true body if the value is true and the false body otherwise
	// Go Code: `if <value> { <true body> } else { <false body> }
	IfElseBool(value Value) (TrueBody, FalseBody BodyImplementation)
}
