//go:generate go run ../../scripts/generate-test.go --impl typescript --testSuffix .test

package typescript

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/types"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/blocks/value"
	"io"
	"os"
	"strconv"
	"strings"
)

type Code interface {
	Write(out io.Writer, indentLevel int) error
}

type OrphanCode struct {
	belongsTo string
	code      []Code
}

func (o *OrphanCode) Add(code ...Code) {
	o.code = append(o.code, code...)
}

func NewOrphanCode(belongsTo string) *OrphanCode {
	return &OrphanCode{
		belongsTo: belongsTo,
		code:      make([]Code, 0),
	}
}

type Line string

func (n Line) Write(out io.Writer, indentLevel int) error {
	_, err := io.WriteString(out, strings.Repeat("\t", indentLevel)+string(n)+"\n")
	return err
}

type Implementation struct {
	code        []Code
	modelBodies map[string]*BodyImplementation
	orphans     []*OrphanCode
}

func (i *Implementation) Add(code ...Code) {
	i.code = append(i.code, code...)
}

func (i *Implementation) RegisterModel(modelName string, body *BodyImplementation) {
	i.modelBodies[modelName] = body
}

func (i *Implementation) AddOrphan(orphan *OrphanCode) {
	i.orphans = append(i.orphans, orphan)
}

func NewImplementation(args map[string]string) agnostic.Implementation {
	return &Implementation{
		code:        make([]Code, 0),
		modelBodies: make(map[string]*BodyImplementation),
		orphans:     make([]*OrphanCode, 0),
	}
}

type BodyImplementation struct {
	code []Code
}

func (b *BodyImplementation) Add(code ...Code) {
	b.code = append(b.code, code...)
}

func (b *BodyImplementation) Write(out io.Writer, indentLevel int) error {
	for _, line := range b.code {
		err := line.Write(out, indentLevel+1)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewBodyImplementation() *BodyImplementation {
	return &BodyImplementation{code: make([]Code, 0)}
}

func (i *Implementation) Write(fileName string) {
	for _, orphan := range i.orphans {
		body, ok := i.modelBodies[orphan.belongsTo]
		if !ok {
			panic(errors.New("no model with name \"" + orphan.belongsTo + "\" found for method"))
		}

		body.Add(orphan.code...)
	}

	file, err := os.Create(fileName + ".ts")
	if err != nil {
		panic(err)
	}

	writer := bufio.NewWriter(file)

	for _, c := range i.code {
		err = c.Write(writer, 0)
		if err != nil {
			panic(err)
		}
	}

	err = writer.Flush()
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}
}

func (i *Implementation) Model(name string, fields ...agnostic.Field) {
	body := NewBodyImplementation()
	for _, field := range fields {
		body.Add(Line(field.Name + ": " + resolveType(field.Type) + ";"))
	}

	i.Add(Line("export class " + name + "{"))
	i.Add(body)
	i.Add(Line("}"))

	i.RegisterModel(name, body)
}

func (i *Implementation) Enum(name string, values ...string) {
	enumBody := NewBodyImplementation()
	for _, v := range values {
		enumBody.Add(Line(v))
	}

	i.Add(Line("enum " + name + "{"))
	i.Add(enumBody)
	i.Add(Line("}"))
}

func (i *Implementation) Method(modelName, methodName string, parameters ...agnostic.Field) agnostic.BodyImplementation {
	orphan := NewOrphanCode(modelName)
	i.AddOrphan(orphan)

	var parametersString strings.Builder
	for i, parameter := range parameters {
		parametersString.WriteString(parameter.Name + ": " + resolveType(parameter.Type))

		if i+1 != len(parameters) {
			parametersString.WriteString(", ")
		}
	}

	methodBody := NewBodyImplementation()

	orphan.Add(Line("public " + methodName + "(" + parametersString.String() + ") {"))
	orphan.Add(methodBody)
	orphan.Add(Line("}"))

	return methodBody
}

func (i *Implementation) ReturnMethod(modelName, methodName string, returnType types.Any, parameters ...agnostic.Field) agnostic.BodyImplementation {
	orphan := NewOrphanCode(modelName)
	i.AddOrphan(orphan)

	var parametersString strings.Builder
	for i, parameter := range parameters {
		parametersString.WriteString(parameter.Name + ": " + resolveType(parameter.Type))

		if i+1 != len(parameters) {
			parametersString.WriteString(", ")
		}
	}

	methodBody := NewBodyImplementation()

	orphan.Add(Line("public " + methodName + "(" + parametersString.String() + "): " + resolveType(returnType) + "{"))
	orphan.Add(methodBody)
	orphan.Add(Line("}"))

	return methodBody
}

func (b *BodyImplementation) Assign(assignee, assigned value.Any) {
	b.Add(Line(resolveValue(assignee) + " = " + resolveValue(assigned) + ";"))
}

func (b *BodyImplementation) Declare(name string, value value.Any) {
	b.Add(Line("let " + name + " = " + resolveValue(value) + ";"))
}

func (b *BodyImplementation) AppendValue(array, value value.Any) {
	b.Add(Line(resolveValue(array) + ".push(" + resolveValue(value) + ");"))
}

func (b *BodyImplementation) AppendArray(array, valueArray value.Any) {
	b.Add(Line(resolveValue(array) + ".push(..." + resolveValue(valueArray) + ");"))
}

func (b *BodyImplementation) RemoveValue(array, index value.Any) {
	b.Add(Line(resolveValue(array) + ".splice(" + resolveValue(index) + ", 1);"))
}

func (b *BodyImplementation) MapPut(mapValue, key, value value.Any) {
	b.Add(Line(resolveValue(mapValue) + ".set(" + resolveValue(key) + ", " + resolveValue(value) + ");"))
}

func (b *BodyImplementation) MapDelete(mapValue, key value.Any) {
	b.Add(Line(resolveValue(mapValue) + ".delete(" + resolveValue(key) + ");"))
}

func (b *BodyImplementation) ForEach(array value.Any, indexName, valueName string) agnostic.BodyImplementation {
	forEachBody := NewBodyImplementation()

	forEachParams := ""
	if indexName == "" {
		if valueName != "" {
			forEachParams = valueName
		}
	} else {
		if valueName == "" {
			forEachParams = "_, " + indexName
		} else {
			forEachParams = valueName + ", " + indexName
		}
	}

	b.Add(Line(resolveValue(array) + ".forEach((" + forEachParams + ") => {"))
	b.Add(forEachBody)
	b.Add(Line("});"))

	return forEachBody
}

func (b *BodyImplementation) If(value value.Any) agnostic.BodyImplementation {
	ifBody := NewBodyImplementation()

	b.Add(Line("if (" + resolveValue(value) + ") {"))
	b.Add(ifBody)
	b.Add(Line("}"))

	return ifBody
}

func (b *BodyImplementation) IfElse(value value.Any) (trueBody, falseBody agnostic.BodyImplementation) {
	ifBody := NewBodyImplementation()
	elseBody := NewBodyImplementation()

	b.Add(Line("if (" + resolveValue(value) + ") {"))
	b.Add(ifBody)
	b.Add(Line("} else {"))
	b.Add(elseBody)
	b.Add(Line("}"))

	return ifBody, elseBody
}

func (b *BodyImplementation) Return(value value.Any) {
	b.Add(Line("return " + resolveValue(value) + ";"))
}

func resolveType(any types.Any) string {
	switch t := any.(type) {
	case types.Base:
		return resolveBaseType(t)
	case types.Model:
		return t.ModelName()
	case types.Array:
		return resolveType(t.Element()) + "[]"
	case types.Map:
		return "Map<" + resolveType(t.Key()) + ", " + resolveType(t.Value()) + ">"
	case types.Pointer:
		panic(errors.New("pointers are not supported yet"))
	default:
		panic(errors.New(fmt.Sprintf("unkown type %T", t)))
	}
}

func resolveBaseType(base types.Base) string {
	switch base {
	case types.BaseBool:
		return "boolean"
	case types.BaseInt:
		fallthrough
	case types.BaseInt32:
		fallthrough
	case types.BaseInt64:
		fallthrough
	case types.BaseFloat32:
		fallthrough
	case types.BaseFloat64:
		return "number"
	case types.BaseString:
		return "string"
	default:
		panic(errors.New("unknown base type " + string(base)))
	}
}

func resolveValue(any value.Any) string {
	switch v := any.(type) {
	case value.Null:
		return "null"
	case value.String:
		return "\"" + v.Value() + "\""
	case value.Int:
		return strconv.Itoa(v.Value())
	case value.Float:
		return strconv.FormatFloat(v.Value(), 'f', -1, 64)
	case value.Array:
		var sb strings.Builder

		sb.WriteString("[")
		for i, element := range v.Elements() {
			sb.WriteString(resolveValue(element))

			if i+1 != len(v.Elements()) {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("]")

		return sb.String()
	case value.Map:
		var sb strings.Builder

		sb.WriteString("new ")
		sb.WriteString(resolveType(types.NewMap(v.KeyType(), v.ValueType())))
		sb.WriteString("([")
		for i, element := range v.Elements() {
			sb.WriteString("[")
			sb.WriteString(resolveValue(element.Key()))
			sb.WriteString(", ")
			sb.WriteString(resolveValue(element.Value()))
			sb.WriteString("]")

			if i-1 != len(v.Elements()) {
				sb.WriteString(", ")
			}
		}
		sb.WriteString("])")

		return sb.String()
	case value.OwnField:
		return "this." + resolveValue(v.Field())
	case value.Id:
		return v.Name()
	case value.ModelField:
		return v.ModelName() + "." + resolveValue(v.Field())
	case value.ArrayElement:
		return resolveValue(v.Array()) + "[" + resolveValue(v.Index()) + "]"
	case value.MapElement:
		return resolveValue(v.Map()) + ".get(" + resolveValue(v.Key()) + ")"
	case value.Combined:
		return resolveValue(v.Left()) + " " + v.Operator().Value() + " " + resolveValue(v.Right())
	case value.IntToString:
		return "String(" + resolveValue(v.IntValue()) + ")"
	default:
		panic(errors.New(fmt.Sprintf("uknown type %T", v)))
	}
}
