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

	ERRADMINEMAILREQUIRED       string = "an email address is required - try again"
	ERRATTRNOTRECOGNIZED        string = "%v attribute is not recognized"
	ERRBATCHCHROMEOSDEVICE      string = "error - %s - ChromeOS device: %s"
	ERRBATCHGROUP               string = "error - %s - group: %s"
	ERRBATCHGROUPSETTINGS       string = "error - %s - group settings for group: %s"
	ERRBATCHMEMBER              string = "error - %s - member: %s - group: %s"
	ERRBATCHMOBILEDEVICE        string = "error - %s - mobile device: %s"
	ERRBATCHMISSINGUSERDATA     string = "primaryEmail, givenName, familyName and password must all be provided"
	ERRBATCHOU                  string = "error - %s - orgunit: %s"
	ERRBATCHUSER                string = "error - %s - user: %s"
	ERRCREATEDIRECTORYSERVICE   string = "error - Creating Directory Service: %v"
	ERRCREATEGRPSETTINGSERVICE  string = "error - Creating Group Setting Service: %v"
	ERRCREATESHEETSERVICE       string = "error - Creating Sheet Service: %v"
	ERREMPTYSTRING              string = "%v cannot be empty string"
	ERRFILENUMBERREQUIRED       string = "a file number is required - try again"
	ERRFLAGNOTRECOGNIZED        string = "%v flag is not recognized"
	ERRINVALIDACTIONTYPE        string = "invalid action type: %v"
	ERRINVALIDADMINEMAIL        string = "invalid admin email - try again"
	ERRINVALIDCONFIGPATH        string = "invalid config path - try again"
	ERRINVALIDCREDPATH          string = "invalid credentials path - try again"
	ERRINVALIDCUSTID            string = "invalid customer id - try again"
	ERRINVALIDDELIVERYSETTING   string = "invalid delivery setting: %v"
	ERRINVALIDDEPROVISIONREASON string = "invalid deprovision reason: %v"
	ERRINVALIDEMAILADDRESS      string = "invalid email address: %v"
	ERRINVALIDFILEFORMAT        string = "invalid file format: %v"
	ERRINVALIDFILENUMBER        string = "file number is invalid - try again"
	ERRINVALIDJSONATTR          string = "attribute string is not valid JSON"
	ERRINVALIDJSONFILE          string = "input file is not valid JSON"
	ERRINVALIDLOGLEVEL          string = "invalid loglevel: %v"
	ERRINVALIDLOGPATH           string = "invalid log path - try again"
	ERRINVALIDORDERBY           string = "invalid order by field: %v"
	ERRINVALIDPAGESARGUMENT     string = "pages argument must be 'all' or a number"
	ERRINVALIDPROJECTIONTYPE    string = "invalid projection type: %v"
	ERRINVALIDRECOVERYPHONE     string = "recovery phone number %v must start with '+'"
	ERRINVALIDROLE              string = "invalid role: %v"
	ERRINVALIDSCHEMACOMPATTR    string = "invalid schema composite attribute: %v"
	ERRINVALIDSEARCHTYPE        string = "invalid search type: %v"
	ERRINVALIDSTRING            string = "invalid string for %v supplied: %v"
	ERRINVALIDVIEWTYPE          string = "invalid view type: %v"
	ERRMAX2ARGSEXCEEDED         string = "exceeded maximum 2 arguments"
	ERRMAX3ARGSEXCEEDED         string = "exceeded maximum 3 arguments"
	ERRMISSINGUSERDATA          string = "firstname, lastname and password must all be provided"
	ERRNOCOMPOSITEATTRS         string = "%v does not have any composite attributes"
	ERRNOCUSTOMFIELDMASK        string = "please provide a custom field mask for custom projection"
	ERRNODEPROVISIONREASON      string = "must provide a deprovision reason"
	ERRNODOMAINWITHUSERKEY      string = "must provide a domain in addition to userkey"
	ERRNOGROUPEMAILADDRESS      string = "group email address must be provided"
	ERRNOINPUTFILE              string = "must provide inputfile"
	ERRNOJSONDEVICEID           string = "deviceId must be included in the JSON input string"
	ERRNOJSONGROUPKEY           string = "groupKey must be included in the JSON input string"
	ERRNOJSONMEMBERKEY          string = "memberKey must be included in the JSON input string"
	ERRNOJSONOUKEY              string = "ouKey must be included in the JSON input string"
	ERRNOJSONUSERKEY            string = "userKey must be included in the JSON input string"
	ERRNOMEMBEREMAILADDRESS     string = "member email address must be provided"
	ERRNONAMEOROUPATH           string = "name and parentOrgUnitPath must be provided"
	ERRNOQUERYABLEATTRS         string = "%v does not have any queryable attributes"
	ERRNOSHEETDATAFOUND         string = "no data found in sheet %s - range: %s"
	ERRNOSHEETRANGE             string = "sheet-range must be provided"
	ERRNOTCOMPOSITEATTR         string = "%v is not a composite attribute"
	ERROBJECTNOTFOUND           string = "%v not found"
	ERROBJECTNOTRECOGNIZED      string = " %v is not recognized"
	ERRPIPEINPUTFILECONFLICT    string = "cannot provide input file when piping in input"
	ERRPROJECTIONFLAGNOTCUSTOM  string = "--projection must be set to 'custom' in order to use custom field mask"
	ERRQUERYABLEFLAG1ARG        string = "only one argument is allowed with --queryable flag"
	ERRQUERYANDCOMPOSITEFLAGS   string = "cannot provide both --composite and --queryable flags"
	ERRQUERYANDDELETEDFLAGS     string = "cannot provide both --query and --deleted flags"
	ERRTOOMANYARGSMAX1          string = "too many arguments, %v has maximum of 1"
	ERRTOOMANYARGSMAX2          string = "too many arguments, %v has maximum of 2"
	ERRUNEXPECTEDATTRCHAR       string = "unexpected character %v found in attribute string"
	ERRUNEXPECTEDQUERYCHAR      string = "unexpected character %v found in query string"

	// Infos

	INFOADMINSET             string = "administrator set to: %v"
	INFOCDEVACTIONPERFORMED  string = "%s successfully performed on ChromeOS device: %s"
	INFOCDEVMOVEPERFORMED    string = "ChromeOS device: %s moved to: %s"
	INFOCDEVUPDATED          string = "ChromeOS device updated: %s"
	INFOCONFIGFILENOTFOUND   string = "Config file not found"
	INFOCREDENTIALPATHSET    string = "service account credential path set to: %v"
	INFOCREDENTIALSSET       string = "credentials set using: %v"
	INFOCUSTOMERIDSET        string = "customer ID set to: %v"
	INFOENVVARSNOTFOUND      string = "No environment variables found"
	INFOGROUPCREATED         string = "group created: %s"
	INFOGROUPALIASCREATED    string = "group alias: %s created for group: %s"
	INFOGROUPALIASDELETED    string = "group alias: %s deleted for group: %s"
	INFOGROUPDELETED         string = "group deleted: %s"
	INFOGROUPSETTINGSCHANGED string = "group settings changed for group: %s"
	INFOINITCANCELLED        string = "init command cancelled"
	INFOINITCOMPLETED        string = "init completed successfully"
	INFOGROUPUPDATED         string = "group updated: %s"
	INFOLOGPATHSET           string = "log path set to: %v"
	INFOMDEVACTIONPERFORMED  string = "%s successfully performed on mobile device: %s"
	INFOMDEVDELETED          string = "mobile device deleted: %s"
	INFOMEMBERCREATED        string = "member: %s created in group: %s"
	INFOMEMBERDELETED        string = "member: %s deleted from group: %s"
	INFOMEMBERUPDATED        string = "member: %s updated in group: %s"
	INFOOUCREATED            string = "orgunit created: %s"
	INFOOUDELETED            string = "orgunit deleted: %s"
	INFOOUUPDATED            string = "orgunit updated: %s"
	INFOSCHEMACREATED        string = "schema created: %s"
	INFOSCHEMADELETED        string = "schema deleted: %s"
	INFOSCHEMAUPDATED        string = "schema updated: %s"
	INFOSETCOMMANDCANCELLED  string = "set command cancelled"
	INFOUSERCREATED          string = "user created: %s"
	INFOUSERALIASCREATED     string = "user alias: %s created for user: %s"
	INFOUSERALIASDELETED     string = "user alias: %s deleted for user: %s"
	INFOUSERDELETED          string = "user deleted: %s"
	INFOUSERUPDATED          string = "user updated: %s"
	INFOUSERUNDELETED        string = "user undeleted: %s"
)
