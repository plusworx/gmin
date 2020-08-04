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

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	cmn "github.com/plusworx/gmin/utils/common"
	mems "github.com/plusworx/gmin/utils/members"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdMemberCmd = &cobra.Command{
	Use:     "group-members <group email address, alias or id> -i <input file path>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Args:    cobra.ExactArgs(1),
	Short:   "Updates a batch of group members",
	Long: `Updates a batch of group members where group member details are provided in a JSON input file.
	
	Examples:	gmin batch-update group-members sales@mycompany.com -i inputfile.json
			gmin bupd gmems engineering@mycompany.com -i inputfile.json
			  
	The contents of the JSON file should look something like this:
	
	{"email":"rudolph.brown@mycompany.com","delivery_settings":"DIGEST","role":"MANAGER"}
	{"email":"emily.bronte@mycompany.com","delivery_settings":"DAILY","role":"MEMBER"}
	{"email":"charles.dickens@mycompany.com","delivery_settings":"NONE","role":"MEMBER"}`,
	RunE: doBatchUpdMember,
}

func doBatchUpdMember(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	if inputFile == "" {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		jsonData := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateMember(ds, jsonData, args[0])
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "not a valid") ||
				strings.Contains(err.Error(), "should be") ||
				strings.Contains(err.Error(), "unrecognized") {
				return backoff.Permanent(err)
			}

			return err
		}, b)
		if err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func updateMember(ds *admin.Service, jsonData string, group string) error {
	var member = admin.Member{}

	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, mems.MemberAttrMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &member)
	if err != nil {
		return err
	}

	if member.Email == "" {
		return errors.New("gmin: error - email must be included in the JSON input string")
	}

	if member.DeliverySettings != "" {
		validDS, err := mems.ValidateDeliverySetting(member.DeliverySettings)
		if err != nil {
			return err
		}
		member.DeliverySettings = validDS
	}

	if member.Role != "" {
		validRole, err := mems.ValidateRole(member.Role)
		if err != nil {
			return err
		}
		member.Role = validRole
	}

	muc := ds.Members.Update(group, member.Email, &member)
	_, err = muc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: member " + member.Email + " updated in group " + group + " ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group member data file")
}
