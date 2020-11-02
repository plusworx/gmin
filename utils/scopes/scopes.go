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

package scopes

const (
	EXITTEXT      string = "<Exit>"
	SELECTALLTEXT string = "<Select All>"
)

// Selection is used to store scope selection data
type Selection struct {
	IsScope  bool
	Selected bool
	Text     string
}

// OauthScopes are valid scopes used by gmin
var OauthScopes = []string{
	"https://www.googleapis.com/auth/admin.directory.device.chromeos",
	"https://www.googleapis.com/auth/admin.directory.device.chromeos.readonly",
	"https://www.googleapis.com/auth/admin.directory.device.mobile",
	"https://www.googleapis.com/auth/admin.directory.device.mobile.action",
	"https://www.googleapis.com/auth/admin.directory.device.mobile.readonly",
	"https://www.googleapis.com/auth/admin.directory.group",
	"https://www.googleapis.com/auth/admin.directory.group.member",
	"https://www.googleapis.com/auth/admin.directory.group.member.readonly",
	"https://www.googleapis.com/auth/admin.directory.group.readonly",
	"https://www.googleapis.com/auth/admin.directory.orgunit",
	"https://www.googleapis.com/auth/admin.directory.orgunit.readonly",
	"https://www.googleapis.com/auth/admin.directory.user",
	"https://www.googleapis.com/auth/admin.directory.user.alias",
	"https://www.googleapis.com/auth/admin.directory.user.alias.readonly",
	"https://www.googleapis.com/auth/admin.directory.user.readonly",
	"https://www.googleapis.com/auth/admin.directory.userschema",
	"https://www.googleapis.com/auth/admin.directory.userschema.readonly",
	"https://www.googleapis.com/auth/apps.groups.settings",
	"https://www.googleapis.com/auth/drive.readonly",
}
