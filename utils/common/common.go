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
	// CLOSEBRACK is )
	CLOSEBRACK
	// COMMA is ,
	COMMA
	// FSLASH is /
	FSLASH
	// OPENBRACK is (
	OPENBRACK
	// TILDE is ~
	TILDE
)

// AttributeStr is a struct to hold list of attribute string parts
type AttributeStr struct {
	Fields []string
}

// Parser represents a parser.
type Parser struct {
	s *StrScanner
}

// scan returns the next token from the underlying scanner
func (p *Parser) scan() (tok Token, lit string) {
	// read the next token from the scanner
	tok, lit = p.s.Scan()

	return tok, lit
}

// scanIgnoreWhitespace scans the next non-whitespace token
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return tok, lit
}

// Parse is the entry point for the parser
func (p *Parser) Parse(attrMap map[string]string) (*AttributeStr, error) {
	attrStr := &AttributeStr{}

	for {
		tok, lit := p.scanIgnoreWhitespace()
		if tok == EOS {
			break
		}

		if tok != IDENT && tok != ASTERISK && tok != OPENBRACK && tok != CLOSEBRACK &&
			tok != COMMA && tok != FSLASH && tok != TILDE {
			return nil, fmt.Errorf("gmin: unexpected character %q found in attribute string", lit)
		}

		if tok == IDENT {
			validAttr := attrMap[lit]
			if validAttr == "" {
				err := fmt.Errorf("gmin: error - attribute %v is unrecognized", lit)
				return nil, err
			}
			lit = validAttr
		}

		if tok == TILDE {
			lit = ","
		}

		attrStr.Fields = append(attrStr.Fields, lit)
	}
	return attrStr, nil
}

// StrScanner represents a lexical scanner
type StrScanner struct {
	strbuf *bytes.Buffer
}

// read reads the next rune from the string
func (s *StrScanner) read() rune {
	ch, _, err := s.strbuf.ReadRune()
	if err != nil {
		return eos
	}
	return ch
}

// Scan returns the next token and literal value
func (s *StrScanner) Scan() (tok Token, lit string) {
	// Read the next rune
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace
	// If we see a letter then consume as an ident
	if unicode.IsSpace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if unicode.IsLetter(ch) {
		s.unread()
		return s.scanIdent()
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

// scanIdent consumes the current rune and all contiguous ident runes
func (s *StrScanner) scanIdent() (tok Token, lit string) {
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
func (s *StrScanner) scanWhitespace() (tok Token, lit string) {
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
func (s *StrScanner) unread() { _ = s.strbuf.UnreadRune() }

// Token represents a lexical token.
type Token int

// eos is end of string rune
var eos = rune(0)

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

// NewParser returns a new instance of Parser
func NewParser(b *bytes.Buffer) *Parser {
	return &Parser{s: NewStrScanner(b)}
}

// NewStrScanner returns a new instance of Scanner
func NewStrScanner(b *bytes.Buffer) *StrScanner {
	return &StrScanner{strbuf: b}
}

// ParseOutputAttrs validates attributes string and formats it for Get and List calls
func ParseOutputAttrs(attrs string, attrMap map[string]string) (string, error) {
	bb := bytes.NewBufferString(attrs)

	p := NewParser(bb)

	as, err := p.Parse(attrMap)
	if err != nil {
		return "", err
	}

	outputStr := strings.Join(as.Fields, "")

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

// ValidateQuery validates query attributes and converts them to correct format
func ValidateQuery(query string, queryMap map[string]string) ([]string, error) {
	var correctQuery string

	convertedQuery := []string{}
	sepQuery := strings.Split(query, "~")

	for _, qPart := range sepQuery {
		colon := false
		trimmedQPart := strings.TrimSpace(qPart)

		splitParts := strings.Split(trimmedQPart, "=")
		if len(splitParts) < 2 {
			splitParts = strings.Split(trimmedQPart, ":")
			colon = true
		}

		lowerPart := strings.ToLower(splitParts[0])

		correctQPart := queryMap[lowerPart]
		if correctQPart == "" {
			err := fmt.Errorf("gmin: error - query attribute %v is unrecognized", splitParts[0])
			return nil, err
		}

		if colon {
			correctQuery = correctQPart + ":" + splitParts[1]
		} else {
			correctQuery = correctQPart + "=" + splitParts[1]
		}

		convertedQuery = append(convertedQuery, correctQuery)
	}

	return convertedQuery, nil
}
