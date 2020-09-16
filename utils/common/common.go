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
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"crypto/sha1"

	cfg "github.com/plusworx/gmin/utils/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	sheet "google.golang.org/api/sheets/v4"
)

const (
	// HashFunction specifies password hash function
	HashFunction string = "SHA-1"
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

// EmptyValues is struct used to extract ForceSendFields from JSON
type EmptyValues struct {
	ForceSendFields []string
}

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
	} else if unicode.IsLetter(ch) || ch == underscore {
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
	case '_':
		return UNDERSCORE, string(ch)
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
		} else if unicode.IsLetter(ch) || ch == underscore {
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
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '.' && ch != underscore {
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
		} else if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != underscore {
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

// underscore is underscore rune
var underscore = '_'

// ValidFileFormats provides valid file format strings
var ValidFileFormats = []string{
	"csv",
	"gsheet",
	"json",
}

// ValidSortOrders provides valid sort order strings
var ValidSortOrders = map[string]string{
	"asc":        "ascending",
	"ascending":  "ascending",
	"desc":       "descending",
	"descending": "descending",
}

// ValidPrimaryShowArgs holds valid primary arguments for the show command
var ValidPrimaryShowArgs = []string{
	"cdev",
	"chromeosdevice",
	"crosdev",
	"crosdevice",
	"group",
	"grp",
	"group-alias",
	"grp-alias",
	"galias",
	"ga",
	"group-member",
	"grp-member",
	"grp-mem",
	"gmember",
	"gmem",
	"mdev",
	"mobdev",
	"mobdevice",
	"mobiledevice",
	"orgunit",
	"ou",
	"schema",
	"sc",
	"ua",
	"ualias",
	"user",
	"user-alias",
}

// CreateDirectoryService function creates and returns Admin Service object
func CreateDirectoryService(scope ...string) (*admin.Service, error) {
	ctx, ts, err := oauthSetup(scope)
	if err != nil {
		return nil, err
	}

	srv, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("gmin: error - New Directory Service: %v", err)
	}
	return srv, nil
}

// CreateSheetService function creates and returns Sheet Service object
func CreateSheetService(scope ...string) (*sheet.Service, error) {
	ctx, ts, err := oauthSetup(scope)
	if err != nil {
		return nil, err
	}

	srv, err := sheet.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("gmin: error - New Sheet Service: %v", err)
	}
	return srv, nil
}

func oauthSetup(scope []string) (context.Context, oauth2.TokenSource, error) {
	adminEmail, err := cfg.ReadConfigString("administrator")
	if err != nil {
		return nil, nil, err
	}

	credentialPath, err := cfg.ReadConfigString("credentialpath")
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	ServiceAccountFilePath := filepath.Join(filepath.ToSlash(credentialPath), cfg.CredentialFile)

	jsonCredentials, err := ioutil.ReadFile(ServiceAccountFilePath)
	if err != nil {
		return nil, nil, err
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, scope...)
	if err != nil {
		return nil, nil, fmt.Errorf("gmin: error - JWTConfigFromJSON: %v", err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)

	return ctx, ts, nil
}

// deDupeStrSlice gets rid of duplicate values in a slice
func deDupeStrSlice(strSlice []string) []string {

	check := make(map[string]int)
	res := make([]string, 0)
	for _, val := range strSlice {
		check[val] = 1
	}

	for s := range check {
		res = append(res, s)
	}

	return res
}

// GminMessage constructs a message for output
func GminMessage(msgTxt string) string {
	return Timestamp() + msgTxt
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

// InputFromStdIn checks to see if there is stdin data and sets up a scanner for it
func InputFromStdIn(inputFile string) (*bufio.Scanner, error) {
	file := os.Stdin
	input, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if input.Mode()&os.ModeNamedPipe == 0 {
		return nil, nil
	}
	if inputFile != "" {
		err = errors.New("gmin: error - cannot provide input file when piping in input")
		return nil, err
	}
	scanner := bufio.NewScanner(os.Stdin)

	return scanner, nil
}

// IsErrRetryable checks to see whether Google API error should allow retry
func IsErrRetryable(e error) bool {
	var retryable bool

	gErr, ok := e.(*googleapi.Error)
	if !ok {
		return false
	}

	body := gErr.Body

	switch {
	case gErr.Code == 403 && (strings.Contains(body, "userRateLimitExceeded") || strings.Contains(body, "quotaExceeded")):
		retryable = true
	case gErr.Code == 403 && strings.Contains(body, "rateLimitExceeded"):
		retryable = true
	case gErr.Code == 429 && strings.Contains(body, "rateLimitExceeded"):
		retryable = true
	}

	return retryable
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

// ParseCustomField parses custom schema names argument
func ParseCustomField(cStr string) []string {
	sArgs := strings.Split(cStr, "~")

	return sArgs
}

// ParseForceSend parses force send fields arguments
func ParseForceSend(fStr string, attrMap map[string]string) ([]string, error) {
	result := []string{}

	fArgs := strings.Split(fStr, "~")
	for _, a := range fArgs {
		s, err := IsValidAttr(a, attrMap)
		if err != nil {
			return nil, err
		}

		result = append(result, strings.Title(strings.TrimSpace(s)))
	}
	return result, nil
}

// ParseInputAttrs parses create and update JSON attribute strings
func ParseInputAttrs(jsonBytes []byte) ([]string, error) {
	m := map[string]interface{}{}
	outStr := []string{}

	err := json.Unmarshal(jsonBytes, &m)
	if err != nil {
		return nil, err
	}
	parseMap(m, &outStr)

	return outStr, nil
}

func parseMap(attrMap map[string]interface{}, outStr *[]string) {
	for key, val := range attrMap {
		if strings.ToLower(key) == "customschemas" {
			*outStr = append(*outStr, key)
			continue
		}
		switch concreteVal := val.(type) {
		case bool:
			*outStr = append(*outStr, key+": "+strconv.FormatBool(concreteVal))
		case map[string]interface{}:
			*outStr = append(*outStr, key)
			parseMap(val.(map[string]interface{}), outStr)
		case []interface{}:
			*outStr = append(*outStr, key)
			parseArray(val.([]interface{}), outStr)
		default:
			parseVal(key+": ", concreteVal, outStr)
		}
	}
}

func parseArray(anArray []interface{}, outStr *[]string) {
	for i, val := range anArray {
		iStr := strconv.Itoa(i)
		switch concreteVal := val.(type) {
		case bool:
			*outStr = append(*outStr, "Index"+iStr+": "+strconv.FormatBool(concreteVal))
		case map[string]interface{}:
			*outStr = append(*outStr, "Index"+iStr)
			parseMap(val.(map[string]interface{}), outStr)
		case []interface{}:
			*outStr = append(*outStr, "Index"+iStr)
			parseArray(val.([]interface{}), outStr)
		default:
			parseVal("Index"+iStr+": ", concreteVal, outStr)
		}
	}
}

func parseVal(idx string, val interface{}, outStr *[]string) {
	switch v := val.(type) {
	case int:
		*outStr = append(*outStr, idx+strconv.Itoa(v))
	case float64:
		*outStr = append(*outStr, idx+fmt.Sprintf("%f", v))
	case string:
		*outStr = append(*outStr, idx+v)
	default:
		*outStr = append(*outStr, idx+"unknown")
	}
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

// ProcessHeader processes header column names
func ProcessHeader(hdr []interface{}) map[int]string {
	hdrMap := make(map[int]string)
	for idx, attr := range hdr {
		strAttr := fmt.Sprintf("%v", attr)
		hdrMap[idx] = strings.ToLower(strAttr)
	}

	return hdrMap
}

// ShowAttrs displays object attributes
func ShowAttrs(attrSlice []string, attrMap map[string]string, filter string) {
	for _, a := range attrSlice {
		s, _ := IsValidAttr(a, attrMap)

		if filter == "" {
			fmt.Println(s)
			continue
		}
		if strings.Contains(strings.ToLower(a), strings.ToLower(filter)) {
			fmt.Println(s)
		}
	}
}

// ShowQueryableAttrs displays user queryable attributes
func ShowQueryableAttrs(filter string, qAttrMap map[string]string) {
	keys := make([]string, 0, len(qAttrMap))
	for k := range qAttrMap {
		keys = append(keys, k)
	}

	vals := make([]string, 0, len(qAttrMap))
	for _, k := range keys {
		vals = append(vals, qAttrMap[k])
	}
	deDupedVals := deDupeStrSlice(vals)
	sort.Strings(deDupedVals)

	for _, v := range deDupedVals {
		if filter == "" {
			fmt.Println(v)
			continue
		}

		if strings.Contains(v, strings.ToLower(filter)) {
			fmt.Println(v)
		}

	}
}

// SliceContainsStr tells whether a slice contains a particular string
func SliceContainsStr(strs []string, s string) bool {
	for _, sComp := range strs {
		if s == sComp {
			return true
		}
	}
	return false
}

// Timestamp gets current formatted time
func Timestamp() string {
	t := time.Now()
	return "[" + t.Format("2006-01-02 15:04:05") + "]"
}

// UniqueStrSlice takes a slice with duplicate values and returns one with unique values
func UniqueStrSlice(inSlice []string) []string {
	outSlice := []string{}
	for _, val := range inSlice {
		ok := SliceContainsStr(outSlice, val)
		if !ok {
			outSlice = append(outSlice, val)
		}
	}
	return outSlice
}

// ValidateHeader validated header column names
func ValidateHeader(hdr map[int]string, attrMap map[string]string) error {
	for idx, hdrAttr := range hdr {
		correctVal, err := IsValidAttr(hdrAttr, attrMap)
		if err != nil {
			return err
		}
		hdr[idx] = correctVal
	}
	return nil
}

// ValidateInputAttrs validates JSON attribute string for create and update calls
func ValidateInputAttrs(attrs []string, attrMap map[string]string) error {
	for _, elem := range attrs {
		if strings.HasPrefix(elem, "Index") {
			continue
		}

		keyVal := strings.Split(elem, ":")
		attrName := keyVal[0]
		s, err := IsValidAttr(attrName, attrMap)
		if err != nil {
			return err
		}

		if s != attrName {
			return fmt.Errorf("gmin: error - %v should be %v in attribute string", attrName, s)
		}
	}
	return nil
}

// ValidateRecoveryPhone validates recovery phone number
func ValidateRecoveryPhone(phoneNo string) error {
	if string(phoneNo[0]) != "+" {
		return fmt.Errorf("gmin: error - recovery phone number %v must start with '+' followed by country code", phoneNo)
	}
	return nil
}
