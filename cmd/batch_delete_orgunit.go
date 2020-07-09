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
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spf13/cobra"
)

var batchDelOrgUnitCmd = &cobra.Command{
	Use:     "orgunits -i <input file path>",
	Aliases: []string{"orgunit", "ous", "ou"},
	Short:   "Deletes a batch of orgunits",
	Long:    `Deletes a batch of orgunits.`,
	RunE:    doBatchDelOrgUnit,
}

func doBatchDelOrgUnit(cmd *cobra.Command, args []string) error {
	var orgunits []string

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
		orgunits = []string{}

		orgunits = append(orgunits, scanner.Text())

		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 30 * time.Second

		err = backoff.Retry(func() error {
			var err error
			err = doDeleteOU(nil, orgunits)
			if err == nil {
				return err
			}

			if strings.Contains(err.Error(), "Org unit not found") ||
				strings.Contains(err.Error(), "Cannot delete entity with members") ||
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

func init() {
	batchDelCmd.AddCommand(batchDelOrgUnitCmd)

	batchDelOrgUnitCmd.Flags().StringVarP(&inputFile, "inputfile", "i", "", "filepath to orgunit data text file")
}