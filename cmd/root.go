// Copyright © 2019 Lucas Wilson-Richter <da.maestro@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"

	epub "github.com/bmaupin/go-epub"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	miniflux "miniflux.app/client"
)

var cfgFile string
var outputFile string
var markRead bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "miniflux-epub",
	Short: "Turns your unread miniflux entries into an epub for offline reading",
	Long: `Turns your unread miniflux entries into an epub for offline reading

Usage: miniflux-epub [--outputfile=filename.epub]
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		minifluxUrl := viper.GetString("MinifluxUrl")
		username := viper.GetString("Username")
		password := viper.GetString("Password")

		// Connect to miniflux API with configured creds
		fmt.Printf("Connecting to %s as %s", minifluxUrl, username)
		mnflx := miniflux.New(minifluxUrl, username, password)

		// Find the configured category
		fmt.Println("Grabbing category list")
		categoryName := viper.GetString("Category")
		categories, err := mnflx.Categories()

		if err != nil {
			handleError(err)
		}

		var category miniflux.Category

		for i := 0; i < len(categories); i++ {
			if categories[i].Title == categoryName {
				category = *categories[i]
				break
			}
		}

		// Get all the unread entries in the category
		unreadFilter := miniflux.Filter{
			Status:    "unread",
			Limit:     100,
			Direction: "asc",
		}

		result, err := mnflx.Entries(&unreadFilter)
		entries := result.Entries

		if err != nil {
			handleError(err)
		}

		matchingEntries := make([]miniflux.Entry, 0)
		for i := 0; i < len(entries); i++ {
			if entries[i].Feed.Category.ID == category.ID {
				fmt.Printf("%d: %s\n", i, entries[i].Title)
				matchingEntries = append(matchingEntries, *entries[i])
			}
		}

		// Make an epub from the entries
		pub := epub.NewEpub("Miniflux Entries")
		pub.SetAuthor("miniflux-epub by @lucaswilric")

		for i := 0; i < len(matchingEntries); i++ {
			entry := matchingEntries[i]
			title := "<h1><a href=\"" + entry.URL + "\">" + entry.Title + "</a></h1>"
			feed := "<p>from <em>" + entry.Feed.Title + "</em></p>"
			content := entry.Content

			sectionBody := title + feed + content

			pub.AddSection(sectionBody, entry.Title, "", "")

		}

		err = pub.Write(viper.GetString("outputfile"))
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleError(err error) {
	fmt.Printf("%v", err)
	os.Exit(1)
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("Username", "", "Loaded from config")
	viper.BindPFlag("Username", rootCmd.PersistentFlags().Lookup("Username"))

	rootCmd.PersistentFlags().String("Password", "", "Loaded from config")
	viper.BindPFlag("Password", rootCmd.PersistentFlags().Lookup("Password"))

	rootCmd.PersistentFlags().StringVar(&outputFile, "outputfile", "miniflux.epub", "output file (default is miniflux.epub)")
	viper.BindPFlag("outputfile", rootCmd.PersistentFlags().Lookup("outputfile"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".miniflux-epub" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".miniflux-epub")
	}

	viper.SetDefault("MinifluxUrl", "https://reader.miniflux.app/")
	viper.SetDefault("Username", "")
	viper.SetDefault("Password", "")
	viper.SetDefault("Category", "All")

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
