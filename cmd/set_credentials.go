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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	cmn "github.com/plusworx/gmin/utils/common"
	cfg "github.com/plusworx/gmin/utils/config"
	gmess "github.com/plusworx/gmin/utils/gminmessages"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setCredentialsCmd = &cobra.Command{
	Use:     "credentials",
	Aliases: []string{"creds", "cred"},
	Example: `gmin set credentials
gmin set creds`,
	Short: "Sets gmin service account credentials",
	Long: `Sets gmin service account credentials.
			
User is presented with a list of files contained in the credentials path directory. The dialogue will look
like this:

1) mycompany.json
2) another_company.json
3) yet_another_company.json

Please choose a file by typing the number:

Once the user chooses a file, that file is copied to gmin_credentials in the credentials path directory.`,
	RunE: doSetCredentials,
}

func askForCredentialsFile(nFiles int) int {
	var response string

	fmt.Println("")
	fmt.Print("Please choose a file by typing the number (q to quit): ")

	_, err := fmt.Scanln(&response)
	if err != nil {
		fmt.Println(gmess.ERR_FILENUMBERREQUIRED)
		return askForCredentialsFile(nFiles)
	}

	if response == "q" {
		return cmn.QUIT
	}

	fileNum, err := strconv.Atoi(response)
	if err != nil {
		fmt.Println(gmess.ERR_INVALIDFILENUMBER)
		return askForCredentialsFile(nFiles)
	}

	if fileNum > nFiles || fileNum < 1 {
		fmt.Println(gmess.ERR_INVALIDFILENUMBER)
		return askForCredentialsFile(nFiles)
	}

	return fileNum
}

func doSetCredentials(cmd *cobra.Command, args []string) error {
	var (
		justFiles  []os.FileInfo
		validFiles []os.FileInfo
	)

	credPath := viper.GetString(cfg.CONFIGCREDPATH)

	files, err := ioutil.ReadDir(credPath)
	if err != nil {
		return err
	}

	// Remove directories from slice
	for _, f := range files {
		if !f.IsDir() {
			justFiles = append(justFiles, f)
		}
	}

	// Remove gmin_credentials from slice
	for _, f := range justFiles {
		if f.Name() != cfg.CREDENTIALFILE {
			validFiles = append(validFiles, f)
		}
	}

	for idx, f := range validFiles {
		fmt.Println(strconv.Itoa(idx+1) + ") " + f.Name())
	}

	fileNum := askForCredentialsFile(len(validFiles))
	if fileNum == cmn.QUIT {
		fmt.Println(gmess.INFO_SETCOMMANDCANCELLED)
		return nil
	}

	// Copy new credentials file
	usedName := validFiles[fileNum-1].Name()
	newCred, err := os.Open(filepath.Join(filepath.ToSlash(credPath), usedName))
	if err != nil {
		return err
	}
	defer newCred.Close()

	gminCred, err := os.Create(filepath.Join(filepath.ToSlash(credPath), cfg.CREDENTIALFILE))
	if err != nil {
		return err
	}
	defer gminCred.Close()

	_, err = io.Copy(gminCred, newCred)
	if err != nil {
		fmt.Println(err)
	}

	err = gminCred.Sync()
	if err != nil {
		return err
	}

	fmt.Println(cmn.GminMessage(fmt.Sprintf(gmess.INFO_CREDENTIALSSET, usedName)))
	return nil
}

func init() {
	setCmd.AddCommand(setCredentialsCmd)
}
