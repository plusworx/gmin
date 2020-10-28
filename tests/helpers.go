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
	"log"
	"reflect"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

var cfgFile string

// TestGroupAttrMap is Group attribute map for testing purposes
var TestGroupAttrMap = map[string]string{
	"admincreated":       "adminCreated",
	"description":        "description",
	"directmemberscount": "directMembersCount",
	"email":              "email",
	"etag":               "etag",
	"id":                 "id",
	"kind":               "kind",
	"name":               "name",
	"noneditablealiases": "nonEditableAliases",
}

// TestUserAttrMap is User attribute map for testing purposes
var TestUserAttrMap = map[string]string{
	"accountid":                  "accountId",
	"address":                    "address",
	"addresses":                  "addresses",
	"addressmeas":                "addressMeAs",
	"agreedtoterms":              "agreedToTerms",
	"aliases":                    "aliases",
	"archived":                   "archived",
	"area":                       "area",
	"buildingid":                 "buildingId",
	"changepasswordatnextlogin":  "changePasswordAtNextLogin",
	"christianname":              "givenName",
	"contenttype":                "contentType",
	"costcenter":                 "costCenter",
	"country":                    "country",
	"countrycode":                "countryCode",
	"creationtime":               "creationTime",
	"customerid":                 "customerId",
	"customgender":               "customGender",
	"customlanguage":             "customLanguage",
	"customprotocol":             "customProtocol",
	"customschemas":              "customSchemas",
	"customtype":                 "customType",
	"deletiontime":               "deletionTime",
	"department":                 "department",
	"description":                "description",
	"deskcode":                   "deskCode",
	"domain":                     "domain",
	"emails":                     "emails",
	"etag":                       "etag",
	"expirationtimeusec":         "expirationTimeUsec",
	"externalids":                "externalIds",
	"extendedaddress":            "extendedAddress",
	"familyname":                 "familyName",
	"fingerprint":                "fingerprint",
	"firstname":                  "givenName",
	"floorname":                  "floorName",
	"floorsection":               "floorSection",
	"formatted":                  "formatted",
	"fullname":                   "fullName",
	"fulltimeequivalent":         "fullTimeEquivalent",
	"gecos":                      "gecos",
	"gender":                     "gender",
	"gid":                        "gid",
	"givenname":                  "givenName",
	"hashfunction":               "hashFunction",
	"homedirectory":              "homeDirectory",
	"id":                         "id",
	"im":                         "im",
	"ims":                        "ims",
	"includeinglobaladdresslist": "includeInGlobalAddressList",
	"ipwhitelisted":              "ipWhiteListed",
	"isadmin":                    "isAdmin",
	"isdelegatedadmin":           "isDelegatedAdmin",
	"isenforcedin2sv":            "isEnforcedIn2Sv",
	"isenrolledin2sv":            "isEnrolledIn2Sv",
	"ismailboxsetup":             "isMailboxSetup",
	"key":                        "key",
	"keywords":                   "keywords",
	"kind":                       "kind",
	"languagecode":               "languageCode",
	"languages":                  "languages",
	"lastlogintime":              "lastLoginTime",
	"lastname":                   "familyName",
	"locality":                   "locality",
	"location":                   "location",
	"locations":                  "locations",
	"name":                       "name",
	"noneditablealiases":         "nonEditableAliases",
	"notes":                      "notes",
	"operatingsystemtype":        "operatingSystemType",
	"organisations":              "organizations",
	"organizations":              "organizations",
	"orgunitpath":                "orgUnitPath",
	"password":                   "password",
	"phones":                     "phones",
	"pobox":                      "poBox",
	"posixaccounts":              "posixAccounts",
	"postalcode":                 "postalCode",
	"primary":                    "primary",
	"primaryemail":               "primaryEmail",
	"protocol":                   "protocol",
	"recoveryemail":              "recoveryEmail",
	"recoveryphone":              "recoveryPhone",
	"region":                     "region",
	"relations":                  "relations",
	"shell":                      "shell",
	"sshpublickeys":              "sshPublicKeys",
	"streetaddress":              "streetAddress",
	"surname":                    "familyName",
	"suspended":                  "suspended",
	"suspensionreason":           "suspensionReason",
	"symbol":                     "symbol",
	"systemid":                   "systemId",
	"thumbnailphotoetag":         "thumbnailPhotoEtag",
	"thumbnailphotourl":          "thumbnailPhotoUrl",
	"title":                      "title",
	"type":                       "type",
	"uid":                        "uid",
	"username":                   "username",
	"value":                      "value",
	"websites":                   "websites",
}

// TestUserQueryAttrMap provides lowercase mappings to valid admin.User query attributes
var TestUserQueryAttrMap = map[string]string{
	"address":           "address",
	"addresspobox":      "addressPoBox",
	"addressextended":   "addressExtended",
	"addressstreet":     "addressStreet",
	"addresslocality":   "addressLocality",
	"addressregion":     "addressRegion",
	"addresspostalcode": "addressPostalCode",
	"addresscountry":    "addressCountry",
	"christianname":     "givenName",
	"directmanager":     "directManager",
	"directmanagerid":   "directManagerId",
	"email":             "email",
	"externalid":        "externalId",
	"familyname":        "familyName",
	"firstname":         "givenName",
	"givenname":         "givenName",
	"im":                "im",
	"isadmin":           "isAdmin",
	"isdelegatedadmin":  "isDelegatedAdmin",
	"isenrolledin2sv":   "isEnrolledIn2Sv",
	"isenforcedin2sv":   "isEnforcedIn2Sv",
	"issuspended":       "isSuspended",
	"lastname":          "familyName",
	"manager":           "manager",
	"managerid":         "managerId",
	"name":              "name",
	"orgcostcenter":     "orgCostCenter",
	"orgdepartment":     "orgDepartment",
	"orgdescription":    "orgDescription",
	"orgname":           "orgName",
	"orgtitle":          "orgTitle",
	"orgunitpath":       "orgUnitPath",
	"phone":             "phone",
	"surname":           "familyName",
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

// GetLogger create logger
func GetLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()
	return sugar
}

// InitConfig initialises config
func InitConfig() {
	viper.SetEnvPrefix("GMIN")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".gmin")
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
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
