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

var batchDelOrgUnitCmd = &cobra.Command{
	Use:     "orgunits [-i input file path]",
	Aliases: []string{"orgunit", "ous", "ou"},
	Short:   "Deletes a batch of orgunits",
	Long: `Deletes a batch of orgunits where orgunit details are provided in a text input file or through a pipe.
	
	Examples:	gmin batch-delete orgunits -i inputfile.txt
			gmin bdel ous -i inputfile.txt
			gmin ls ous -o TestOU -a orgunitpath | jq '.organizationUnits[] | .orgUnitPath' -r | gmin bdel ou
			
	The input should have the orgunit paths or ids to be deleted on separate lines like this:
	
	Engineering/Skunkworx
	Engineering/SecretOps
	Engineering/Surplus`,
	RunE: doBatchDelOrgUnit,
}

func doBatchDelOrgUnit(cmd *cobra.Command, args []string) error {
	var ouPaths = []string{}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		return err
	}

	customerID, err := cfg.ReadConfigString("customerid")
	if err != nil {
		return err
	}

	scanner, err := cmn.InputFromStdIn(inputFile)
	if err != nil {
		return err
	}

	if inputFile == "" && scanner == nil {
		err := errors.New("gmin: error - must provide inputfile")
		return err
	}

	if scanner == nil {
		file, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner = bufio.NewScanner(file)
	}

	for scanner.Scan() {
		text := scanner.Text()
		ouPaths = []string{}

		if text[0] == '/' {
			text = text[1:]
		}
		ouPaths = append(ouPaths, text)

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = deleteOU(ds, customerID, ouPaths)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Resource Not Found") ||
				strings.Contains(err.Error(), "Org unit not found") ||
				strings.Contains(err.Error(), "Bad Request") {
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

func deleteOU(ds *admin.Service, customerID string, ouPaths []string) error {
	oudc := ds.Orgunits.Delete(customerID, ouPaths)

	err := oudc.Do()
	if err != nil {
		return err
	}

	fmt.Println("**** gmin: orgunit " + ouPaths[0] + " deleted ****")

	return nil
}

func init() {
	batchDelCmd.AddCommand(batchDelOrgUnitCmd)

	batchDelOrgUnitCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data text file")
}
