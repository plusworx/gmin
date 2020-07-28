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
	cfg "github.com/plusworx/gmin/utils/config"
	"github.com/spf13/cobra"
	admin "google.golang.org/api/admin/directory/v1"
)

var batchUpdOUCmd = &cobra.Command{
	Use:     "orgunits -i <input file path>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Short:   "Updates a batch of orgunits",
	Long: `Updates a batch of orgunits.
	
	Examples: gmin batch-update orgunits -i inputfile.txt -p /Skunkworks/Engineering
	          gmin bupd ous -i inputfile.txt -b`,
	RunE: doBatchUpdMember,
}

func doBatchUpdOU(cmd *cobra.Command, args []string) error {
	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
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
		ouName := scanner.Text()

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = updateOU(ds, customerID, ouName)
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

func updateOU(ds *admin.Service, customerID string, ouName string) error {
	var (
		orgunit *admin.OrgUnit
		ouPath  = []string{}
	)

	ouPath = append(ouPath, ouName)
	orgunit = new(admin.OrgUnit)

	if blockInherit {
		orgunit.BlockInheritance = true
	}

	if unblockInherit {
		orgunit.BlockInheritance = false
		orgunit.ForceSendFields = append(orgunit.ForceSendFields, "BlockInheritance")
	}

	if parentOUPath != "" {
		orgunit.ParentOrgUnitPath = parentOUPath
	}

	ouuc := ds.Orgunits.Update(customerID, ouPath, orgunit)
	_, err := ouuc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: orgunit " + ouName + " updated ****")

	return nil
}

func init() {
	batchUpdateCmd.AddCommand(batchUpdOUCmd)

	batchUpdOUCmd.Flags().BoolVarP(&blockInherit, "blockinherit", "b", false, "block orgunit policy inheritance")
	batchUpdOUCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data file")
	batchUpdOUCmd.Flags().StringVarP(&parentOUPath, "parentpath", "p", "", "orgunit parent path")
	batchUpdOUCmd.Flags().BoolVarP(&unblockInherit, "unblockinherit", "u", false, "unblock orgunit policy inheritance")
}
