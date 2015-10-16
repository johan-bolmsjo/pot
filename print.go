package pot

// Number of space characters used per indentation level.
const indentSize = 4

// Space character buffer used for indentation.
var spaceBuffer []byte

// Return a buffer of space characters corresponding to the specified
// indentation level.
func indentSpace(indentLevel int) []byte {
	nbSpaces := indentLevel * indentSize
	for i := len(spaceBuffer); i < nbSpaces; i++ {
		spaceBuffer = append(spaceBuffer, ' ')
	}
	return spaceBuffer[:nbSpaces]
}

// Pretty print POT text buffer.
// Returns a byte slice or an error on parsing errors.
func PrettyPrint(pot []byte) ([]byte, error) {
	buf := newPrintBuf(nil)
	scanner := newParserScannerErrorSink(NewParser(pot), buf.errorSink())
	for scanner.Scan() {
		switch subparser := scanner.SubParser().(type) {
		case *Dict:
			prettyPrintDict(buf, subparser, 0)
		case *List:
			printList(buf, subparser)
		case *String:
			buf.printf("%s", subparser)
		}
		buf.term('\n')
	}

	buf.flush()
	return buf.bytes(), buf.err()
}

// Pretty print dictionary.
// Any errors are sent to the print buffer's error sink.
func prettyPrintDict(buf *printBuf, dict *Dict, indentLevel int) {
	var key *DictKey
	spaces0 := indentSpace(indentLevel)
	spaces1 := indentSpace(indentLevel + 1)

	buf.write([]byte("{\n"))
	scanner := newParserScannerErrorSink(dict, buf.errorSink())
	for scanner.Scan() {
		switch subparser := scanner.SubParser().(type) {
		case *Dict:
			if subparser.IsEmpty() {
				buf.printf("%s%s\t{ }\n", spaces1, key)
			} else {
				buf.printf("%s%s ", spaces1, key)
				prettyPrintDict(buf, subparser, indentLevel+1)
			}
		case *DictKey:
			key = subparser
		case *List:
			buf.printf("%s%s\t", spaces1, key)
			printList(buf, subparser)
			buf.term('\n')
		case *String:
			buf.printf("%s%s\t%s\n", spaces1, key, subparser)
		}
	}
	buf.write(spaces0)
	buf.write([]byte("}")).term('\n')
}

// Print dictionary on a single line.
// Any errors are sent to the print buffer's error sink.
func printDict(buf *printBuf, dict *Dict) {
	var key *DictKey

	buf.write([]byte("{ "))
	scanner := newParserScannerErrorSink(dict, buf.errorSink())
	for scanner.Scan() {
		switch subparser := scanner.SubParser().(type) {
		case *Dict:
			buf.printf("%s ", key)
			printDict(buf, subparser)
		case *DictKey:
			key = subparser
		case *List:
			buf.printf("%s ", key)
			printList(buf, subparser)
		case *String:
			buf.printf("%s %s ", key, subparser)
		}
	}
	buf.write([]byte("}")).term(' ')
}

// Print list on a single line.
// Any errors are sent to the print buffer's error sink.
func printList(buf *printBuf, list *List) {
	buf.write([]byte("[ "))
	scanner := newParserScannerErrorSink(list, buf.errorSink())
	for scanner.Scan() {
		switch subparser := scanner.SubParser().(type) {
		case *Dict:
			printDict(buf, subparser)
		case *List:
			printList(buf, subparser)
		case *String:
			buf.printf("%s ", subparser)
		}
	}
	buf.write([]byte("]")).term(' ')
}
