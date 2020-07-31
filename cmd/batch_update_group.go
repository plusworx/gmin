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
	grps "github.com/plusworx/gmin/utils/groups"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdGrpCmd = &cobra.Command{
	Use:     "groups -i <input file path>",
	Aliases: []string{"group", "grps", "grp"},
	Short:   "Updates a batch of groups",
	Long:    `Updates a batch of groups.`,
	RunE:    doBatchUpdGrp,
}

func doBatchUpdGrp(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryGroupScope)
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
			err = updateGroup(ds, jsonData)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "not valid") ||
				strings.Contains(err.Error(), "unrecognized") ||
				strings.Contains(err.Error(), "should be") {
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

func updateGroup(ds *admin.Service, jsonData string) error {
	var group *admin.Group

	group = new(admin.Group)
	jsonBytes := []byte(jsonData)

	if !json.Valid(jsonBytes) {
		return errors.New("gmin: error - attribute string is not valid JSON")
	}

	outStr, err := cmn.ParseInputAttrs(jsonBytes)
	if err != nil {
		return err
	}

	err = cmn.ValidateInputAttrs(outStr, grps.GroupAttrMap)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &group)
	if err != nil {
		return err
	}

	if group.Email == "" {
		return errors.New("gmin: error - group email must be included in the JSON input string")
	}

	guc := ds.Groups.Update(group.Email, group)
	updatedGrp, err := guc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: group " + updatedGrp.Email + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdGrpCmd)

	batchUpdGrpCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data file")
}
