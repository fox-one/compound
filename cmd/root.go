package cmd

import (
	"compound/config"
	"compound/core"
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yiplee/structs"
)

var (
	cfgFile     string
	cfg         core.Config
	debugMode   bool
	initialized bool
)

var rootCmd = cobra.Command{
	Use:   "compound",
	Short: "compound engine",
}

func init() {
	cobra.OnInitialize(initConfig, initLogging, initDone)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file. default is ~/.rings-node.yaml")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable or disable debug model")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ver string) {
	rootCmd.Version = ver
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {
	if initialized {
		return
	}

	if cfgFile == "" {
		dir, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		filename := path.Join(dir, ".rings-node.yaml")
		info, err := os.Stat(filename)
		if !os.IsNotExist(err) && !info.IsDir() {
			cfgFile = filename
		}
	}

	if cfgFile != "" {
		logrus.Debugln("use config file", cfgFile)
	}

	if err := config.Load(cfgFile, &cfg); err != nil {
		panic(err)
	}
}

func initLogging() {
	if initialized {
		return
	}

	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp: true,
	}
	logrus.SetFormatter(formatter)

	structs.DefaultTagName = "json"
}

func initDone() {
	initialized = true
}
