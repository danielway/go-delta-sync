package golang

import (
	"errors"
	"github.com/JosephNaberhaus/go-delta-sync/agnostic/targets/test/generate"
	. "github.com/dave/jennifer/jen"
)

type GoTestImplementation struct {
	packageName string
	code        []Code
}

func (g *GoTestImplementation) Receiver() string {
	return "TestMode"
}

func (g *GoTestImplementation) Add(c ...Code) {
	g.code = append(g.code, c...)
}

func (g *GoTestImplementation) Write(fileName string) {
	jenFile := NewFile(g.packageName)
	jenFile.Add(lines(g.code...))
	err := jenFile.Save(fileName + ".go")
	if err != nil {
		panic(err)
	}
}

func (g *GoTestImplementation) Test(testCase generate.Case) {
	for _, fact := range testCase.Facts {
		testBody := make([]Code, 0)

		// Create an instance of the test model
		createTestModel := Id("model").Op(":=").Id("TestModel").Block()
		testBody = append(testBody, createTestModel)

		// Call the test model method
		inputs := make([]Code, 0, len(fact.Inputs))
		for _, input := range fact.Inputs {
			inputs = append(inputs, Lit(input))
		}
		if fact.Output == nil {
			callTestMethod := Id("model").Dot(testCase.Name).Call(inputs...)
			testBody = append(testBody, callTestMethod)
		} else {
			callTestMethod := Id("output").Op(":=").Id("model").Dot(testCase.Name).Call(inputs...)
			testBody = append(testBody, callTestMethod)
		}

		// Assert that the output matches the expected value
		if fact.Output != nil {
			assertOutput := Id("require").Dot("Equal").Call(Id("t"), Lit(fact.Output), Id("output"))
			testBody = append(testBody, assertOutput)
		}

		for _, sideEffect := range fact.SideEffects {
			assertSideEffect := Id("require").Dot("Equal").Call(Id("t"), resolveValue(sideEffect.ExpectedValue), Id("model").Dot(sideEffect.FieldName))
			testBody = append(testBody, assertSideEffect)
		}

		testName := "Test" + testCase.Name + fact.Name
		g.Add(Func().Id(testName).Params(Id("t").Op("*").Id("testing").Dot("t")).Block(testBody...))
	}
}

func TestImplementation(args map[string]string) generate.Implementation {
	packageName, ok := args["package"]
	if !ok {
		panic(errors.New("no package name supplied"))
	}

	return &GoTestImplementation{
		code:        make([]Code, 0),
		packageName: packageName,
	}
}
