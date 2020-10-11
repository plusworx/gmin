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

package gminmessages

const (
	// Errors

	ErrAdminEmailRequired       string = "an email address is required - try again"
	ErrAttrNotRecognized        string = "%v attribute is not recognized"
	ErrBatchChromeOSDevice      string = "error - %s - ChromeOS device: %s"
	ErrBatchGroup               string = "error - %s - group: %s"
	ErrBatchGroupSettings       string = "error - %s - group settings for group: %s"
	ErrBatchMember              string = "error - %s - member: %s - group: %s"
	ErrBatchMobileDevice        string = "error - %s - mobile device: %s"
	ErrBatchMissingUserData     string = "primaryEmail, givenName, familyName and password must all be provided"
	ErrBatchOU                  string = "error - %s - orgunit: %s"
	ErrBatchUser                string = "error - %s - user: %s"
	ErrCreateDirectoryService   string = "error - Creating Directory Service: %v"
	ErrCreateGrpSettingService  string = "error - Creating Group Setting Service: %v"
	ErrCreateSheetService       string = "error - Creating Sheet Service: %v"
	ErrEmptyString              string = "%v cannot be empty string"
	ErrFileNumberRequired       string = "a file number is required - try again"
	ErrFlagNotRecognized        string = "%v flag is not recognized"
	ErrInvalidActionType        string = "invalid action type: %v"
	ErrInvalidAdminEmail        string = "invalid admin email - try again"
	ErrInvalidConfigPath        string = "invalid config path - try again"
	ErrInvalidCredPath          string = "invalid credentials path - try again"
	ErrInvalidCustID            string = "invalid customer id - try again"
	ErrInvalidDeliverySetting   string = "invalid delivery setting: %v"
	ErrInvalidDeprovisionReason string = "invalid deprovision reason: %v"
	ErrInvalidEmailAddress      string = "invalid email address: %v"
	ErrInvalidFileFormat        string = "invalid file format: %v"
	ErrInvalidFileNumber        string = "file number is invalid - try again"
	ErrInvalidJSONAttr          string = "attribute string is not valid JSON"
	ErrInvalidJSONFile          string = "input file is not valid JSON"
	ErrInvalidLogLevel          string = "invalid loglevel: %v"
	ErrInvalidLogPath           string = "invalid log path - try again"
	ErrInvalidOrderBy           string = "invalid order by field: %v"
	ErrInvalidPagesArgument     string = "pages argument must be 'all' or a number"
	ErrInvalidProjectionType    string = "invalid projection type: %v"
	ErrInvalidRecoveryPhone     string = "recovery phone number %v must start with '+'"
	ErrInvalidRole              string = "invalid role: %v"
	ErrInvalidSchemaCompAttr    string = "invalid schema composite attribute: %v"
	ErrInvalidSearchType        string = "invalid search type: %v"
	ErrInvalidString            string = "invalid string for %v supplied: %v"
	ErrInvalidViewType          string = "invalid view type: %v"
	ErrMax1ArgExceeded          string = "exceeded maximum 1 arguments"
	ErrMax2ArgsExceeded         string = "exceeded maximum 2 arguments"
	ErrMax3ArgsExceeded         string = "exceeded maximum 3 arguments"
	ErrMissingUserData          string = "firstname, lastname and password must all be provided"
	ErrNoCompositeAttrs         string = "%v does not have any composite attributes"
	ErrNoCustomFieldMask        string = "please provide a custom field mask for custom projection"
	ErrNoDeprovisionReason      string = "must provide a deprovision reason"
	ErrNoDomainWithUserKey      string = "must provide a domain in addition to userkey"
	ErrNoGroupEmailAddress      string = "group email address must be provided"
	ErrNoInputFile              string = "must provide inputfile"
	ErrNoJSONDeviceID           string = "deviceId must be included in the JSON input string"
	ErrNoJSONGroupKey           string = "groupKey must be included in the JSON input string"
	ErrNoJSONMemberKey          string = "memberKey must be included in the JSON input string"
	ErrNoJSONOUKey              string = "ouKey must be included in the JSON input string"
	ErrNoJSONUserKey            string = "userKey must be included in the JSON input string"
	ErrNoMemberEmailAddress     string = "member email address must be provided"
	ErrNoNameOrOuPath           string = "name and parentOrgUnitPath must be provided"
	ErrNoQueryableAttrs         string = "%v does not have any queryable attributes"
	ErrNoSheetDataFound         string = "no data found in sheet %s - range: %s"
	ErrNoSheetRange             string = "sheet-range must be provided"
	ErrNotCompositeAttr         string = "%v is not a composite attribute"
	ErrObjectNotFound           string = "%v not found"
	ErrObjectNotRecognized      string = " %v is not recognized"
	ErrPipeInputFileConflict    string = "cannot provide input file when piping in input"
	ErrProjectionFlagNotCustom  string = "--projection must be set to 'custom' in order to use custom field mask"
	ErrQueryableFlag1Arg        string = "only one argument is allowed with --queryable flag"
	ErrQueryAndCompositeFlags   string = "cannot provide both --composite and --queryable flags"
	ErrQueryAndDeletedFlags     string = "cannot provide both --query and --deleted flags"
	ErrTooManyArgsMax1          string = "too many arguments, %v has maximum of 1"
	ErrTooManyArgsMax2          string = "too many arguments, %v has maximum of 2"
	ErrUnexpectedAttrChar       string = "unexpected character %v found in attribute string"

	// Infos

	InfoAdminSet             string = "administrator set to: %v"
	InfoCDevActionPerformed  string = "%s successfully performed on ChromeOS device: %s"
	InfoCDevMovePerformed    string = "ChromeOS device: %s moved to: %s"
	InfoCDevUpdated          string = "ChromeOS device updated: %s"
	InfoConfigFileNotFound   string = "Config file not found"
	InfoCredentialPathSet    string = "service account credential path set to: %v"
	InfoCredentialsSet       string = "credentials set using: %v"
	InfoCustomerIDSet        string = "customer ID set to: %v"
	InfoEnvVarsNotFound      string = "No environment variables found"
	InfoGroupCreated         string = "group created: %s"
	InfoGroupAliasCreated    string = "group alias: %s created for group: %s"
	InfoGroupAliasDeleted    string = "group alias: %s deleted for group: %s"
	InfoGroupDeleted         string = "group deleted: %s"
	InfoGroupSettingsChanged string = "group settings changed for group: %s"
	InfoInitCancelled        string = "init command cancelled"
	InfoInitCompleted        string = "init completed successfully"
	InfoGroupUpdated         string = "group updated: %s"
	InfoLogPathSet           string = "log path set to: %v"
	InfoMDevActionPerformed  string = "%s successfully performed on mobile device: %s"
	InfoMDevDeleted          string = "mobile device deleted: %s"
	InfoMemberCreated        string = "member: %s created in group: %s"
	InfoMemberDeleted        string = "member: %s deleted from group: %s"
	InfoMemberUpdated        string = "member: %s updated in group: %s"
	InfoOUCreated            string = "orgunit created: %s"
	InfoOUDeleted            string = "orgunit deleted: %s"
	InfoOUUpdated            string = "orgunit updated: %s"
	InfoSchemaCreated        string = "schema created: %s"
	InfoSchemaDeleted        string = "schema deleted: %s"
	InfoSchemaUpdated        string = "schema updated: %s"
	InfoSetCommandCancelled  string = "set command cancelled"
	InfoUserCreated          string = "user created: %s"
	InfoUserAliasCreated     string = "user alias: %s created for user: %s"
	InfoUserAliasDeleted     string = "user alias: %s deleted for user: %s"
	InfoUserDeleted          string = "user deleted: %s"
	InfoUserUpdated          string = "user updated: %s"
	InfoUserUndeleted        string = "user undeleted: %s"
)
