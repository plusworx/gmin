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
	"fmt"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	flgnm "github.com/plusworx/gmin/utils/flagnames"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	admin "google.golang.org/api/admin/directory/v1"
)

var updateOUCmd = &cobra.Command{
	Use:     "orgunit <orgunit path or id>",
	Aliases: []string{"ou"},
	Args:    cobra.ExactArgs(1),
	Example: `gmin update orgunit Sales -n "New Name" -d "New description"
gmin upd ou Engineering/Aerodynamics -p Engineering/Aeronautics`,
	Short: "Updates an orgunit",
	Long:  `Updates an orgunit .`,
	RunE:  doUpdateOU,
}

func doUpdateOU(cmd *cobra.Command, args []string) error {
	logger.Debugw("starting doUpdateOU()",
		"args", args)

	var (
		flagsPassed []string
		orgunit     *admin.OrgUnit
	)

	orgunit = new(admin.OrgUnit)

	// Collect names of command flags passed in
	cmd.Flags().Visit(func(f *pflag.Flag) {
		flagsPassed = append(flagsPassed, f.Name)
	})

	// Process command flags
	err := processUpdOUFlags(cmd, orgunit, flagsPassed)
	if err != nil {
		logger.Error(err)
		return err
	}

	ds, err := cmn.CreateDirectoryService(admin.AdminDirectoryOrgunitScope)
	if err != nil {
		logger.Error(err)
		return err
	}

	customerID, err := cfg.ReadConfigString(cfg.CONFIGCUSTID)
	if err != nil {
		logger.Error(err)
		return err
	}

	ouuc := ds.Orgunits.Update(customerID, args[0], orgunit)
	_, err = ouuc.Do()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof(gmess.INFO_OUUPDATED, args[0])
	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_OUUPDATED, args[0])))

	logger.Debug("finished doUpdateOU()")
	return nil
}

func init() {
	updateCmd.AddCommand(updateOUCmd)

	updateOUCmd.Flags().BoolVarP(&blockInherit, flgnm.FLG_BLOCKINHERIT, "b", false, "block orgunit policy inheritance")
	updateOUCmd.Flags().StringVarP(&orgUnitDesc, flgnm.FLG_DESCRIPTION, "d", "", "orgunit description")
	updateOUCmd.Flags().StringVarP(&orgUnitName, flgnm.FLG_NAME, "n", "", "orgunit name")
	updateOUCmd.Flags().StringVarP(&parentOUPath, flgnm.FLG_PARENTPATH, "p", "", "orgunit parent path")
}

func processUpdOUFlags(cmd *cobra.Command, orgunit *admin.OrgUnit, flagNames []string) error {
	logger.Debugw("starting processUpdOUFlags()",
		"flagNames", flagNames)
	for _, flName := range flagNames {
		if flName == flgnm.FLG_BLOCKINHERIT {
			flgBlkInheritVal, err := cmd.Flags().GetBool(flgnm.FLG_BLOCKINHERIT)
			if err != nil {
				logger.Error(err)
				return err
			}
			uoBlockInheritFlag(orgunit, flgBlkInheritVal)
		}
		if flName == flgnm.FLG_DESCRIPTION {
			flgDescriptionVal, err := cmd.Flags().GetString(flgnm.FLG_DESCRIPTION)
			if err != nil {
				logger.Error(err)
				return err
			}
			uoDescriptionFlag(orgunit, flgDescriptionVal)
		}
		if flName == flgnm.FLG_NAME {
			flgNameVal, err := cmd.Flags().GetString(flgnm.FLG_NAME)
			if err != nil {
				logger.Error(err)
				return err
			}
			err = uoNameFlag(orgunit, "--"+flName, flgNameVal)
			if err != nil {
				return err
			}
		}
		if flName == flgnm.FLG_PARENTPATH {
			flgParentPathVal, err := cmd.Flags().GetString(flgnm.FLG_PARENTPATH)
			if err != nil {
				logger.Error(err)
				return err
			}
			err = uoParentPathFlag(orgunit, "--"+flName, flgParentPathVal)
			if err != nil {
				return err
			}
		}
	}
	logger.Debug("finished processUpdOUFlags()")
	return nil
}

func uoBlockInheritFlag(orgunit *admin.OrgUnit, flgVal bool) {
	logger.Debugw("starting uoBlockInheritFlag()",
		"flgVal", flgVal)
	if flgVal {
		orgunit.BlockInheritance = true
	} else {
		orgunit.BlockInheritance = false
		orgunit.ForceSendFields = append(orgunit.ForceSendFields, "BlockInheritance")
	}
	logger.Debug("finished uoBlockInheritFlag()")
}

func uoDescriptionFlag(orgunit *admin.OrgUnit, flgVal string) {
	logger.Debugw("starting uoDescriptionFlag()",
		"flgVal", flgVal)
	if flgVal == "" {
		orgunit.ForceSendFields = append(orgunit.ForceSendFields, "Description")
	}
	orgunit.Description = flgVal
	logger.Debug("finished uoDescriptionFlag()")
}

func uoNameFlag(orgunit *admin.OrgUnit, flagName string, flgVal string) error {
	logger.Debugw("starting uoNameFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			return err
		}
	}
	orgunit.Name = flgVal
	logger.Debug("finished uoNameFlag()")
	return nil
}

func uoParentPathFlag(orgunit *admin.OrgUnit, flagName string, flgVal string) error {
	logger.Debugw("starting uoParentPathFlag()",
		"flagName", flagName,
		"flgVal", flgVal)
	if flgVal == "" {
		err := fmt.Errorf(gmess.ERR_EMPTYSTRING, flagName)
		if err != nil {
			return err
		}
	}
	orgunit.ParentOrgUnitPath = flgVal
	logger.Debug("finished uoParentPathFlag()")
	return nil
}
