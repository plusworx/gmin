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
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"strings"

	"crypto/sha1"

	cfg "github.com/plusworx/gmin/utils/config"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const (
	// AttrStr is attribute literal
	AttrStr string = "attribute"
	// HashFunction specifies password hash function
	HashFunction string = "SHA-1"
	// RepeatTxt is used to signal a repeated attribute in attribute argument processing
	RepeatTxt string = "**** repeat ****"
	// RoleStr is role literal
	RoleStr string = "role"
)

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

	var ServiceAccountFilePath = cfg.CredentialsFile

	ctx := context.Background()

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

// SliceContainsStr tells whether strs contains s
func SliceContainsStr(strs []string, s string) bool {
	for _, sComp := range strs {
		if s == sComp {
			return true
		}
	}
	return false
}

// ValidateArgs validates attributes and converts them to correct format
func ValidateArgs(args string, attrMap map[string]string, argType string) ([]string, error) {
	convertedArgs := []string{}
	lowerArgs := strings.ToLower(args)
	sepArgs := strings.Split(lowerArgs, "~")

	for _, arg := range sepArgs {
		trimmedArg := strings.TrimSpace(arg)
		correctArg := attrMap[trimmedArg]
		if correctArg == "" {
			err := fmt.Errorf("gmin: error - %v %v is unrecognized", argType, arg)
			return nil, err
		}
		convertedArgs = append(convertedArgs, correctArg)
	}

	return convertedArgs, nil
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
