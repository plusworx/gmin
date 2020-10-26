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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	gset "google.golang.org/api/groupssettings/v1"
	"google.golang.org/api/option"
	sheet "google.golang.org/api/sheets/v4"
)

const (
	// Call Types

	CALLTYPECREATE = iota
	CALLTYPEDELETE
	CALLTYPEMANAGE
	CALLTYPEMOVE
	CALLTYPEUNDELETE
	CALLTYPEUPDATE
)
const (
	// Object Types

	OBJTYPECROSDEV = iota
	OBJTYPEGROUP
	OBJTYPEGRPSET
	OBJTYPEMEMBER
	OBJTYPEMOBDEV
	OBJTYPEORGUNIT
	OBJTYPEUSER
)
const (
	// Service Types

	// SRVTYPEADMIN is used to request admin service
	SRVTYPEADMIN = iota
	// SRVTYPEGRPSETTING is used to request sheet service
	SRVTYPEGRPSETTING
	// SRVTYPESHEET is used to request sheet service
	SRVTYPESHEET
)

const (
	// QUIT is used for terminating commands
	QUIT int = 99
	// TIMEFORMAT is used to format timestamp
	TIMEFORMAT string = "2006-01-02T15:04:05Z0700"
)

// EmptyValues is struct used to extract ForceSendFields from JSON
type EmptyValues struct {
	ForceSendFields []string
}

var globalFlagValues = []string{
	"loglevel",
}

// Logger passed from logging package
var Logger *zap.SugaredLogger

// ValidFileFormats provides valid file format strings
var ValidFileFormats = []string{
	"csv",
	"gsheet",
	"json",
	"text",
	"txt",
}

// validLogLevels provides valid log level strings
var validLogLevels = []string{
	"debug",
	"error",
	"info",
	"warn",
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
	"chromeos-device",
	"cros-dev",
	"cros-device",
	"group",
	"grp",
	"group-alias",
	"group-settings",
	"grp-settings",
	"grp-set",
	"gsettings",
	"gset",
	"grp-alias",
	"galias",
	"ga",
	"group-member",
	"grp-member",
	"grp-mem",
	"gmember",
	"gmem",
	"mdev",
	"mob-dev",
	"mob-device",
	"mobile-device",
	"orgunit",
	"ou",
	"schema",
	"sc",
	"ua",
	"ualias",
	"user",
	"user-alias",
	"usr",
}

// CreateService function creates and returns a service object
func CreateService(serviceType int, scope ...string) (interface{}, error) {
	var srv interface{}

	ctx, ts, err := oauthSetup(scope)
	if err != nil {
		return nil, err
	}

	// Admin service
	if serviceType == SRVTYPEADMIN {
		srv, err = admin.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			err = fmt.Errorf(gmess.ERR_CREATEDIRECTORYSERVICE, err)
			Logger.Error(err)
			return nil, err
		}
	}

	// Group Setting service
	if serviceType == SRVTYPEGRPSETTING {
		srv, err = gset.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			err = fmt.Errorf(gmess.ERR_CREATEGRPSETTINGSERVICE, err)
			Logger.Error(err)
			return nil, err
		}
	}

	// Sheet service
	if serviceType == SRVTYPESHEET {
		srv, err = sheet.NewService(ctx, option.WithTokenSource(ts))
		if err != nil {
			err = fmt.Errorf(gmess.ERR_CREATESHEETSERVICE, err)
			Logger.Error(err)
			return nil, err
		}
	}

	return srv, nil
}

// deDupeStrSlice gets rid of duplicate values in a slice
func deDupeStrSlice(strSlice []string) []string {
	Logger.Debugw("starting deDupeStrSlice()",
		"strSlice", strSlice)
	defer Logger.Debug("finished deDupeStrSlice()")

	check := make(map[string]int)
	res := make([]string, 0)
	for _, val := range strSlice {
		check[val] = 1
	}

	for str := range check {
		res = append(res, str)
	}
	return res
}

// GminMessage constructs a message for output
func GminMessage(msgTxt string) string {
	Logger.Debugw("starting GminMessage()",
		"msgTxt", msgTxt)
	defer Logger.Debug("finished GminMessage()")
	return Timestamp() + " gmin: " + msgTxt
}

// Hostname gets machine hostname if possible
func Hostname() string {
	Logger.Debug("starting Hostname()")
	defer Logger.Debug("finished Hostname()")

	var hName string

	hName, err := os.Hostname()
	if err != nil {
		hName = "unavailable"
	}
	return hName
}

// InputFromStdIn checks to see if there is stdin data and sets up a scanner for it
func InputFromStdIn(inputFile string) (*bufio.Scanner, error) {
	Logger.Debugw("starting InputFromStdIn()",
		"inputFile", inputFile)
	defer Logger.Debug("finished InputFromStdIn()")

	file := os.Stdin
	input, err := file.Stat()
	if err != nil {
		Logger.Error(err)
		return nil, err
	}
	if input.Mode()&os.ModeNamedPipe == 0 {
		return nil, nil
	}
	if inputFile != "" {
		err = errors.New(gmess.ERR_PIPEINPUTFILECONFLICT)
		Logger.Error(err)
		return nil, err
	}
	scanner := bufio.NewScanner(os.Stdin)

	return scanner, nil
}

// IPAddress gets IP address of machine if possible
func IPAddress() string {
	Logger.Debug("starting IPAddress()")
	defer Logger.Debug("finished IPAddress()")

	var ip string

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		ip = "unavailable"
	}
	defer conn.Close()

	ip = conn.LocalAddr().String()

	return ip
}

// IsErrRetryable checks to see whether Google API error should allow retry
func IsErrRetryable(e error) bool {
	Logger.Debugw("starting IsErrRetryable()",
		"e", e)
	defer Logger.Debug("finished IsErrRetryable()")

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

// IsValidAttr checks to see whether or not an attribute is valid
func IsValidAttr(attr string, attrMap map[string]string) (string, error) {
	Logger.Debugw("starting IsValidAttr()",
		"attr", attr,
		"attrMap", attrMap)
	defer Logger.Debug("finished IsValidAttr()")

	lowerAttr := strings.ToLower(attr)

	validAttr := attrMap[lowerAttr]
	if validAttr == "" {
		err := fmt.Errorf(gmess.ERR_ATTRNOTRECOGNIZED, attr)
		Logger.Error(err)
		return "", err
	}

	return validAttr, nil
}

func oauthSetup(scope []string) (context.Context, oauth2.TokenSource, error) {
	Logger.Debugw("starting oauthSetup()",
		"scope", scope)
	defer Logger.Debug("finished oauthSetup()")

	adminEmail, err := cfg.ReadConfigString(cfg.CONFIGADMIN)
	if err != nil {
		return nil, nil, err
	}

	credentialPath, err := cfg.ReadConfigString(cfg.CONFIGCREDPATH)
	if err != nil {
		return nil, nil, err
	}

	ctx := context.Background()

	ServiceAccountFilePath := filepath.Join(filepath.ToSlash(credentialPath), cfg.CREDENTIALFILE)

	jsonCredentials, err := ioutil.ReadFile(ServiceAccountFilePath)
	if err != nil {
		Logger.Error(err)
		return nil, nil, err
	}

	config, err := google.JWTConfigFromJSON(jsonCredentials, scope...)
	if err != nil {
		Logger.Error(err)
		return nil, nil, fmt.Errorf(gmess.ERR_JWTCONFIGFROMJSON, err)
	}
	config.Subject = adminEmail

	ts := config.TokenSource(ctx)

	return ctx, ts, nil
}

// ParseForceSend parses force send fields arguments
func ParseForceSend(fStr string, attrMap map[string]string) ([]string, error) {
	Logger.Debugw("starting ParseForceSend()",
		"fStr", fStr,
		"attrMap", attrMap)
	defer Logger.Debug("finished ParseForceSend()")

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
	Logger.Debug("starting ParseInputAttrs()")
	defer Logger.Debug("finished ParseInputAttrs()")

	m := map[string]interface{}{}
	outStr := []string{}

	err := json.Unmarshal(jsonBytes, &m)
	if err != nil {
		Logger.Error(err)
		return nil, err
	}
	parseMap(m, &outStr)

	return outStr, nil
}

func parseMap(attrMap map[string]interface{}, outStr *[]string) {
	Logger.Debugw("starting parseMap()",
		"attrMap", attrMap)
	defer Logger.Debug("finished parseMap()")

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
	Logger.Debugw("starting parseArray()",
		"anArray", anArray)
	defer Logger.Debug("finished parseArray()")

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
	Logger.Debugw("starting parseVal()",
		"idx", idx,
		"val", val)
	defer Logger.Debug("finished parseVal()")

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

// ProcessHeader processes header column names
func ProcessHeader(hdr []interface{}) map[int]string {
	Logger.Debugw("starting ProcessHeader()",
		"hdr", hdr)
	defer Logger.Debug("finished ProcessHeader()")

	hdrMap := make(map[int]string)
	for idx, attr := range hdr {
		strAttr := fmt.Sprintf("%v", attr)
		hdrMap[idx] = strings.ToLower(strAttr)
	}
	return hdrMap
}

// ShowAttrs displays object attributes
func ShowAttrs(attrSlice []string, attrMap map[string]string, filter string) {
	Logger.Debugw("starting ShowAttrs()",
		"attrSlice", attrSlice,
		"attrMap", attrMap,
		"filter", filter)
	defer Logger.Debug("finished ShowAttrs()")

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

// ShowAttrVals displays object attribute enumerated values or names of attributes that have them
func ShowAttrVals(attrSlice []string, filter string) {
	Logger.Debugw("starting ShowAttrVals()",
		"attrSlice", attrSlice,
		"filter", filter)
	defer Logger.Debug("finished ShowAttrVals()")

	for _, a := range attrSlice {
		if filter == "" {
			fmt.Println(a)
			continue
		}
		if strings.Contains(strings.ToLower(a), strings.ToLower(filter)) {
			fmt.Println(a)
		}
	}
}

// ShowFlagValues displays enumerated flag values
func ShowFlagValues(flagSlice []string, filter string) {
	Logger.Debugw("starting ShowFlagValues()",
		"flagSlice", flagSlice,
		"filter", filter)
	defer Logger.Debug("finished ShowFlagValues()")

	for _, value := range flagSlice {
		if filter == "" {
			fmt.Println(value)
			continue
		}
		ok := strings.Contains(strings.ToLower(value), strings.ToLower(filter))
		if ok {
			fmt.Println(value)
		}
	}
}

// ShowGlobalFlagValues displays enumerated global flag values
func ShowGlobalFlagValues(lenArgs int, args []string, filter string) error {
	Logger.Debugw("starting ShowGlobalFlagValues()",
		"lenArgs", lenArgs,
		"args", args,
		"filter", filter)
	defer Logger.Debug("finished ShowGlobalFlagValues()")

	if lenArgs == 1 {
		ShowFlagValues(globalFlagValues, filter)
	}

	if lenArgs == 2 {
		flag := strings.ToLower(args[1])

		switch {
		case flag == flgnm.FLG_LOGLEVEL:
			ShowFlagValues(validLogLevels, filter)
		default:
			err := fmt.Errorf(gmess.ERR_FLAGNOTRECOGNIZED, args[1])
			Logger.Error(err)
			return err
		}
	}
	return nil
}

// ShowQueryableAttrs displays user queryable attributes
func ShowQueryableAttrs(filter string, qAttrMap map[string]string) {
	Logger.Debugw("starting ShowQueryableAttrs()",
		"filter", filter,
		"qAttrMap", qAttrMap)
	defer Logger.Debug("finished ShowQueryableAttrs()")

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
	Logger.Debugw("starting SliceContainsStr()",
		"strs", strs,
		"s", s)
	defer Logger.Debug("finished SliceContainsStr()")

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
	return "[" + t.Format(TIMEFORMAT) + "]"
}

// UniqueStrSlice takes a slice with duplicate values and returns one with unique values
func UniqueStrSlice(inSlice []string) []string {
	Logger.Debugw("starting UniqueStrSlice()",
		"inSlice", inSlice)
	defer Logger.Debug("finished UniqueStrSlice()")

	outSlice := []string{}
	for _, val := range inSlice {
		ok := SliceContainsStr(outSlice, val)
		if !ok {
			outSlice = append(outSlice, val)
		}
	}
	return outSlice
}

// Username gets username of current user if possible
func Username() string {
	Logger.Debug("starting Username()")
	defer Logger.Debug("finished Username()")

	var (
		uName       string
		currentUser *user.User
	)

	currentUser, err := user.Current()
	if err != nil {
		uName = "unavailable"
	}
	uName = currentUser.Username
	return uName
}

// ValidateHeader validated header column names
func ValidateHeader(hdr map[int]string, attrMap map[string]string) error {
	Logger.Debugw("starting ValidateHeader()",
		"hdr", hdr,
		"attrMap", attrMap)
	defer Logger.Debug("finished ValidateHeader()")

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
	Logger.Debugw("starting ValidateInputAttrs()",
		"attrs", attrs,
		"attrMap", attrMap)
	defer Logger.Debug("finished ValidateInputAttrs()")

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
			err = fmt.Errorf(gmess.ERR_ATTRSHOULDBE, attrName, s)
			Logger.Error(err)
			return err
		}
	}
	return nil
}

// ValidateRecoveryPhone validates recovery phone number
func ValidateRecoveryPhone(phoneNo string) error {
	Logger.Debugw("starting ValidateRecoveryPhone()",
		"phoneNo", phoneNo)
	defer Logger.Debug("finished ValidateRecoveryPhone()")

	if string(phoneNo[0]) != "+" {
		err := fmt.Errorf(gmess.ERR_INVALIDRECOVERYPHONE, phoneNo)
		Logger.Error(err)
		return err
	}
	return nil
}
