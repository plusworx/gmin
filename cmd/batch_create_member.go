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
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchCrtMemberCmd = &cobra.Command{
	Use:     "group-members -g <group email address or id> -i <input file path>",
	Aliases: []string{"group-member", "grp-members", "grp-member", "gmembers", "gmember", "gmems", "gmem"},
	Short:   "Creates a batch of group members",
	Long:    `Creates a batch of group members.`,
	RunE:    doBatchCrtMember,
}

func doBatchCrtMember(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupMemberScope)
	if err != nil {
		return err
	}

	if group == "" {
		err := errors.New("gmin: error - group must be provided")
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
			err = createMember(ds, jsonData)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "invalid") ||
				strings.Contains(err.Error(), "Member already exists") {
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

func createMember(ds *admin.Service, jsonData string) error {
	var member *admin.Member

	member = new(admin.Member)
	jsonBytes := []byte(jsonData)

	err := json.Unmarshal(jsonBytes, &member)
	if err != nil {
		return err
	}

	mic := ds.Members.Insert(group, member)
	newMember, err := mic.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** member " + newMember.Email + " created in group " + group + " ****")

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtMemberCmd)

	batchCrtMemberCmd.Flags().StringVarP(&group, "group", "g", "", "email address or id of group")
	batchCrtMemberCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group member data file")
}
