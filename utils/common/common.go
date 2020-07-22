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

package common

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"unicode"

	"crypto/sha1"

	cfg "github.com/plusworx/gmin/utils/config"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const (
	// HashFunction specifies password hash function
	HashFunction string = "SHA-1"
	// RepeatTxt is used to signal a repeated attribute in attribute argument processing
	RepeatTxt string = "**** repeat ****"
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
	// VALUE is query or input attribute value
	VALUE
)

// OutputAttrStr is a struct to hold list of output attribute string parts
type OutputAttrStr struct {
	Fields []string
}

// QueryStr is a struct to hold list of query string parts
type QueryStr struct {
	Parts []string
}

// OutputAttrParser represents a parser of get and list attribute strings
type OutputAttrParser struct {
	oas *OutputAttrScanner
}

// Parse is the entry point for the parser
func (oap *OutputAttrParser) Parse(attrMap map[string]string) (*OutputAttrStr, error) {
	attrStr := &OutputAttrStr{}

	for {
		tok, lit := oap.scanIgnoreWhitespace()
		if tok == EOS {
			break
		}

		if tok != IDENT && tok != ASTERISK && tok != OPENBRACK && tok != CLOSEBRACK &&
			tok != COMMA && tok != FSLASH && tok != TILDE {
			return nil, fmt.Errorf("gmin: unexpected character %q found in attribute string", lit)
		}

		if tok == IDENT && oap.oas.bCustomSchema == false {
			lowerLit := strings.ToLower(lit)
			validAttr := attrMap[lowerLit]
			if validAttr == "" {
				err := fmt.Errorf("gmin: error - attribute %v is unrecognized", lit)
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
func (oap *OutputAttrParser) scan() (tok Token, lit string) {
	// read the next token from the scanner
	tok, lit = oap.oas.Scan()

	return tok, lit
}

// scanIgnoreWhitespace scans the next non-whitespace token
func (oap *OutputAttrParser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = oap.scan()
	if tok == WS {
		tok, lit = oap.scan()
	}
	return tok, lit
}

// OutputAttrScanner represents a lexical scanner for get and list attribute strings
type OutputAttrScanner struct {
	bCustomSchema bool
	s             *Scanner
}

// Scan returns the next token and literal value from OutputAttrScanner
func (oas *OutputAttrScanner) Scan() (tok Token, lit string) {
	// Read the next rune
	ch := oas.s.read()

	// If we see whitespace then consume all contiguous whitespace
	// If we see a letter then consume as an ident
	if unicode.IsSpace(ch) {
		oas.s.unread()
		return oas.s.scanWhitespace()
	} else if unicode.IsLetter(ch) {
		oas.s.unread()
		return oas.s.scanIdent()
	}

	// Otherwise read the individual character
	switch ch {
	case eos:
		return EOS, ""
	case '*':
		return ASTERISK, string(ch)
	case ')':
		return CLOSEBRACK, string(ch)
	case ',':
		return COMMA, string(ch)
	case '/':
		return FSLASH, string(ch)
	case '(':
		return OPENBRACK, string(ch)
	case '~':
		return TILDE, string(ch)
	}

	return ILLEGAL, string(ch)
}

// QueryParser represents a parser of query strings
type QueryParser struct {
	qs *QueryScanner
}

// scan returns the next token from the underlying QueryScanner
func (qp *QueryParser) scan() (tok Token, lit string) {
	// read the next token from the scanner
	tok, lit = qp.qs.Scan()

	return tok, lit
}

// scanIgnoreWhitespace scans the next non-whitespace token for QueryParser
func (qp *QueryParser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = qp.scan()
	if tok == WS {
		tok, lit = qp.scan()
	}
	return tok, lit
}

// Parse is the entry point for the QueryParser
func (qp *QueryParser) Parse(qAttrMap map[string]string) (*QueryStr, error) {
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
			return nil, fmt.Errorf("gmin: unexpected character %q found in query string", lit)
		}

		if tok == IDENT && !strings.Contains(lit, ".") {
			lowerLit := strings.ToLower(lit)
			validAttr := qAttrMap[lowerLit]
			if validAttr == "" {
				err := fmt.Errorf("gmin: error - query attribute %v is unrecognized", lit)
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
	s             *Scanner
}

// Scan returns the next token and literal value from QueryScanner
func (qs *QueryScanner) Scan() (tok Token, lit string) {
	// Read the next rune
	ch := qs.s.read()

	if qs.bConFieldName {
		// If we see whitespace then consume all contiguous whitespace
		// If we see a letter then consume as an ident
		if unicode.IsSpace(ch) {
			qs.s.unread()
			return qs.s.scanWhitespace()
		} else if unicode.IsLetter(ch) {
			qs.s.unread()
			tok, lit := qs.scanIdent()
			qs.bConFieldName = false
			qs.bConOperator = true
			return tok, lit
		}
	}

	if qs.bConOperator {
		qs.s.unread()
		tok, lit := qs.scanOperator()
		qs.bConOperator = false
		qs.bConValue = true
		return tok, lit
	}

	if qs.bConValue {
		qs.s.unread()
		tok, lit := qs.scanValue()
		qs.bConValue = false
		return tok, lit
	}

	// Otherwise read the individual character
	switch ch {
	case eos:
		return EOS, ""
	case '*':
		return ASTERISK, string(ch)
	case '\\':
		return BSLASH, string(ch)
	case ')':
		return CLOSEBRACK, string(ch)
	case ']':
		return CLOSESQBRACK, string(ch)
	case ':':
		return COLON, string(ch)
	case ',':
		return COMMA, string(ch)
	case '=':
		return EQUALS, string(ch)
	case '/':
		return FSLASH, string(ch)
	case '>':
		return GT, string(ch)
	case '<':
		return LT, string(ch)
	case '(':
		return OPENBRACK, string(ch)
	case '[':
		return OPENSQBRACK, string(ch)
	case '\'':
		return SINGLEQUOTE, string(ch)
	case '~':
		return TILDE, string(ch)
	}

	return ILLEGAL, string(ch)
}

// scanIdent consumes the current rune and all contiguous ident runes for QueryScanner
func (qs *QueryScanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(qs.s.read())

	// Read every subsequent ident character into the buffer
	// Non-ident characters and EOS will cause the loop to exit
	for {
		if ch := qs.s.read(); ch == eos {
			break
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' {
			qs.s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier
	return IDENT, buf.String()
}

// scanOperator consumes the current rune and all contiguous operator runes
func (qs *QueryScanner) scanOperator() (tok Token, lit string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(qs.s.read())

	// Read every subsequent operator character into the buffer
	// Non-operator characters and EOS will cause the loop to exit
	for {
		if ch := qs.s.read(); ch == eos {
			break
		} else if !isOperator(ch) {
			qs.s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as an operator
	return OP, buf.String()
}

// scanValue consumes the current rune and all contiguous value runes
func (qs *QueryScanner) scanValue() (tok Token, lit string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(qs.s.read())

	// Read every subsequent value character into the buffer
	// Tilde character and EOS will cause the loop to exit
	for {
		if ch := qs.s.read(); ch == eos {
			break
		} else if ch == tilde {
			qs.s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a value
	return VALUE, buf.String()
}

// Scanner represents a lexical scanner
type Scanner struct {
	strbuf *bytes.Buffer
}

// read reads the next rune from the string
func (s *Scanner) read() rune {
	ch, _, err := s.strbuf.ReadRune()
	if err != nil {
		return eos
	}
	return ch
}

// scanIdent consumes the current rune and all contiguous ident runes
func (s *Scanner) scanIdent() (tok Token, lit string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent ident character into the buffer
	// Non-ident characters and EOS will cause the loop to exit
	for {
		if ch := s.read(); ch == eos {
			break
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	// Otherwise return as a regular identifier
	return IDENT, buf.String()
}

// scanWhitespace consumes the current rune and all contiguous whitespace
func (s *Scanner) scanWhitespace() (tok Token, lit string) {
	// Create a buffer and read the current character into it
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer
	// Non-whitespace characters and EOS will cause the loop to exit
	for {
		if ch := s.read(); ch == eos {
			break
		} else if !unicode.IsSpace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return WS, buf.String()
}

// unread places the previously read rune back on the reader
func (s *Scanner) unread() { _ = s.strbuf.UnreadRune() }

// Token represents a lexical token.
type Token int

// eos is end of string rune
var eos = rune(0)

// tilde is tilde rune
var tilde = '~'

// ValidSortOrders provides valid sort order strings
var ValidSortOrders = map[string]string{
	"asc":        "ascending",
	"ascending":  "ascending",
	"desc":       "descending",
	"descending": "descending",
}

// CreateDirectoryService function creates and returns Admin Service object
func CreateDirectoryService(scope ...string) (*admin.Service, error) {
	adminEmail, err := cfg.ReadConfigString("administrator")
	if err != nil {
		return nil, err
	}

	credentialPath, err := cfg.ReadConfigString("credentialpath")
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	ServiceAccountFilePath := filepath.Join(filepath.ToSlash(credentialPath), cfg.CredentialFile)

	jsonCredentials, err := ioutil.ReadFile(ServiceAccountFilePath)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, scope...)
	if err != nil {
		return nil, fmt.Errorf("JWTConfigFromJSON: %v", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("NewService: %v", err)
	}
	return srv, nil
}

// HashPassword creates a password hash
func HashPassword(password string) (string, error) {
	hasher := sha1.New()

	_, err := hasher.Write([]byte(password))
	if err != nil {
		return "", err
	}

	hashedBytes := hasher.Sum(nil)
	hexSha1 := hex.EncodeToString(hashedBytes)

	return hexSha1, nil
}

// isOperator checks to see whether or not rune is an operator symbol
func isOperator(ch rune) bool {
	switch ch {
	case '=':
		return true
	case ':':
		return true
	case '>':
		return true
	case '<':
		return true
	}

	return false
}

// IsValidAttr checks to see whether or not an attribute is valid
func IsValidAttr(attr string, attrMap map[string]string) (string, error) {
	lowerAttr := strings.ToLower(attr)

	validAttr := attrMap[lowerAttr]
	if validAttr == "" {
		err := fmt.Errorf("gmin: error - attribute %v is unrecognized", attr)
		return "", err
	}

	return validAttr, nil
}

// NewOutputAttrParser returns a new instance of OutputAttrParser
func NewOutputAttrParser(b *bytes.Buffer) *OutputAttrParser {
	return &OutputAttrParser{oas: NewOutputAttrScanner(b)}
}

// NewOutputAttrScanner returns a new instance of OutputAttrScanner
func NewOutputAttrScanner(b *bytes.Buffer) *OutputAttrScanner {
	scanr := &Scanner{strbuf: b}
	return &OutputAttrScanner{s: scanr}
}

// NewQueryParser returns a new instance of QueryParser
func NewQueryParser(b *bytes.Buffer) *QueryParser {
	return &QueryParser{qs: NewQueryScanner(b)}
}

// NewQueryScanner returns a new instance of QueryScanner
func NewQueryScanner(b *bytes.Buffer) *QueryScanner {
	scanr := &Scanner{strbuf: b}
	return &QueryScanner{s: scanr}
}

// ParseOutputAttrs validates attributes string and formats it for Get and List calls
func ParseOutputAttrs(attrs string, attrMap map[string]string) (string, error) {
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
	bb := bytes.NewBufferString(query)

	p := NewQueryParser(bb)

	qs, err := p.Parse(qAttrMap)
	if err != nil {
		return "", err
	}

	outputStr := strings.Join(qs.Parts, "")

	return outputStr, nil
}

// SliceContainsStr tells whether strs contains s
func SliceContainsStr(strs []string, s string) bool {
	for _, sComp := range strs {
		if s == sComp {
			return true
		}
	}
	return false
}
