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
	Long: `Updates a batch of group members.
	
	Examples: gmin batch-update members anothergroup@mycompany.com -i inputfile.txt -d DAILY
	          gmin bupd mem finance@mycompany.com -i inputfile.txt -r MEMBER`,
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
		memberKey := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateMember(ds, memberKey, args[0])
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") {
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

func updateMember(ds *admin.Service, memberKey string, group string) error {
	var member = admin.Member{}

	if deliverySetting != "" {
		validDS, err := mems.ValidateDeliverySetting(deliverySetting)
		if err != nil {
			return err
		}
		member.DeliverySettings = validDS
	}

	if role != "" {
		validRole, err := mems.ValidateRole(role)
		if err != nil {
			return err
		}
		member.Role = validRole
	}

	muc := ds.Members.Update(group, memberKey, &member)
	_, err := muc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: member " + memberKey + " updated in group " + group + " ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdMemberCmd)

	batchUpdMemberCmd.Flags().StringVarP(&deliverySetting, "deliverysetting", "d", "", "member delivery setting")
	batchUpdMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to member data file")
	batchUpdMemberCmd.Flags().StringVarP(&role, "role", "r", "", "member role")
}