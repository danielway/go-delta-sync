package value

import "github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/types"

// Refers to a literal null/nil/empty value
type Null struct {
	valueType
	methodIndependent
}

func NewNull() Null {
	return Null{}
}

// Refers to a literal string value
type String struct {
	valueType
	methodIndependent
	value string
}

func (s String) Value() string {
	return s.value
}

func NewString(value string) String {
	return String{value: value}
}

// Refers a literal int value
type Int struct {
	valueType
	methodIndependent
	value int
}

func (i Int) Value() int {
	return i.value
}

func NewInt(value int) Int {
	return Int{value: value}
}

// Refers to a literal floating point value
type Float struct {
	valueType
	methodIndependent
	value float64
}

func (f Float) Value() float64 {
	return f.value
}

func NewFloat(value float64) Float {
	return Float{value: value}
}

// Refers to a literal boolean value
type Bool struct {
	valueType
	methodIndependent
	value bool
}

func (b Bool) Value() bool {
	return b.value
}

func NewBool(value bool) Bool {
	return Bool{value: value}
}

// Refers to an array literal
type Array struct {
	valueType
	elementType types.Any
	elements    []Any
}

func (a Array) ElementType() types.Any {
	return a.elementType
}

func (a Array) Elements() []Any {
	return a.elements
}

func (a Array) IsMethodDependent() bool {
	for _, element := range a.elements {
		if element.IsMethodDependent() {
			return true
		}
	}

	return false
}

func NewArray(elementType types.Any, element ...Any) Array {
	return Array{
		elementType: elementType,
		elements:    element,
	}
}