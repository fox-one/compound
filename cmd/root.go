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
	cobra.OnInitialize(func() {
		onInitialize(initConfig, initLog)
	})

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file. default is $HOME/.compound.yaml")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable or disable debug model")
}

func onInitialize(fs ...func()) {
	if initialized {
		return
	}

	if len(fs) > 0 {
		for _, f := range fs {
			f()
		}
	}

	initialized = true
}

func initConfig() {
	if cfgFile == "" {
		dir, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		filename := path.Join(dir, ".compound.yaml")
		fmt.Println(filename)
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

	fmt.Println(cfg)
}

func initLog() {
	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	structs.DefaultTagName = "json"
}

// Run run command
func Run(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
