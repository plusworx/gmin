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

var batchCrtGroupCmd = &cobra.Command{
	Use:     "groups -i <input file path>",
	Aliases: []string{"group", "grps", "grp"},
	Short:   "Creates a batch of groups",
	Long: `Creates a batch of groups where group details are provided in a JSON input file.
	
	Examples: 	gmin batch-create groups -i inputfile.json
			gmin bcrt grp -i inputfile.json
			  
	The contents of the JSON file should look something like this:
	
	{"description":"Finance group","mail":"finance@mycompany.com","name":"Finance"}
	{"description":"Marketing group","email":"marketing@mycompany.com","name":"Marketing"}
	{"description":"Sales group","email":"sales@mycompany.com","name":"Sales"}`,
	RunE: doBatchCrtGroup,
}

func doBatchCrtGroup(cmd *cobra.Command, args []string) error {
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
			err = createGroup(ds, jsonData)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Missing required field") ||
				strings.Contains(err.Error(), "invalid character") ||
				strings.Contains(err.Error(), "Entity already exists") {
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

func createGroup(ds *admin.Service, jsonData string) error {
	var group *admin.Group

	group = new(admin.Group)
	jsonBytes := []byte(jsonData)

	err := json.Unmarshal(jsonBytes, &group)
	if err != nil {
		return err
	}

	gic := ds.Groups.Insert(group)
	newGroup, err := gic.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: group " + newGroup.Email + " created ****")

	return nil
}

func init() {
	batchCreateCmd.AddCommand(batchCrtGroupCmd)

	batchCrtGroupCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to group data file")
}
