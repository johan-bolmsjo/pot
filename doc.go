/*
Package pot (Pieces of Text) provides a lightweight text serialization parser
that is intended to be used together with Go's encoding.TextUnmarshaler
interface.

There is only de-serialization support, it's easy enough to generate properly
formated POT directly from a encoding.Textmarshaler.


Format

The POT format is similar to JSON but with some differences required by the
application that POT was created for.

There are two compound types, a dictionary and a list and two string types. The
dictionary type may hold duplicate keys and the key order is maintained. This
makes it more of an itemized list than a dictionary type. The list type simply
holds a sequence of other types.

There are no numeric or boolean types, all parsing eventually produces
strings. It's up to an applications TextUnmarshaler functions to parse these
strings. Strings are separated by space, strings may contain space if quoted
or escaped.

The characters that may be used for keys in dictionaries are artificially
limited similar to variable names in most programming languages. The characters
a-z, A-Z, 0-9 are allowed in any position, the character '-' is allowed in any
position but the first. Dictionary keys are delimited by values by ':'.


Syntax Examples

Dictionaries:

	{ fruit: orange price: 10.5 }

	{ animal:	zebra
	  class:	mammal
	  weight-range: [ 240kg 370kg ] }

Lists:

	[ this is a list with seven strings ]

	[ "this is a list with one string" ]

Strings:

	this-is-a-string
	"this is a string"

Escape Codes

The escape character is '\'. Characters '{', '}', '[', ']', ':', ' ' must be
quoted or escaped in strings. Characters '\' and '"' must be escaped in
strings. Additionally '\n' produces a new-line, '\r' a carriage return and '\t'
a tab.


Usage

Create a new root level parser and call parser.Next() until it returns nil or an
error. There is also ParserScanner type that wraps a parser interface to provide
a bufio.Scanner like API

Example:

	parser := pot.NewParser([]byte("{ fruit: orange price: 10.5 }"))
	parse(parser)

	func parse(parser pot.Parser) {
		switch parser := parser.(type) {
		case *pot.DictKey:
			fmt.Printf("%s ", parser)
		case *pot.String:
			fmt.Printf("%s ", parser)
		}
		var node pot.Parser
		var err error
		for node, err = parser.Next(); node != nil; node, err = parser.Next() {
			parse(node)
		}
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}
*/
package pot
