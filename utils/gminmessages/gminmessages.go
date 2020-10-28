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

	ERR_ADMINEMAILREQUIRED       string = "an email address is required - try again"
	ERR_ATTRNOTRECOGNIZED        string = "%v attribute is not recognized"
	ERR_ATTRSHOULDBE             string = "%v should be %v in attribute string"
	ERR_BATCHCHROMEOSDEVICE      string = "error - %s - ChromeOS device: %s"
	ERR_BATCHGROUP               string = "error - %s - group: %s"
	ERR_BATCHGROUPSETTINGS       string = "error - %s - group settings for group: %s"
	ERR_BATCHMEMBER              string = "error - %s - member: %s - group: %s"
	ERR_BATCHMOBILEDEVICE        string = "error - %s - mobile device: %s"
	ERR_BATCHMISSINGUSERDATA     string = "primaryEmail, givenName, familyName and password must all be provided"
	ERR_BATCHOU                  string = "error - %s - orgunit: %s"
	ERR_BATCHUSER                string = "error - %s - user: %s"
	ERR_CALLTYPENOTRECOGNIZED    string = "%v call type not recognized"
	ERR_CREATEDIRECTORYSERVICE   string = "error - Creating Directory Service: %v"
	ERR_CREATEGRPSETTINGSERVICE  string = "error - Creating Group Setting Service: %v"
	ERR_CREATESHEETSERVICE       string = "error - Creating Sheet Service: %v"
	ERR_EMPTYSTRING              string = "%v cannot be empty string"
	ERR_FILENUMBERREQUIRED       string = "a file number is required - try again"
	ERR_FLAGNOTRECOGNIZED        string = "%v flag is not recognized"
	ERR_INVALIDACTIONTYPE        string = "invalid action type: %v"
	ERR_INVALIDADMINEMAIL        string = "invalid admin email - try again"
	ERR_INVALIDCONFIGPATH        string = "invalid config path - try again"
	ERR_INVALIDCREDPATH          string = "invalid credentials path - try again"
	ERR_INVALIDCUSTID            string = "invalid customer id - try again"
	ERR_INVALIDDELIVERYSETTING   string = "invalid delivery setting: %v"
	ERR_INVALIDDEPROVISIONREASON string = "invalid deprovision reason: %v"
	ERR_INVALIDEMAILADDRESS      string = "invalid email address: %v"
	ERR_INVALIDFILEFORMAT        string = "invalid file format: %v"
	ERR_INVALIDFILENUMBER        string = "file number is invalid - try again"
	ERR_INVALIDJSONATTR          string = "attribute string is not valid JSON"
	ERR_INVALIDJSONFILE          string = "input file is not valid JSON"
	ERR_INVALIDLOGLEVEL          string = "invalid loglevel: %v"
	ERR_INVALIDLOGPATH           string = "invalid log path - try again"
	ERR_INVALIDLOGROTATIONCOUNT  string = "invalid log rotation count - try again"
	ERR_INVALIDLOGROTATIONTIME   string = "invalid log rotation time - try again"
	ERR_INVALIDORDERBY           string = "invalid order by field: %v"
	ERR_INVALIDPAGESARGUMENT     string = "pages argument must be 'all' or a number"
	ERR_INVALIDPROJECTIONTYPE    string = "invalid projection type: %v"
	ERR_INVALIDRECOVERYPHONE     string = "recovery phone number %v must start with '+'"
	ERR_INVALIDROLE              string = "invalid role: %v"
	ERR_INVALIDSCHEMACOMPATTR    string = "invalid schema composite attribute: %v"
	ERR_INVALIDSEARCHTYPE        string = "invalid search type: %v"
	ERR_INVALIDSTRING            string = "invalid string for %v supplied: %v"
	ERR_INVALIDVIEWTYPE          string = "invalid view type: %v"
	ERR_JWTCONFIGFROMJSON        string = "error - JWTConfigFromJSON: %v"
	ERR_MAX2ARGSEXCEEDED         string = "exceeded maximum 2 arguments"
	ERR_MAX3ARGSEXCEEDED         string = "exceeded maximum 3 arguments"
	ERR_MISSINGUSERDATA          string = "firstname, lastname and password must all be provided"
	ERR_MUSTBENUMBER             string = "value entered must be a number - try again"
	ERR_NOCOMPOSITEATTRS         string = "%v does not have any composite attributes"
	ERR_NOCUSTOMFIELDMASK        string = "please provide a custom field mask for custom projection"
	ERR_NODEPROVISIONREASON      string = "must provide a deprovision reason"
	ERR_NODOMAINWITHUSERKEY      string = "must provide a domain in addition to userkey"
	ERR_NOGROUPEMAILADDRESS      string = "group email address must be provided"
	ERR_NOINPUTFILE              string = "must provide inputfile"
	ERR_NOJSONDEVICEID           string = "deviceId must be included in the JSON input string"
	ERR_NOJSONGROUPKEY           string = "groupKey must be included in the JSON input string"
	ERR_NOJSONMEMBERKEY          string = "memberKey must be included in the JSON input string"
	ERR_NOJSONOUKEY              string = "ouKey must be included in the JSON input string"
	ERR_NOJSONUSERKEY            string = "userKey must be included in the JSON input string"
	ERR_NOMEMBEREMAILADDRESS     string = "member email address must be provided"
	ERR_NONAMEOROUPATH           string = "name and parentOrgUnitPath must be provided"
	ERR_NOQUERYABLEATTRS         string = "%v does not have any queryable attributes"
	ERR_NOSHEETDATAFOUND         string = "no data found in sheet %s - range: %s"
	ERR_NOSHEETRANGE             string = "sheet-range must be provided"
	ERR_NOTCOMPOSITEATTR         string = "%v is not a composite attribute"
	ERR_NOTFOUNDINCONFIG         string = "%v not found in config"
	ERR_OBJECTNOTFOUND           string = "%v not found"
	ERR_OBJECTNOTRECOGNIZED      string = " %v is not recognized"
	ERR_PIPEINPUTFILECONFLICT    string = "cannot provide input file when piping in input"
	ERR_PROJECTIONFLAGNOTCUSTOM  string = "--projection must be set to 'custom' in order to use custom field mask"
	ERR_QUERYABLEFLAG1ARG        string = "only one argument is allowed with --queryable flag"
	ERR_QUERYANDCOMPOSITEFLAGS   string = "cannot provide both --composite and --queryable flags"
	ERR_QUERYANDDELETEDFLAGS     string = "cannot provide both --query and --deleted flags"
	ERR_TOOMANYARGSMAX1          string = "too many arguments, %v has maximum of 1"
	ERR_TOOMANYARGSMAX2          string = "too many arguments, %v has maximum of 2"
	ERR_UNEXPECTEDATTRCHAR       string = "unexpected character %v found in attribute string"
	ERR_UNEXPECTEDQUERYCHAR      string = "unexpected character %v found in query string"

	// Infos

	INFO_ADMINIS              string = "admin is %v"
	INFO_ADMINSET             string = "administrator set to: %v"
	INFO_CDEVACTIONPERFORMED  string = "%s successfully performed on ChromeOS device: %s"
	INFO_CDEVMOVEPERFORMED    string = "ChromeOS device: %s moved to: %s"
	INFO_CDEVUPDATED          string = "ChromeOS device updated: %s"
	INFO_CONFIGFILENOTFOUND   string = "Config file not found"
	INFO_CREDENTIALPATHSET    string = "service account credential path set to: %v"
	INFO_CREDENTIALSSET       string = "credentials set using: %v"
	INFO_CUSTOMERIDSET        string = "customer ID set to: %v"
	INFO_ENVVARSNOTFOUND      string = "No environment variables found"
	INFO_GROUPCREATED         string = "group created: %s"
	INFO_GROUPALIASCREATED    string = "group alias: %s created for group: %s"
	INFO_GROUPALIASDELETED    string = "group alias: %s deleted for group: %s"
	INFO_GROUPDELETED         string = "group deleted: %s"
	INFO_GROUPSETTINGSCHANGED string = "group settings changed for group: %s"
	INFO_INITCANCELLED        string = "init command cancelled"
	INFO_INITCOMPLETED        string = "init completed successfully"
	INFO_GROUPUPDATED         string = "group updated: %s"
	INFO_LOGPATHSET           string = "log path set to: %v"
	INFO_LOGROTATIONCOUNTSET  string = "log rotation count set to: %v"
	INFO_LOGROTATIONTIMESET   string = "log rotation time set to: %v"
	INFO_MDEVACTIONPERFORMED  string = "%s successfully performed on mobile device: %s"
	INFO_MDEVDELETED          string = "mobile device deleted: %s"
	INFO_MEMBERCREATED        string = "member: %s created in group: %s"
	INFO_MEMBERDELETED        string = "member: %s deleted from group: %s"
	INFO_MEMBERUPDATED        string = "member: %s updated in group: %s"
	INFO_OUCREATED            string = "orgunit created: %s"
	INFO_OUDELETED            string = "orgunit deleted: %s"
	INFO_OUUPDATED            string = "orgunit updated: %s"
	INFO_SCHEMACREATED        string = "schema created: %s"
	INFO_SCHEMADELETED        string = "schema deleted: %s"
	INFO_SCHEMAUPDATED        string = "schema updated: %s"
	INFO_SETCOMMANDCANCELLED  string = "set command cancelled"
	INFO_USERCREATED          string = "user created: %s"
	INFO_USERALIASCREATED     string = "user alias: %s created for user: %s"
	INFO_USERALIASDELETED     string = "user alias: %s deleted for user: %s"
	INFO_USERDELETED          string = "user deleted: %s"
	INFO_USERUPDATED          string = "user updated: %s"
	INFO_USERUNDELETED        string = "user undeleted: %s"
)
