package pot

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func testParseString(text string) {
	var buf bytes.Buffer
	if err := testParse(&buf, NewParser([]byte(text))); err != nil {
		fmt.Printf("error: %s\n", err)
	} else {
		fmt.Printf("%s\n", buf.Bytes())
	}
}

func testParse(wr io.Writer, parser Parser) error {
	ps := NewParserScanner(parser)
	for ps.Scan() {
		switch subparser := ps.Result().(type) {
		case *Dict:
			fmt.Fprintf(wr, "{ ")
			if err := testParse(wr, subparser); err != nil {
				return err
			}
			fmt.Fprintf(wr, "} ")
		case DictKey:
			fmt.Fprintf(wr, "%s ", subparser)
		case *List:
			fmt.Fprintf(wr, "[ ")
			if err := testParse(wr, subparser); err != nil {
				return err
			}
			fmt.Fprintf(wr, "] ")
		case String:
			fmt.Fprintf(wr, "%s ", subparser)
		}
	}
	return ps.Error()
}

func ExampleParserDict1() {
	testParseString("{ fruit: orange price: 10.5 }")
	// Output:
	// { fruit: orange price: 10.5 }
}

func ExampleParserDict2() {
	testParseString("{fruit:orange price:10.5}")
	// Output:
	// { fruit: orange price: 10.5 }
}

var exampleParserDict3 = `
        { animal:       zebra
          class:        mammal
          weight-range: [ 240kg 370kg ]
          foods:        [ "dry grass" apples ] }
`

func ExampleParserDict3() {
	testParseString(exampleParserDict3)
	// Output:
	// { animal: zebra class: mammal weight-range: [ 240kg 370kg ] foods: [ "dry grass" apples ] }
}

func ExampleParserDict4() {
	testParseString("{ a: { aa: [] ab: {} } b: \"\"}")
	// Output:
	// { a: { aa: [ ] ab: { } } b: "" }
}

func ExampleParserDict5() {
	testParseString("{ -invalid-key: 0 }")
	// Output:
	// error: 1:2: invalid character '-' in key
}

func ExampleParserDict6() {
	testParseString("{ : foo }")
	// Output:
	// error: 1:2: invalid character ':' in key
}

func ExampleParserDict7() {
	testParseString("{ foo: }")
	// Output:
	// error: 1:7: key without value in dictionary
}

func ExampleParserDict8() {
	testParseString("{ unterminated-key}")
	// Output:
	// error: 1:18: end of input while parsing key
}

func ExampleParserList1() {
	testParseString("[ unterminated\\ block")
	// Output:
	// error: 1:21: end of input while parsing '[]' block
}

func ExampleParserString1() {
	testParseString("\"this is a long string\"")
	// Output:
	// "this is a long string"
}

func ExampleParserString2() {
	testParseString("this\\ is\\ a\\ long\\ string")
	// Output:
	// "this is a long string"
}

func ExampleParserString3() {
	testParseString("\"unterminated\\ quote")
	// Output:
	// error: 1:20: miss-matched quotes in string
}

func ExampleParserString4() {
	testParseString("\"escape codes: \"\\{\\}\\[\\]\\:\\ \\\\\\\"\\n\\r\\tthe-end")
	// Output:
	// "escape codes: {}[]: \\\"\n\r\tthe-end"
}

func ExampleParserString5() {
	testParseString("invalid-escape-code-\\m")
	// Output:
	// error: 1:21: invalid escape code \m
}

func ExampleParserString6() {
	testParseString("unterminated-escape-code-\\")
	// Output:
	// error: 1:26: unterminated escape code in string
}

func ExampleParserString7() {
	testParseString("]")
	// Output:
	// error: 1:0: invalid character ']' in string
}

func ExampleParserString8() {
	testParseString("\nunescaped-or-unquoted-colon-in-string:")
	// Output:
	// error: 2:37: invalid character ':' in string
}

// Exercise some functions not possible to run through examples.
func TestParserCoverage(t *testing.T) {
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
	ps := NewParserScanner(parser)
	for ps.Scan() {
	}
	if err = ps.Error(); err == nil {
		t.Errorf("ParserScanner.Error() = %v want !nil", err)
	}
	if parser = ps.Result(); parser != nil {
		t.Errorf("ParserScanner.Result() = %v want nil", parser)
	}

	dictTextInput := "{}"
	dict := NewDictParser([]byte(dictTextInput))
	dictTextOutput := string(dict.Text())
	if dictTextInput != dictTextOutput {
		t.Errorf("Dict.Text() = %v want %v", dictTextOutput, dictTextInput)
	}

	listTextInput := "[]"
	list := NewListParser([]byte(listTextInput))
	listTextOutput := string(list.Text())
	if listTextInput != listTextOutput {
		t.Errorf("List.Text() = %v want %v", listTextOutput, listTextInput)
	}

	rootTextInput := "{ hello: world }"
	root := NewParser([]byte(rootTextInput))
	rootTextOutput := string(root.Text())
	if rootTextInput != rootTextOutput {
		t.Errorf("Root.Text() = %v want %v", rootTextOutput, rootTextInput)
	}

	dictKey := DictKey([]byte("hello"))
	dictKey.Text()

	potString := String([]byte("world"))
	potString.Text()
}
