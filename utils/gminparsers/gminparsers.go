/*
Copyright © 2020 Chris Duncan <chris.duncan@plusworx.uk>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package gminparsers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	gmess "github.com/plusworx/gmin/utils/gminmessages"
	lg "github.com/plusworx/gmin/utils/logging"
)

const (
	// Special tokens

	// ILLEGAL is an illegal character
	ILLEGAL Token = iota
	// EOS is end of string
	EOS
	// WS is whitespace
	WS

	// Literals

	//IDENT is field name
	IDENT

	// Misc characters

	// ASTERISK IS *
	ASTERISK
	// BSLASH is \
	BSLASH
	// CLOSEBRACK is )
	CLOSEBRACK
	// CLOSESQBRACK is ]
	CLOSESQBRACK
	// COLON is :
	COLON
	// COMMA is ,
	COMMA
	// EQUALS is =
	EQUALS
	// FSLASH is /
	FSLASH
	// GT is >
	GT
	// LT is <
	LT
	// OP is an operator
	OP
	// OPENBRACK is (
	OPENBRACK
	// OPENSQBRACK is [
	OPENSQBRACK
	// SINGLEQUOTE is '
	SINGLEQUOTE
	// TILDE is ~
	TILDE
	// UNDERSCORE is _
	UNDERSCORE
	// VALUE is query or input attribute value
	VALUE
)

// OutputAttrParser represents a parser of get and list attribute strings
type OutputAttrParser struct {
	oas *OutputAttrScanner
}

// Parse is the entry point for the parser
func (oap *OutputAttrParser) Parse(attrMap map[string]string) (*OutputAttrStr, error) {
	lg.Debugw("starting outputAttrParser Parse()",
		"attrMap", attrMap)
	defer lg.Debug("finished Parse()")

	attrStr := &OutputAttrStr{}

	for {
		tok, lit := oap.scanIgnoreWhitespace()
		if tok == EOS {
			break
		}

		if tok != IDENT && tok != ASTERISK && tok != OPENBRACK && tok != CLOSEBRACK &&
			tok != COMMA && tok != FSLASH && tok != TILDE {
			err := fmt.Errorf(gmess.ERR_UNEXPECTEDATTRCHAR, lit)
			lg.Error(err)
			return nil, err
		}

		if tok == IDENT && oap.oas.bCustomSchema == false {
			lowerLit := strings.ToLower(lit)
			validAttr := attrMap[lowerLit]
			if validAttr == "" {
				err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, lit)
				lg.Error(err)
				return nil, err
			}
			lit = validAttr
			if validAttr == "customSchemas" {
				oap.oas.bCustomSchema = true
			}
		}

		if tok == TILDE {
			lit = ","
			oap.oas.bCustomSchema = false
		}

		attrStr.Fields = append(attrStr.Fields, lit)
	}
	return attrStr, nil
}

// scan returns the next token from the underlying OutputAttrScanner
func (oap *OutputAttrParser) scan() (Token, string) {
	lg.Debug("starting outputAttrParser scan()")
	defer lg.Debug("finished scan()")

	// read the next token from the scanner
	tok, lit := oap.oas.Scan()

	return tok, lit
}

// scanIgnoreWhitespace scans the next non-whitespace token
func (oap *OutputAttrParser) scanIgnoreWhitespace() (Token, string) {
	lg.Debug("starting outputAttrParser scanIgnoreWhitespace()")
	defer lg.Debug("finished scanIgnoreWhitespace()")

	tok, lit := oap.scan()
	if tok == WS {
		tok, lit = oap.scan()
	}
	return tok, lit
}

// OutputAttrScanner represents a lexical scanner for get and list attribute strings
type OutputAttrScanner struct {
	bCustomSchema bool
	scanr         *Scanner
}

// Scan returns the next token and literal value from OutputAttrScanner
func (oas *OutputAttrScanner) Scan() (Token, string) {
	lg.Debug("starting outputAttrScanner Scan()")
	defer lg.Debug("finished Scan()")

	// Read the next rune
	ch := oas.scanr.read()

	// If we see whitespace then consume all contiguous whitespace
	// If we see a letter then consume as an ident
	if unicode.IsSpace(ch) {
		oas.scanr.unread()
		return oas.scanr.scanWhitespace()
	} else if unicode.IsLetter(ch) || ch == underscore {
		oas.scanr.unread()
		return oas.scanr.scanIdent()
	}

	// Otherwise read the individual character
	if ch == eos {
		return EOS, ""
	}
	if ch == '*' {
		return ASTERISK, string(ch)
	}
	if ch == ')' {
		return CLOSEBRACK, string(ch)
	}
	if ch == ',' {
		return COMMA, string(ch)
	}
	if ch == '/' {
		return FSLASH, string(ch)
	}
	if ch == '(' {
		return OPENBRACK, string(ch)
	}
	if ch == '~' {
		return TILDE, string(ch)
	}
	if ch == '_' {
		return UNDERSCORE, string(ch)
	}

	return ILLEGAL, string(ch)
}

// OutputAttrStr is a struct to hold list of output attribute string parts
type OutputAttrStr struct {
	Fields []string
}

// QueryParser represents a parser of query strings
type QueryParser struct {
	qs *QueryScanner
}

// scan returns the next token from the underlying QueryScanner
func (qp *QueryParser) scan() (Token, string) {
	lg.Debug("starting queryParser scan()")
	defer lg.Debug("finished scan()")

	// read the next token from the scanner
	tok, lit := qp.qs.Scan()

	return tok, lit
}

// scanIgnoreWhitespace scans the next non-whitespace token for QueryParser
func (qp *QueryParser) scanIgnoreWhitespace() (Token, string) {
	lg.Debug("starting queryParser scanIgnoreWhitespace()")
	defer lg.Debug("finished scanIgnoreWhitespace()")

	tok, lit := qp.scan()
	if tok == WS {
		tok, lit = qp.scan()
	}
	return tok, lit
}

// Parse is the entry point for the QueryParser
func (qp *QueryParser) Parse(qAttrMap map[string]string) (*QueryStr, error) {
	lg.Debugw("starting queryParser Parse()",
		"qAttrMap", qAttrMap)
	defer lg.Debug("finished Parse()")

	qStr := &QueryStr{}
	qp.qs.bConFieldName = true

	for {
		tok, lit := qp.scan()
		if tok == EOS {
			break
		}

		if tok != ASTERISK && tok != BSLASH && tok != COLON && tok != COMMA && tok != CLOSEBRACK &&
			tok != CLOSESQBRACK && tok != FSLASH && tok != GT && tok != EQUALS && tok != IDENT && tok != LT &&
			tok != OP && tok != OPENBRACK && tok != OPENSQBRACK && tok != TILDE && tok != VALUE {
			err := fmt.Errorf(gmess.ERR_UNEXPECTEDQUERYCHAR, lit)
			lg.Error(err)
			return nil, err
		}

		if tok == IDENT && !strings.Contains(lit, ".") {
			lowerLit := strings.ToLower(lit)
			validAttr := qAttrMap[lowerLit]
			if validAttr == "" {
				err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, lit)
				lg.Error(err)
				return nil, err
			}
			lit = validAttr
		}

		if tok == VALUE {
			lowerLit := strings.ToLower(lit)
			if lowerLit == "true" || lowerLit == "false" {
				lit = lowerLit
			}
		}

		if tok == TILDE {
			lit = " "
			qp.qs.bConFieldName = true
		}

		qStr.Parts = append(qStr.Parts, lit)
	}
	return qStr, nil
}

// QueryScanner represents a lexical scanner for query strings
type QueryScanner struct {
	bConFieldName bool
	bConOperator  bool
	bConValue     bool
	scanr         *Scanner
}

// Scan returns the next token and literal value from QueryScanner
func (qs *QueryScanner) Scan() (Token, string) {
	lg.Debug("starting queryScanner Scan()")
	defer lg.Debug("finished Scan()")

	// Read the next rune
	ch := qs.scanr.read()

	if qs.bConFieldName {
		// If we see whitespace then consume all contiguous whitespace
		// If we see a letter then consume as an ident
		if unicode.IsSpace(ch) {
			qs.scanr.unread()
			return qs.scanr.scanWhitespace()
		}
		if unicode.IsLetter(ch) || ch == underscore {
			qs.scanr.unread()
			tok, lit := qs.scanIdent()
			qs.bConFieldName = false
			qs.bConOperator = true
			return tok, lit
		}
	}

	if qs.bConOperator {
		qs.scanr.unread()
		tok, lit := qs.scanOperator()
		qs.bConOperator = false
		qs.bConValue = true
		return tok, lit
	}

	if qs.bConValue {
		qs.scanr.unread()
		tok, lit := qs.scanValue()
		qs.bConValue = false
		return tok, lit
	}

	// Otherwise read the individual character
	if ch == eos {
		return EOS, ""
	}
	if ch == '*' {
		return ASTERISK, string(ch)
	}
	if ch == '\\' {
		return BSLASH, string(ch)
	}
	if ch == ')' {
		return CLOSEBRACK, string(ch)
	}
	if ch == ']' {
		return CLOSESQBRACK, string(ch)
	}
	if ch == ':' {
		return COLON, string(ch)
	}
	if ch == ',' {
		return COMMA, string(ch)
	}
	if ch == '=' {
		return EQUALS, string(ch)
	}
	if ch == '/' {
		return FSLASH, string(ch)
	}
	if ch == '>' {
		return GT, string(ch)
	}
	if ch == '<' {
		return LT, string(ch)
	}
	if ch == '(' {
		return OPENBRACK, string(ch)
	}
	if ch == '[' {
		return OPENSQBRACK, string(ch)
	}
	if ch == '\'' {
		return SINGLEQUOTE, string(ch)
	}
	if ch == '~' {
		return TILDE, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanIdent consumes the current rune and all contiguous ident runes for QueryScanner
func (qs *QueryScanner) scanIdent() (Token, string) {
	lg.Debug("starting queryScanner scanIndent()")
	defer lg.Debug("finished scanIndent()")

	// Create a buffer and read the current character into it
	var (
		buf bytes.Buffer
		ch  rune
	)

	buf.WriteRune(qs.scanr.read())

	// Read every subsequent ident character into the buffer
	// Non-ident characters and EOS will cause the loop to exit
	for {
		ch = qs.scanr.read()
		if ch == eos {
			break
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != underscore {
			qs.scanr.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as a regular identifier
	return IDENT, buf.String()
}

// scanOperator consumes the current rune and all contiguous operator runes
func (qs *QueryScanner) scanOperator() (Token, string) {
	lg.Debug("starting queryScanner scanOperator()")
	defer lg.Debug("finished scanOperator()")

	// Create a buffer and read the current character into it
	var (
		buf bytes.Buffer
		ch  rune
	)

	buf.WriteRune(qs.scanr.read())

	// Read every subsequent operator character into the buffer
	// Non-operator characters and EOS will cause the loop to exit
	for {
		ch = qs.scanr.read()
		if ch == eos {
			break
		}
		if !isOperator(ch) {
			qs.scanr.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as an operator
	return OP, buf.String()
}

// scanValue consumes the current rune and all contiguous value runes
func (qs *QueryScanner) scanValue() (Token, string) {
	lg.Debug("starting queryScanner scanValue()")
	defer lg.Debug("finished scanValue()")

	// Create a buffer and read the current character into it
	var (
		buf bytes.Buffer
		ch  rune
	)

	buf.WriteRune(qs.scanr.read())

	// Read every subsequent value character into the buffer
	// Tilde character and EOS will cause the loop to exit
	for {
		ch = qs.scanr.read()
		if ch == eos {
			break
		}
		if ch == tilde {
			qs.scanr.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as a value
	return VALUE, buf.String()
}

// QueryStr is a struct to hold list of query string parts
type QueryStr struct {
	Parts []string
}

// Scanner represents a lexical scanner
type Scanner struct {
	strbuf *bytes.Buffer
}

// Token represents a lexical token.
type Token int

// read reads the next rune from the string
func (scanr *Scanner) read() rune {
	lg.Debug("starting Scanner read()")
	defer lg.Debug("finished read()")

	ch, _, err := scanr.strbuf.ReadRune()
	if err != nil {
		return eos
	}
	return ch
}

// scanIdent consumes the current rune and all contiguous ident runes
func (scanr *Scanner) scanIdent() (Token, string) {
	lg.Debug("starting Scanner scanIdent()")
	defer lg.Debug("finished scanIdent()")

	// Create a buffer and read the current character into it
	var (
		buf bytes.Buffer
		ch  rune
	)
	buf.WriteRune(scanr.read())

	// Read every subsequent ident character into the buffer
	// Non-ident characters and EOS will cause the loop to exit
	for {
		ch = scanr.read()
		if ch == eos {
			break
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != underscore {
			scanr.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}

	// Otherwise return as a regular identifier
	return IDENT, buf.String()
}

// scanWhitespace consumes the current rune and all contiguous whitespace
func (scanr *Scanner) scanWhitespace() (Token, string) {
	lg.Debug("starting Scanner scanWhitespace()")
	defer lg.Debug("finished scanWhitespace()")

	// Create a buffer and read the current character into it
	var (
		buf bytes.Buffer
		ch  rune
	)

	buf.WriteRune(scanr.read())

	// Read every subsequent whitespace character into the buffer
	// Non-whitespace characters and EOS will cause the loop to exit

	for {
		ch = scanr.read()
		if ch == eos {
			break
		}
		if !unicode.IsSpace(ch) {
			scanr.unread()
			break
		}
		buf.WriteRune(ch)
	}

	return WS, buf.String()
}

// unread places the previously read rune back on the reader
func (scanr *Scanner) unread() {
	lg.Debug("starting Scanner unread()")
	defer lg.Debug("finished unread()")

	_ = scanr.strbuf.UnreadRune()
}

// eos is end of string rune
var eos = rune(0)

// tilde is tilde rune
var tilde = '~'

// underscore is underscore rune
var underscore = '_'

// isOperator checks to see whether or not rune is an operator symbol
func isOperator(ch rune) bool {
	lg.Debugw("starting isOperator()",
		"ch", ch)
	defer lg.Debug("finished isOperator()")

	if ch == '=' || ch == ':' || ch == '>' || ch == '<' {
		return true
	}
	return false
}

// NewOutputAttrParser returns a new instance of OutputAttrParser
func NewOutputAttrParser(buf *bytes.Buffer) *OutputAttrParser {
	lg.Debug("starting NewOutputAttrParser()")
	defer lg.Debug("finished NewOutputAttrParser()")

	return &OutputAttrParser{oas: NewOutputAttrScanner(buf)}
}

// NewOutputAttrScanner returns a new instance of OutputAttrScanner
func NewOutputAttrScanner(buf *bytes.Buffer) *OutputAttrScanner {
	lg.Debug("starting NewOutputAttrScanner()")
	defer lg.Debug("finished NewOutputAttrScanner()")

	scanr := &Scanner{strbuf: buf}
	return &OutputAttrScanner{scanr: scanr}
}

// NewQueryParser returns a new instance of QueryParser
func NewQueryParser(buf *bytes.Buffer) *QueryParser {
	lg.Debug("starting NewQueryParser()")
	defer lg.Debug("finished NewQueryParser()")

	return &QueryParser{qs: NewQueryScanner(buf)}
}

// NewQueryScanner returns a new instance of QueryScanner
func NewQueryScanner(buf *bytes.Buffer) *QueryScanner {
	lg.Debug("starting NewQueryScanner()")
	defer lg.Debug("finished NewQueryScanner()")

	scanr := &Scanner{strbuf: buf}
	return &QueryScanner{scanr: scanr}
}

// ParseOutputAttrs validates attributes string and formats it for Get and List calls
func ParseOutputAttrs(attrs string, attrMap map[string]string) (string, error) {
	lg.Debugw("starting ParseOutputAttrs()",
		"attrs", attrs,
		"attrMap", attrMap)
	defer lg.Debug("finished ParseOutputAttrs()")

	bb := bytes.NewBufferString(attrs)

	p := NewOutputAttrParser(bb)

	as, err := p.Parse(attrMap)
	if err != nil {
		return "", err
	}

	outputStr := strings.Join(as.Fields, "")

	return outputStr, nil
}

// ParseQuery validates query string and formats it for queries in list calls
func ParseQuery(query string, qAttrMap map[string]string) (string, error) {
	lg.Debugw("starting ParseQuery()",
		"query", query,
		"qAttrMap", qAttrMap)
	defer lg.Debug("finished ParseQuery()")

	bb := bytes.NewBufferString(query)

	p := NewQueryParser(bb)

	qs, err := p.Parse(qAttrMap)
	if err != nil {
		return "", err
	}

	outputStr := strings.Join(qs.Parts, "")

	return outputStr, nil
}
