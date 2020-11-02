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
	"strings"

	"github.com/manifoldco/promptui"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	scp "github.com/plusworx/gmin/utils/scopes"
	"github.com/spf13/cobra"
)

var getScopesCmd = &cobra.Command{
	Use:     "scopes",
	Aliases: []string{"scps", "scp"},
	Example: `gmin get scopes
gmin get scps`,
	Short: "Outputs list of valid oauth scopes used by gmin and allows selection of one or more as a comma-delimited string",
	Long:  `Outputs list of valid oauth scopes used by gmin and allows selection of one or more as a comma-delimited string.`,
	RunE:  doGetScopes,
}

func doGetScopes(cmd *cobra.Command, args []string) error {
	chosenScopes := []string{}
	selections := []scp.Selection{}
	// Add Exit and Select All
	selections = append(selections, scp.Selection{IsScope: false, Selected: false, Text: scp.SELECTALLTEXT})
	selections = append(selections, scp.Selection{IsScope: false, Selected: false, Text: scp.EXITTEXT})

	for _, scope := range scp.OauthScopes {
		sel := scp.Selection{IsScope: true, Selected: false, Text: scope}
		selections = append(selections, sel)
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "\U000025B6 {{if .IsScope}}{{if .Selected}}[x]{{else}}[ ]{{end}}{{end}} {{.Text | green}}",
		Inactive: "  {{if .IsScope}}{{if .Selected}}[x]{{else}}[ ]{{end}}{{end}} {{.Text}}",
	}

	prompt := promptui.Select{
		Label:        "Select oauth scope(s)",
		HideHelp:     true,
		HideSelected: true,
		Items:        selections,
		Size:         25,
		Templates:    templates,
	}

	err := makeSelection(prompt, selections, 0)
	if err != nil {
		return err
	}

	for _, sel := range selections {
		if sel.Selected && sel.IsScope {
			chosenScopes = append(chosenScopes, sel.Text)
		}
	}
	fmt.Println(strings.Join(chosenScopes, ","))
	return nil
}

func init() {
	getCmd.AddCommand(getScopesCmd)
}

func makeSelection(prompt promptui.Select, selections []scp.Selection, cPos int) error {
	var doAll bool
	prompt.CursorPos = cPos
	numSel, str, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt || err.Error() == "^D" {
			fmt.Println(gmess.INFO_GETSCOPESTERMINATED)
			return nil
		}
		return err
	}

	if strings.Contains(str, "Select") {
		if selections[numSel].Selected == false {
			doAll = true
		} else {
			doAll = false
		}
		for i := 0; i < len(selections); i++ {
			if doAll {
				selections[i].Selected = true
			} else {
				selections[i].Selected = false
			}
		}
		makeSelection(prompt, selections, numSel)
		return nil
	}

	if strings.Contains(str, "Exit") {
		return nil
	}

	if selections[numSel].Selected == true {
		selections[numSel].Selected = false
	} else {
		selections[numSel].Selected = true
	}

	makeSelection(prompt, selections, numSel)
	return nil
}
