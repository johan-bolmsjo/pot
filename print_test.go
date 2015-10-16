package pot

import (
	"fmt"
	"testing"
)

func testPrettyPrint(pot string) {
	buf, err := PrettyPrint([]byte(pot))
	if err != nil {
		fmt.Printf("error: %s\n", err)
	} else {
		fmt.Printf("%s\n", buf)
	}
}

var examplePrint1 = `
words as root parser strings [ and a list ]

{ a: dictionary that: { should: be indented: [ properly { and: { dictionaries: [ in a ] list: that } }  [ should\ not  ] ] } }

{ a: dictionary that: contains an: empty dictionary: {} }
{}
`

func ExamplePrettyPrint() {
	testPrettyPrint(examplePrint1)
	// Output:
	// words
	// as
	// root
	// parser
	// strings
	// [ and a list ]
	// {
	//     a: dictionary
	//     that: {
	//         should:   be
	//         indented: [ properly { and: { dictionaries: [ in a ] list: that } } [ "should not" ] ]
	//     }
	// }
	// {
	//     a:          dictionary
	//     that:       contains
	//     an:         empty
	//     dictionary: { }
	// }
	// {
	// }
}

// Exercise some functions not possible to run through examples.
func Test_PrintCoverage(t *testing.T) {
	es := new(errorSink)
	buf := newPrintBuf(es)
	buf.write([]byte("test"))
	buf.flush()
	es.e = fmt.Errorf("test")
	if bytes := buf.bytes(); bytes != nil {
		t.Errorf("printBuf.bytes() = %v want nil", bytes)
	}
}
