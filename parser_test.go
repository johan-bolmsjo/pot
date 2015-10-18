package pot

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func testParseString(pot string) {
	testParse(NewParser([]byte(pot)))
}

func testParse(parser Parser) {
	var buf bytes.Buffer
	if err := testParseDescent(&buf, parser); err != nil {
		fmt.Printf("error: %s\n", err)
	} else {
		fmt.Printf("%s\n", buf.Bytes())
	}
}

func testParseDescent(wr io.Writer, parser Parser) error {
	scanner := NewParserScanner(parser)
	for scanner.Scan() {
		switch parser := scanner.SubParser().(type) {
		case *Dict:
			fmt.Fprintf(wr, "{ ")
			if err := testParseDescent(wr, parser); err != nil {
				return err
			}
			fmt.Fprintf(wr, "} ")
		case *DictKey:
			fmt.Fprintf(wr, "%s ", parser)
		case *List:
			fmt.Fprintf(wr, "[ ")
			if err := testParseDescent(wr, parser); err != nil {
				return err
			}
			fmt.Fprintf(wr, "] ")
		case *String:
			fmt.Fprintf(wr, "%s ", parser)
		}
	}
	return scanner.Err()
}

func Example_ParserDict1() {
	testParse(NewDictParser([]byte("{ fruit: orange price: 10.5 }")))
	// Output:
	// fruit: orange price: 10.5
}

func Example_ParserDict2() {
	testParseString("{fruit:orange price:10.5}")
	// Output:
	// { fruit: orange price: 10.5 }
}

var example_ParserDict3 = `
        { animal:       zebra
          class:        mammal
          weight-range: [ 240kg 370kg ]
          foods:        [ "dry grass" apples ] }
`

func Example_ParserDict3() {
	testParseString(example_ParserDict3 + "\r")
	// Output:
	// { animal: zebra class: mammal weight-range: [ 240kg 370kg ] foods: [ "dry grass" apples ] }
}

func Example_ParserDict4() {
	testParseString("{ a: { aa: [] ab: {} } b: \"\"}")
	// Output:
	// { a: { aa: [ ] ab: { } } b: "" }
}

func Example_ParserDict5() {
	testParseString("{ -invalid-key: 0 }")
	// Output:
	// error: 1:2: invalid character '-' in key
}

func Example_ParserDict6() {
	testParseString("{ : foo }")
	// Output:
	// error: 1:2: invalid character ':' in key
}

func Example_ParserDict7() {
	testParseString("{ foo: }")
	// Output:
	// error: 1:7: key without value in dictionary
}

func Example_ParserDict8() {
	testParseString("{ unterminated-key}")
	// Output:
	// error: 1:18: end of input while parsing key
}

func Example_ParserList1() {
	testParse(NewListParser([]byte("[ unterminated\\ block")))
	// Output:
	// error: 1:21: end of input while parsing '[]' block
}

func Example_ParserString1() {
	testParseString("\"this is a long string\"")
	// Output:
	// "this is a long string"
}

func Example_ParserString2() {
	testParseString("this\\ is\\ a\\ long\\ string")
	// Output:
	// "this is a long string"
}

func Example_ParserString3() {
	testParseString("\"unterminated\\ quote")
	// Output:
	// error: 1:20: miss-matched quotes in string
}

func Example_ParserString4() {
	testParseString("\"escape codes: \"\\{\\}\\[\\]\\:\\ \\\\\\\"\\n\\r\\tthe-end")
	// Output:
	// "escape codes: {}[]: \\\"\n\r\tthe-end"
}

func Example_ParserString5() {
	testParseString("invalid-escape-code-\\m")
	// Output:
	// error: 1:21: invalid escape code \m
}

func Example_ParserString6() {
	testParseString("unterminated-escape-code-\\")
	// Output:
	// error: 1:26: unterminated escape code in string
}

func Example_ParserString7() {
	testParseString("]")
	// Output:
	// error: 1:0: invalid character ']' in string
}

func Example_ParserString8() {
	testParseString("\nunescaped-or-unquoted-colon-in-string:")
	// Output:
	// error: 2:37: invalid character ':' in string
}

// Test that stripping space from the right does not strip data that has already been consumed.
func TestParserBuf_TrimSpaceRight(t *testing.T) {
	buf := newParserBuf([]byte("    "))
	buf.trimBytesLeft(2)
	buf.trimSpaceRight()
	if buf.location.Column != 2 || len(buf.bytes) != 0 {
		t.Errorf("column = %d, len = %d", buf.location.Column, len(buf.bytes))
	}
}

func testCheckBytesAndLocation(t *testing.T, prefix string, parser Parser, pot string, location Location) {
	parserBytes := parser.Bytes()
	if !bytes.Equal(parserBytes, []byte(pot)) {
		t.Errorf("%s.Bytes() is '%s', expected '%s'", prefix, parserBytes, pot)
	}
	parserLocation := parser.Location()
	if parserLocation != location {
		t.Errorf("%s.Location() is '%s', expected '%s'", prefix, &parserLocation, location)
	}
}

func testParserBytesAndLocationFunctions(t *testing.T, scanner *ParserScanner) {
	for scanner.Scan() {
		switch parser := scanner.SubParser().(type) {
		case *Dict:
			testCheckBytesAndLocation(t, "Dict.Bytes()", parser, "{ key: value }", Location{0, 3})
			testParserBytesAndLocationFunctions(t, NewParserScanner(parser))
		case *DictKey:
			testCheckBytesAndLocation(t, "DictKey", parser, "key", Location{0, 5})
		case *List:
			testCheckBytesAndLocation(t, "List", parser, "[ { key: value } ]", Location{0, 1})
			testParserBytesAndLocationFunctions(t, NewParserScanner(parser))
		case *String:
			testCheckBytesAndLocation(t, "String", parser, "value", Location{0, 10})
		}
	}
	if err := scanner.Err(); err != nil {
		t.Error(err)
	}
}

// Test Bytes functions on parser interfaces.
func Test_ParserBytesAndLocationFunctions(t *testing.T) {
	parserStr := " [ { key: value } ]"
	parser := NewParser([]byte(parserStr))
	testCheckBytesAndLocation(t, "Root", parser, parserStr, Location{0, 0})
	testParserBytesAndLocationFunctions(t, NewParserScanner(parser))
}

// Test that injecting an error aborts parser iteration.
func TestParserScanner_InjectError(t *testing.T) {
	scanner := NewParserScanner(NewListParser([]byte("[ this is a list ]")))
	if !scanner.Scan() {
		t.Errorf("First call to Scan() returned false")
	}
	err := &ParseError{}
	scanner.InjectError(err)
	if scanner.Scan() {
		t.Errorf("Second call to Scan() returned true")
	}
	if errBack := scanner.Err().(*ParseError); err != errBack {
		t.Errorf("Injected error '%p' is not the same as the returned error '%p'", err, errBack)
	}
}

func TestLocation_Add(t *testing.T) {
	a := Location{1, 2}
	b := Location{3, 4}
	c := Location{5, 6}
	d := a.Add(&b).Add(&c)
	if d != &a {
		t.Errorf("Add() returned '%p', expected the receiver '%p'", d, &a)
	}
	e := Location{9, 12}
	if *d != e {
		t.Errorf("Add() resulted in '%v', expected '%v'", d, e)
	}
}

func TestLocation_Identifier(t *testing.T) {
	err := ParseError{
		Identifier: "",
		Location:   Location{1, 1},
		Message:    "test",
	}
	errs := [2]ParseError{err, err}
	errs[1].Identifier = "test"

	var errsOut [2]string
	errsExpect := [2]string{"2:1: test", "test:2:1: test"}

	for i := range errsOut {
		errsOut[i] = errs[i].Error()
		if errsOut[i] != errsExpect[i] {
			t.Errorf("Unexpected error message '%s', expected '%s'", errsOut[i], errsExpect[i])
		}
	}
}

// Exercise some functions not possible to run through examples.
func Test_ParserCoverage(t *testing.T) {
	var s String
	parser, err := s.Next()
	if parser != nil || err != nil {
		t.Errorf("String.Next() = (%v, %v) want (nil, nil)", parser, err)
	}

	var k DictKey
	parser, err = k.Next()
	if parser != nil || err != nil {
		t.Errorf("DictKey.Next() = (%v, %v) want (nil, nil)", parser, err)
	}

	parser = NewParser([]byte("something-fail:"))
	scanner := NewParserScanner(parser)
	for scanner.Scan() {
	}
	if err = scanner.Err(); err == nil {
		t.Errorf("ParserScanner.Error() = %v want !nil", err)
	}
	if parser = scanner.SubParser(); parser != nil {
		t.Errorf("ParserScanner.SubParser() = %v want nil", parser)
	}
}
