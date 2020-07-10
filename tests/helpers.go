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

package tests

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

// AboutCompActualExpected compares actual test results in admin.UserAbout object with expected results
func AboutCompActualExpected(about *admin.UserAbout, expected map[string]string) error {
	aboutVal := reflect.ValueOf(about)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(aboutVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// AddressCompActualExpected compares actual test results in admin.UserAddress object with expected results
func AddressCompActualExpected(address *admin.UserAddress, expected map[string]string) error {
	var fieldStr string

	addrVal := reflect.ValueOf(address)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(addrVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// DummyDirectoryService function creates and returns dummy Admin Service object
func DummyDirectoryService(scope ...string) (*admin.Service, error) {
	adminEmail := "admin@mycompany.org"
	credentials := `{
	"type": "service_account",
	"project_id": "test",
	"private_key_id": "1234",
	"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkuRIIQqbhw64ORg\naYlc7iqk8tvt+ozuS+ibVsk=\n-----END PRIVATE KEY-----\n",
	"client_email": "account@gserviceaccount.com",
	"client_id": "1234",
	"auth_uri": "https://accounts.google.com/o/oauth2/auth",
	"token_uri": "https://oauth2.googleapis.com/token",
	"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/service-account%40gserviceaccount.com"
  }`

	ctx := context.Background()

	jsonCredentials := []byte(credentials)

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

// EmailCompActualExpected compares actual test results in admin.UserEmail object with expected results
func EmailCompActualExpected(email *admin.UserEmail, expected map[string]string) error {
	var fieldStr string

	emailVal := reflect.ValueOf(email)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(emailVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// ExtIDCompActualExpected compares actual test results in admin.UserExternalId object with expected results
func ExtIDCompActualExpected(extID *admin.UserExternalId, expected map[string]string) error {
	extIDVal := reflect.ValueOf(extID)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(extIDVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// GenderCompActualExpected compares actual test results in admin.UserGender object with expected results
func GenderCompActualExpected(gender *admin.UserGender, expected map[string]string) error {
	genVal := reflect.ValueOf(gender)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(genVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// ImCompActualExpected compares actual test results in admin.UserIm object with expected results
func ImCompActualExpected(im *admin.UserIm, expected map[string]string) error {
	var fieldStr string

	imVal := reflect.ValueOf(im)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(imVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// KeywordCompActualExpected compares actual test results in admin.UserKeyword object with expected results
func KeywordCompActualExpected(keyword *admin.UserKeyword, expected map[string]string) error {
	keyVal := reflect.ValueOf(keyword)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(keyVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// LanguageCompActualExpected compares actual test results in admin.UserLanguage object with expected results
func LanguageCompActualExpected(lang *admin.UserLanguage, expected map[string]string) error {
	langVal := reflect.ValueOf(lang)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(langVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// LocationCompActualExpected compares actual test results in admin.UserLocation object with expected results
func LocationCompActualExpected(loc *admin.UserLocation, expected map[string]string) error {
	locVal := reflect.ValueOf(loc)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(locVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// OrgCompActualExpected compares actual test results in admin.UserOrganization object with expected results
func OrgCompActualExpected(org *admin.UserOrganization, expected map[string]string) error {
	var fieldStr string

	orgVal := reflect.ValueOf(org)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(orgVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else if key == "FullTimeEquivalent" {
			fieldInt := fieldVal.Int()
			fieldStr = strconv.FormatInt(fieldInt, 10)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// PhoneCompActualExpected compares actual test results in admin.UserPhone object with expected results
func PhoneCompActualExpected(phone *admin.UserPhone, expected map[string]string) error {
	var fieldStr string

	phoneVal := reflect.ValueOf(phone)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(phoneVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// PosixCompActualExpected compares actual test results in admin.UserPosixAccount object with expected results
func PosixCompActualExpected(posix *admin.UserPosixAccount, expected map[string]string) error {
	var fieldStr string

	posixVal := reflect.ValueOf(posix)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(posixVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else if key == "Gid" || key == "Uid" {
			fieldInt := fieldVal.Uint()
			fieldStr = strconv.FormatUint(fieldInt, 10)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// RelationCompActualExpected compares actual test results in admin.UserRelation object with expected results
func RelationCompActualExpected(relation *admin.UserRelation, expected map[string]string) error {
	relVal := reflect.ValueOf(relation)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(relVal).FieldByName(key)
		fieldStr := fieldVal.String()

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// SSHKeyCompActualExpected compares actual test results in admin.UserSshPublicKey object with expected results
func SSHKeyCompActualExpected(sshkey *admin.UserSshPublicKey, expected map[string]string) error {
	var fieldStr string

	keyVal := reflect.ValueOf(sshkey)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(keyVal).FieldByName(key)

		if key == "ExpirationTimeUsec" {
			fieldInt := fieldVal.Int()
			fieldStr = strconv.FormatInt(fieldInt, 10)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// UserCompActualExpected compares actual test results in admin.User object with expected results
func UserCompActualExpected(user *admin.User, expected map[string]string) error {
	var fieldStr string

	userVal := reflect.ValueOf(user)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(userVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}

// WebsiteCompActualExpected compares actual test results in admin.UserWebsite object with expected results
func WebsiteCompActualExpected(website *admin.UserWebsite, expected map[string]string) error {
	var fieldStr string

	webVal := reflect.ValueOf(website)

	for key, expVal := range expected {
		fieldVal := reflect.Indirect(webVal).FieldByName(key)

		if expVal == "true" || expVal == "false" {
			fieldBool := fieldVal.Bool()
			fieldStr = strconv.FormatBool(fieldBool)
		} else {
			fieldStr = fieldVal.String()
		}

		if fieldStr != expVal {
			return fmt.Errorf("Expected %v: %v; Got: %v", key, expVal, fieldStr)
		}
	}

	return nil
}
