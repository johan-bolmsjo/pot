package pot

//-----------------------------------------------------------------------------

// Parser interface implemented by Dict, DictKey, List and String parsers.
type Parser interface {
	// Get the next parser, nil or an error.
	Next() (Parser, error)

	// Get the text the parser was initialized with.
	Text() []byte
}

//-----------------------------------------------------------------------------

// Root level parser capable of parsing multiple root level objects from the
// same text input.
type Root struct {
	org []byte     // Text the parser was initialized with.
	buf *parserBuf // Text buffer the parser operates on.
}

// Create a new root level parser parsing the supplied text.
func NewParser(text []byte) Parser {
	return &Root{text, newParserBuf(text)}
}

// Get the next parser or nil on end of input or an error.
// The returned parser may be a Dict, List or String.
func (p *Root) Next() (Parser, error) {
	return scanValue(p.buf)
}

// Get the text the parser was initialized with.
func (p *Root) Text() []byte {
	return p.org
}

//-----------------------------------------------------------------------------

// Dictionary parser.
type Dict struct {
	org   []byte     // Text the parser was initialized with.
	buf   *parserBuf // Text buffer the parser operates on.
	count int        // Number of returned parsers.
}

// Create a new dictionary parser parsing the supplied text.
func NewDictParser(text []byte) *Dict {
	return newDictParser(newParserBuf(text))
}

// Create a new dictionary parser parsing the supplied parser buffer.
func newDictParser(buf *parserBuf) *Dict {
	org := buf.head
	buf.stripBlock('{', '}')
	// Trim space to make IsEmpty() work out of the gate.
	buf.trimSpaceLeft()
	return &Dict{org, buf, 0}
}

// Get the next parser or nil on end of input or an error.
// Every even call returns a key which is of type DictKey.
// Every odd call returns a value which may be a Dict, List or String.
func (p *Dict) Next() (parser Parser, err error) {
	if p.count%2 == 0 {
		parser, err = scanKey(p.buf)
	} else {
		parser, err = scanValue(p.buf)
		if parser == nil && err == nil {
			err = p.buf.errorf("key without value in dictionary")
		}
	}
	if parser != nil {
		p.count++
	}
	return
}

// Check if the parser has consumed all data.
func (p *Dict) IsEmpty() bool {
	return p.buf.len() == 0
}

// Get the text the parser was initialized with.
func (p *Dict) Text() []byte {
	return p.org
}

//-----------------------------------------------------------------------------

// Dictionary key type.
// A unique "string" type is used to be able to separate keys from string values
// in type switches.
type DictKey String

// Returns nil as keys does not contain sub parsers.
func (v DictKey) Next() (Parser, error) {
	return nil, nil
}

// Get the text the parser was initialized with.
func (v DictKey) Text() []byte {
	return v
}

// Format as a POT dictionary key.
func (v DictKey) String() string {
	return string(v) + ":"
}

//-----------------------------------------------------------------------------

// List parser.
type List struct {
	org   []byte     // Text the parser was initialized with.
	buf   *parserBuf // Text buffer the parser operates on.
	count int        // Number of returned parsers.
}

// Create a new list parser parsing the supplied text.
func NewListParser(text []byte) *List {
	return newListParser(newParserBuf(text))
}

// Create a new list parser parsing the supplied parser buffer.
func newListParser(buf *parserBuf) *List {
	org := buf.head
	buf.stripBlock('[', ']')
	return &List{org, buf, 0}
}

// Get the next parser or nil on end of input or an error.
// The returned parser may be a Dict, List or String.
func (p *List) Next() (parser Parser, err error) {
	if parser, err = scanValue(p.buf); parser != nil {
		p.count++
	}
	return
}

// Get the text the parser was initialized with.
func (p *List) Text() []byte {
	return p.org
}

//-----------------------------------------------------------------------------

// String parser.
type String []byte

// Returns nil as strings does not contain sub parsers.
func (v String) Next() (Parser, error) {
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
func (v String) String() string {
	quote := false
	newBuf := false

	t := v
	for i, c := range v {
		switch c {
		case '{', '}', '[', ']', ':', ' ':
			quote = true
		case '\n', '\r', '\t', '\\', '"':
			if !newBuf {
				t = make([]byte, 0, len(v))
				t = append(t, v[:i]...)
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

// Get the text the parser was initialized with.
func (v String) Text() []byte {
	return v
}

//-----------------------------------------------------------------------------

// Scans a DictKey from a text buffer.
// Returns a parser or nil when there is no more input or an error.
func scanKey(buf *parserBuf) (Parser, error) {
	buf.trimSpaceLeft()
	if buf.len() == 0 {
		return nil, nil
	}
	for i, c := range buf.bytes() {
		switch {
		case validKeyChar(i, c):
		case c == ':' && i > 0:
			t := buf.split(i)
			buf.trimBytesLeft(1) // ':'
			return DictKey(t.bytes()), nil
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
	if buf.len() > 0 {
		switch buf.at(0) {
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
	for i, c := range buf.bytes() {
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
	for i, c = range buf.bytes() {
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
	return String(str.bytes()), nil
}

var escapeCodeToChar = map[byte]byte{
	'n': '\n',
	'r': '\r',
	't': '\t',
}

// Evaluate escape codes and quotes in string buffer.
// Returns a String parser or an error.
func evalStringBuffer(buf *parserBuf) (Parser, error) {
	quoted := false
	escaped := false

	tr := make([]byte, 0, buf.len())
	for i, c := range buf.bytes() {
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

	return String(tr), nil
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

//-----------------------------------------------------------------------------
