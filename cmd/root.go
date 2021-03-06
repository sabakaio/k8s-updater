// Copyright © 2016 Sabaka OU <hello@sabaka.io>
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

	log "github.com/Sirupsen/logrus"
	"github.com/sabakaio/k8s-updater/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

var cfgFile string

// Kubernetes API client
var k *client.Client

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "k8s-updater",
	Short: "Kubernetes cluster containers autoupdater",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		update()
	},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig, initLogging, initClient)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.k8s-updater.yaml)")
	RootCmd.PersistentFlags().StringP("loglevel", "l", log.DebugLevel.String(), "Log level")
	RootCmd.PersistentFlags().StringP("host", "H", "", "Kubernetes host to connect to")
	RootCmd.PersistentFlags().StringP("namespace", "n", api.NamespaceDefault, "Kubernetes namespace")
	RootCmd.PersistentFlags().Bool("dry-run", false, "Get versions but do not update")
	viper.BindPFlag("loglevel", RootCmd.PersistentFlags().Lookup("loglevel"))
	viper.BindPFlag("host", RootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("dryrun", RootCmd.PersistentFlags().Lookup("dry-run"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".k8s-updater") // name of config file (without extension)
	viper.AddConfigPath("$HOME")        // adding home directory as first search path
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initLogging() {
	level, err := log.ParseLevel(viper.GetString("loglevel"))
	if err != nil {
		log.Fatalln(err)
	}
	log.SetLevel(level)
	log.Debugln("Loglevel set to", log.GetLevel().String())
}

// Create Kubernetes API client for given host. Shared for all subcommands
func initClient() {
	var err error
	k, err = util.CreateClient(viper.GetString("host"))
	if err != nil {
		log.Fatalln("Can't connect to Kubernetes API:", err)
	}
}
