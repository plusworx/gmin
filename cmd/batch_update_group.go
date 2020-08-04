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
	Long: `Updates a batch of groups where group details are provided in a JSON input file.
	
	Examples:	gmin batch-update groups -i inputfile.json
			gmin bupd grps -i inputfile.json
			  
	The contents of the JSON file should look something like this:
	
	{"groupKey":"034gixby5n7pqal","email":"testgroup@mycompany.com","name":"Testing","description":"This is a testing group for all your testing needs."}
	{"groupKey":"032hioqz3p4ulyk","email":"info@mycompany.com","name":"Information","description":"Group for responding to general queries."}
	{"groupKey":"045fijmz6w8nkqc","email":"webmaster@mycompany.com","name":"Webmaster","description":"Group for responding to website queries."}

N.B. groupKey (group email address, alias or id) must be provided.`,
	RunE: doBatchUpdGrp,
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
				strings.Contains(err.Error(), "should be") ||
				strings.Contains(err.Error(), "must be included") {
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
	var (
		group  *admin.Group
		grpKey = grps.Key{}
	)

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

	err = json.Unmarshal(jsonBytes, &grpKey)
	if err != nil {
		return err
	}

	if grpKey.GroupKey == "" {
		return errors.New("gmin: error - groupKey must be included in the JSON input string")
	}

	err = json.Unmarshal(jsonBytes, &group)
	if err != nil {
		return err
	}

	guc := ds.Groups.Update(grpKey.GroupKey, group)
	_, err = guc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: group " + grpKey.GroupKey + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdGrpCmd)

	batchUpdGrpCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data file")
}
