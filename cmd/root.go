package cmd

import (
	"compound/config"
	"compound/core"
	"errors"
	"fmt"
	"os"

	"github.com/fox-one/pkg/store/db"
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file. default is ./config/config.yaml")
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
		filename := "./config/config.yaml"
		info, err := os.Stat(filename)
		if !os.IsNotExist(err) && !info.IsDir() {
			cfgFile = filename
		}
	}

	if cfgFile == "" {
		panic(errors.New("no config file"))
	}

	logrus.Infoln("use config file: ", cfgFile)

	if err := config.Load(cfgFile, &cfg); err != nil {
		panic(err)
	}

	logrus.Infoln("load config successful!!")
}

func initLog() {
	if debugMode {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	structs.DefaultTagName = "json"
}

func migrateDB() {
	logrus.Infoln("start migrate db")
	database := provideDatabase()
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		logrus.Errorln("migrate db error:", err)
		panic(err)
	}
}

// Run run command
func Run(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
