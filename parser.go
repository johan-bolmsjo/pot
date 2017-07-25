package pot

// Parser interface implemented by Dict, DictKey, List and String parsers.
type Parser interface {
	// Parser name.
	Name() string

	// Get the next parser, nil or an error.
	Next() (Parser, error)

	// Get text the parser was initialized with.
	Bytes() []byte

	// Get parser start location in the original text input.
	// The location is reset when using NewParser, NewDictParser or NewListParser.
	Location() Location
}

// Root level parser capable of parsing multiple root level objects from the
// same text input.
type Root struct {
	org parserBuf  // Text the parser was initialized with.
	buf *parserBuf // Text buffer the parser operates on.
}

// Create a new root level parser parsing the supplied text.
func NewParser(pot []byte) Parser {
	buf := newParserBuf(pot)
	return &Root{*buf, buf}
}

func (root *Root) Name() string {
	return "root"
}

// Get the next parser or nil on end of input or an error.
// The returned parser may be a Dict, List or String.
func (root *Root) Next() (Parser, error) {
	return scanValue(root.buf)
}

// Get text the parser was initialized with.
func (root *Root) Bytes() []byte {
	return root.org.bytes
}

// Get parser start location in the original text input.
func (root *Root) Location() Location {
	return root.org.location
}

// Dictionary parser.
type Dict struct {
	org   parserBuf  // Text the parser was initialized with.
	buf   *parserBuf // Text buffer the parser operates on.
	count int        // Number of returned parsers.
}

// Create a new dictionary parser parsing the supplied text.
func NewDictParser(pot []byte) *Dict {
	return newDictParser(newParserBuf(pot))
}

// Create a new dictionary parser parsing the supplied parser buffer.
func newDictParser(buf *parserBuf) *Dict {
	dict := &Dict{*buf, buf, 0}
	buf.stripBlock('{', '}')
	// Trim space to make IsEmpty() work out of the gate.
	buf.trimSpaceLeft()
	return dict
}

func (dict *Dict) Name() string {
	return "dictionary"
}

// Get the next parser or nil on end of input or an error.
// Every even call returns a key which is of type DictKey.
// Every odd call returns a value which may be a Dict, List or String.
func (dict *Dict) Next() (parser Parser, err error) {
	if dict.count%2 == 0 {
		parser, err = scanKey(dict.buf)
	} else {
		parser, err = scanValue(dict.buf)
		if parser == nil && err == nil {
			err = dict.buf.errorf("key without value in dictionary")
		}
	}
	if parser != nil {
		dict.count++
	}
	return
}

// Check if the parser has consumed all data.
func (dict *Dict) IsEmpty() bool {
	return len(dict.buf.bytes) == 0
}

// Get text the parser was initialized with.
func (dict *Dict) Bytes() []byte {
	return dict.org.bytes
}

// Get parser start location in the original text input.
func (dict *Dict) Location() Location {
	return dict.org.location
}

// Dictionary key type.
// A unique "string" type is used to be able to separate keys from string values
// in type switches.
type DictKey String

func (key *DictKey) Name() string {
	return "dictionary-key"
}

// Returns nil as keys does not contain sub parsers.
func (key *DictKey) Next() (Parser, error) {
	return nil, nil
}

// Get text the parser was initialized with.
func (key *DictKey) Bytes() []byte {
	return (*String)(key).Bytes()
}

// Format as a POT dictionary key.
func (key *DictKey) String() string {
	return string(key.Bytes()) + ":"
}

// Get parser start location in the original text input.
func (key *DictKey) Location() Location {
	return (*String)(key).Location()
}

// List parser.
type List struct {
	org parserBuf  // Text the parser was initialized with.
	buf *parserBuf // Text buffer the parser operates on.
}

// Create a new list parser parsing the supplied text.
func NewListParser(pot []byte) *List {
	return newListParser(newParserBuf(pot))
}

// Create a new list parser parsing the supplied parser buffer.
func newListParser(buf *parserBuf) *List {
	list := &List{*buf, buf}
	buf.stripBlock('[', ']')
	return list
}

func (list *List) Name() string {
	return "list"
}

// Get the next parser or nil on end of input or an error.
// The returned parser may be a Dict, List or String.
func (list *List) Next() (Parser, error) {
	return scanValue(list.buf)
}

// Get text the parser was initialized with.
func (list *List) Bytes() []byte {
	return list.org.bytes
}

// Get parser start location in the original text input.
func (list *List) Location() Location {
	return list.org.location
}

// String parser.
type String parserBuf

func (str *String) Name() string {
	return "string"
}

// Returns nil as strings does not contain sub parsers.
func (str *String) Next() (Parser, error) {
	return nil, nil
}

var charToEscapeCode = map[byte]byte{
	'\n': 'n',
	'\r': 'r',
	'\t': 't',
	'\\': '\\',
	'"':  '"',
}

// Format as a POT string value.
func (str *String) String() string {
	quote := false
	newBuf := false

	t := str.bytes
	for i, c := range str.bytes {
		switch c {
		case '{', '}', '[', ']', ':', ' ':
			quote = true
		case '\n', '\r', '\t', '\\', '"':
			if !newBuf {
				t = make([]byte, 0, len(str.bytes))
				t = append(t, str.bytes[:i]...)
				newBuf = true
			}
			t = append(t, '\\', charToEscapeCode[c])
		default:
			if newBuf {
				t = append(t, c)
			}
		}
	}
	if quote || len(t) == 0 {
		return "\"" + string(t) + "\""
	}
	return string(t)
}

// Get text the parser was initialized with.
func (str *String) Bytes() []byte {
	return str.bytes
}

// Get parser start location in the original text input.
func (str *String) Location() Location {
	return str.location
}

// Scans a DictKey from a text buffer.
// Returns a parser or nil when there is no more input or an error.
func scanKey(buf *parserBuf) (Parser, error) {
	buf.trimSpaceLeft()
	if len(buf.bytes) == 0 {
		return nil, nil
	}
	for i, c := range buf.bytes {
		switch {
		case validKeyChar(i, c):
		case c == ':' && i > 0:
			t := buf.split(i)
			buf.trimBytesLeft(1) // Eat ':'
			return (*DictKey)(t), nil
		default:
			buf.trimBytesLeft(i)
			return nil, buf.errorf("invalid character '%c' in key", c)
		}
	}
	buf.trimAll()
	return nil, buf.errorf("end of input while parsing key")
}

// Scans a Dict, List or String from a text buffer.
// Returns a parser or nil when there is no more input or an error.
func scanValue(buf *parserBuf) (parser Parser, err error) {
	buf.trimSpaceLeft()
	if len(buf.bytes) > 0 {
		switch buf.bytes[0] {
		case '{':
			if buf, err = scanBlock(buf, '{', '}'); err == nil {
				parser = newDictParser(buf)
			}
		case '[':
			if buf, err = scanBlock(buf, '[', ']'); err == nil {
				parser = newListParser(buf)
			}
		default:
			parser, err = scanString(buf)
		}
	}
	return
}

// Scans a Dict or List block from a text buffer.
// Returns a new text buffer containing the block or an error.
func scanBlock(buf *parserBuf, begChar, endChar byte) (*parserBuf, error) {
	quoted := false
	escaped := false
	scope := 0
	for i, c := range buf.bytes {
		switch c {
		case '\\':
			escaped = !escaped
		case '"':
			if !escaped {
				quoted = !quoted
			}
			escaped = false
		case begChar:
			if !quoted && !escaped {
				scope++
			}
			escaped = false
		case endChar:
			if !quoted && !escaped {
				scope--
				if scope == 0 {
					return buf.split(i + 1), nil
				}
			}
			escaped = false
		default:
			escaped = false
		}
	}
	buf.trimAll()
	return nil, buf.errorf("end of input while parsing '%c%c' block", begChar, endChar)
}

// Scans a String from a text buffer.
// Returns a parser or an error.
func scanString(buf *parserBuf) (Parser, error) {
	quoted := false
	escaped := false
	eval := false

	var i int
	var c byte
loop:
	for i, c = range buf.bytes {
		switch c {
		case '\\':
			escaped = !escaped
			eval = true
		case '"':
			if !escaped {
				quoted = !quoted
			}
			escaped = false
			eval = true
		case '{', '}', '[', ']', ' ', '\n', '\r', '\t':
			if !quoted && !escaped {
				if i == 0 {
					buf.trimBytesLeft(i)
					return nil, buf.errorf("invalid character '%c' in string", c)
				} else {
					i--
					break loop
				}
			}
			escaped = false
		case ':':
			if !quoted && !escaped {
				buf.trimBytesLeft(i)
				return nil, buf.errorf("invalid character '%c' in string", c)
			}
			escaped = false
		}
	}

	str := buf.split(i + 1)
	if eval {
		return evalStringBuffer(str)
	}
	return (*String)(str), nil
}

var escapeCodeToChar = map[byte]byte{
	'n': '\n',
	'r': '\r',
	't': '\t',
}

// Evaluate escape codes and quotes in string.
// Returns the modified parser buffer as a string parser or an error.
func evalStringBuffer(buf *parserBuf) (Parser, error) {
	quoted := false
	escaped := false

	tr := make([]byte, 0, len(buf.bytes))
	for i, c := range buf.bytes {
		switch c {
		case '\\':
			if escaped {
				tr = append(tr, c)
			}
			escaped = !escaped
		case '"':
			if escaped {
				tr = append(tr, c)
			} else {
				quoted = !quoted
			}
			escaped = false
		case 'n', 'r', 't':
			if escaped {
				tr = append(tr, escapeCodeToChar[c])
			} else {
				tr = append(tr, c)
			}
			escaped = false
		case '{', '}', '[', ']', ':', ' ':
			tr = append(tr, c)
			escaped = false
		default:
			if escaped {
				buf.trimBytesLeft(i)
				return nil, buf.errorf("invalid escape code \\%c", c)
			} else {
				tr = append(tr, c)
			}
		}
	}

	if quoted {
		buf.trimAll()
		return nil, buf.errorf("miss-matched quotes in string")
	}
	if escaped {
		buf.trimAll()
		return nil, buf.errorf("unterminated escape code in string")
	}

	buf.bytes = tr
	return (*String)(buf), nil
}

// Check if 'c' is a valid key character at index 'i'.
func validKeyChar(i int, c byte) bool {
	if c >= 'a' && c < 'z' || c >= 'A' && c < 'Z' || c >= '0' && c < '9' {
		return true
	}
	if i > 0 && c == '-' {
		return true
	}
	return false
}
