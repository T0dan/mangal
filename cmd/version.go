package cmd

import (
	"encoding/json"
	"github.com/metafates/mangal/constants"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
)

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(versionLatestCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of mangal",
	Long:  `All software has versions. This is mangal's`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("mangal version " + constants.Version)
	},
}

var versionLatestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Print the latest version number of mangal",
	Long:  `It will fetch the latest version from github and print it`,
	Run: func(cmd *cobra.Command, args []string) {
		resp, err := http.Get("https://api.github.com/repos/metafates/mangal/releases/latest")
		if err != nil {
			cmd.PrintErr(err)
			os.Exit(1)
		}

		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		var release struct {
			TagName string `json:"tag_name"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			cmd.PrintErr(err)
			os.Exit(1)
		}

		// remove the v from the tag name
		latestVersion := release.TagName[1:]

		cmd.Println("mangal latest version " + latestVersion)
	},
}