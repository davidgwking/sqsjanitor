// Copyright © 2017 David King <davidgwking@gmail.com>
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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var awsRegion string
var awsAccessKeyID string
var awsSecretAccessKey string

var RootCmd = &cobra.Command{
	Use:   "sqsjanitor",
	Short: "",
	Long:  ``,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sqsjanitor.yaml)")

	RootCmd.PersistentFlags().StringVar(&awsRegion, "aws-region", "", "aws region")
	RootCmd.PersistentFlags().StringVar(&awsAccessKeyID, "aws-access-key-id", "", "aws access key id")
	RootCmd.PersistentFlags().StringVar(&awsSecretAccessKey, "aws-secret-access-key", "", "aws secret access key")

	viper.BindPFlag("aws-region", RootCmd.PersistentFlags().Lookup("aws-region"))
	viper.BindPFlag("aws-access-key-id", RootCmd.PersistentFlags().Lookup("aws-access-key-id"))
	viper.BindPFlag("aws-secret-access-key", RootCmd.PersistentFlags().Lookup("aws-secret-access-key"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".sqsjanitor")
	viper.AddConfigPath("$HOME")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
