package main

import (
	. "launchpad.net/gocheck"
)

func (s *MySuite) TestSpecWithHeadingAndSimpleSteps(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "Example step", lineNo: 3},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	c.Assert(spec.heading.lineNo, Equals, 1)
	c.Assert(spec.heading.value, Equals, "Spec Heading")

	c.Assert(len(spec.scenarios), Equals, 1)
	c.Assert(spec.scenarios[0].heading.lineNo, Equals, 2)
	c.Assert(spec.scenarios[0].heading.value, Equals, "Scenario Heading")
	c.Assert(len(spec.scenarios[0].steps), Equals, 1)
	c.Assert(spec.scenarios[0].steps[0].value, Equals, "Example step")
}

func (s *MySuite) TestStepsAndComments(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: commentKind, value: "A comment with some text and **bold** characters", lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&token{kind: commentKind, value: "Another comment", lineNo: 4},
		&token{kind: stepKind, value: "Example step", lineNo: 5},
		&token{kind: commentKind, value: "Third comment", lineNo: 6},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	c.Assert(spec.heading.value, Equals, "Spec Heading")

	c.Assert(len(spec.comments), Equals, 3)
	c.Assert(spec.comments[0].lineNo, Equals, 2)
	c.Assert(spec.comments[0].value, Equals, "A comment with some text and **bold** characters")

	c.Assert(len(spec.scenarios), Equals, 1)
	scenario := spec.scenarios[0]

	c.Assert(spec.comments[1].lineNo, Equals, 4)
	c.Assert(spec.comments[1].value, Equals, "Another comment")

	c.Assert(spec.comments[2].lineNo, Equals, 6)
	c.Assert(spec.comments[2].value, Equals, "Third comment")

	c.Assert(scenario.heading.value, Equals, "Scenario Heading")
	c.Assert(len(scenario.steps), Equals, 1)
}

func (s *MySuite) TestStepsWithParam(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: tableHeader, args: []string{"id"}, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&token{kind: stepKind, value: "enter {static} with {dynamic}", lineNo: 4, args: []string{"user", "id"}},
		&token{kind: stepKind, value: "sample \\{static\\}", lineNo: 5, args: []string{"user"}},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	step := spec.scenarios[0].steps[0]
	c.Assert(step.value, Equals, "enter {} with {}")
	c.Assert(step.lineNo, Equals, 4)
	c.Assert(len(step.args), Equals, 2)
	c.Assert(step.args[0].value, Equals, "user")
	c.Assert(step.args[0].argType, Equals, static)
	c.Assert(step.args[1].value, Equals, "id")
	c.Assert(step.args[1].argType, Equals, dynamic)

	escapedStep := spec.scenarios[0].steps[1]
	c.Assert(escapedStep.value, Equals, "sample \\{static\\}")
	c.Assert(len(escapedStep.args), Equals, 0)
}

func (s *MySuite) TestStepsWithKeywords(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "sample {static} and {dynamic}", lineNo: 3, args: []string{"name"}},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result, NotNil)
	c.Assert(result.ok, Equals, false)
	c.Assert(result.error.message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}' on line: 3")
}

func (s *MySuite) TestContextWithKeywords(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: context, value: "sample {static} and {dynamic}", lineNo: 3, args: []string{"name"}},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result, NotNil)
	c.Assert(result.ok, Equals, false)
	c.Assert(result.error.message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}' on line: 3")
}

func (s *MySuite) TestSpecWithDataTable(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading"},
		&token{kind: commentKind, value: "Comment before data table"},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
		&token{kind: commentKind, value: "Comment before data table"},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	c.Assert(spec.dataTable, NotNil)
	c.Assert(len(spec.dataTable.get("id")), Equals, 2)
	c.Assert(len(spec.dataTable.get("name")), Equals, 2)
	c.Assert(spec.dataTable.get("id")[0], Equals, "1")
	c.Assert(spec.dataTable.get("id")[1], Equals, "2")
	c.Assert(spec.dataTable.get("name")[0], Equals, "foo")
	c.Assert(spec.dataTable.get("name")[1], Equals, "bar")
}

func (s *MySuite) TestStepWithInlineTable(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "Step with inline table", lineNo: 3},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	inlineTable := spec.scenarios[0].steps[0].inlineTable
	c.Assert(inlineTable, NotNil)
	c.Assert(len(inlineTable.get("id")), Equals, 2)
	c.Assert(len(inlineTable.get("name")), Equals, 2)
	c.Assert(inlineTable.get("id")[0], Equals, "1")
	c.Assert(inlineTable.get("id")[1], Equals, "2")
	c.Assert(inlineTable.get("name")[0], Equals, "foo")
	c.Assert(inlineTable.get("name")[1], Equals, "bar")
}

func (s *MySuite) TestContextWithInlineTable(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading"},
		&token{kind: context, value: "Context with inline table"},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
		&token{kind: scenarioKind, value: "Scenario Heading"},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	inlineTable := spec.contexts[0].inlineTable

	c.Assert(inlineTable, NotNil)
	c.Assert(len(inlineTable.get("id")), Equals, 2)
	c.Assert(len(inlineTable.get("name")), Equals, 2)
	c.Assert(inlineTable.get("id")[0], Equals, "1")
	c.Assert(inlineTable.get("id")[1], Equals, "2")
	c.Assert(inlineTable.get("name")[0], Equals, "foo")
	c.Assert(inlineTable.get("name")[1], Equals, "bar")
}

func (s *MySuite) TestWarningWhenParsingMultipleDataTable(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading"},
		&token{kind: commentKind, value: "Comment before data table"},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
		&token{kind: commentKind, value: "Comment before data table"},
		&token{kind: tableHeader, args: []string{"phone"}, lineNo: 7},
		&token{kind: tableRow, args: []string{"1"}},
		&token{kind: tableRow, args: []string{"2"}},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	c.Assert(len(result.warnings), Equals, 1)
	c.Assert(result.warnings[0], Equals, "multiple data table present, ignoring table at line no: 7")

}

func (s *MySuite) TestWarningWhenParsingTableOccursWithoutStep(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: tableHeader, args: []string{"id", "name"}, lineNo: 3},
		&token{kind: tableRow, args: []string{"1", "foo"}, lineNo: 4},
		&token{kind: tableRow, args: []string{"2", "bar"}, lineNo: 5},
		&token{kind: stepKind, value: "Step", lineNo: 6},
		&token{kind: commentKind, value: "comment in between", lineNo: 7},
		&token{kind: tableHeader, args: []string{"phone"}, lineNo: 8},
		&token{kind: tableRow, args: []string{"1"}},
		&token{kind: tableRow, args: []string{"2"}},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)
	c.Assert(len(result.warnings), Equals, 2)
	c.Assert(result.warnings[0], Equals, "table not associated with a step, ignoring table at line no: 3")
	c.Assert(result.warnings[1], Equals, "table not associated with a step, ignoring table at line no: 8")

}

func (s *MySuite) TestAddSpecTags(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: specTag, args: []string{"tag1", "tag2"}, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.tags), Equals, 2)
	c.Assert(spec.tags[0], Equals, "tag1")
	c.Assert(spec.tags[1], Equals, "tag2")
}

func (s *MySuite) TestAddSpecTagsAndScenarioTags(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: specTag, args: []string{"tag1", "tag2"}, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&token{kind: scenarioTag, args: []string{"tag3", "tag4"}, lineNo: 2},
	}

	spec, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, true)

	c.Assert(len(spec.tags), Equals, 2)
	c.Assert(spec.tags[0], Equals, "tag1")
	c.Assert(spec.tags[1], Equals, "tag2")

	c.Assert(len(spec.scenarios[0].tags), Equals, 2)
	c.Assert(spec.scenarios[0].tags[0], Equals, "tag3")
	c.Assert(spec.scenarios[0].tags[1], Equals, "tag4")
}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutADataTable(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "Step with a {dynamic}", args: []string{"foo"}, lineNo: 3, lineText: "*Step with a <foo>"},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, false)
	c.Assert(result.error.message, Equals, "No data table found for dynamic paramter <foo> : *Step with a <foo> lineNo: 3")

}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutDataTableHeaderValue(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: tableHeader, args: []string{"id, name"}, lineNo: 2},
		&token{kind: tableRow, args: []string{"123, hello"}, lineNo: 3},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 4},
		&token{kind: stepKind, value: "Step with a {dynamic}", args: []string{"foo"}, lineNo: 5, lineText: "*Step with a <foo>"},
	}

	_, result := new(specParser).createSpecification(tokens)

	c.Assert(result.ok, Equals, false)
	c.Assert(result.error.message, Equals, "No data table column found for dynamic paramter <foo> : *Step with a <foo> lineNo: 5")

}
