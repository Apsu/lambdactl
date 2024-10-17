package cmd

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"lambdactl/pkg/ui"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var lambdaFS embed.FS

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "lambdactl",
	Short: "A CLI for managing Lambda instances",
	RunE: func(cmd *cobra.Command, args []string) error {
		return ui.Start()
	},
}

func Execute(fs embed.FS) {
	lambdaFS = fs

	initConfig()
	checkRequiredConfig()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.lambda.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".lambda")
	}

	// Override with LAMBDA_* env vars, where . and - map to _
	viper.SetEnvPrefix("LAMBDA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error was produced
			fmt.Printf("Error reading config file: %v\n", err)
		}
	}
}

func checkRequiredConfig() {
	requiredKeys := []string{"api-url", "api-key"}
	missingKeys := []string{}

	for _, key := range requiredKeys {
		if !viper.IsSet(key) {
			missingKeys = append(missingKeys, key)
		}
	}

	if len(missingKeys) > 0 {
		fmt.Printf("Missing required configuration keys: %v\n", missingKeys)
		os.Exit(1)
	}
}
