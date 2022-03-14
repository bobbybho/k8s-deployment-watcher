package cmd

import (
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dwsh",
		Args:  cobra.NoArgs,
		Short: "dwsh commands",
		Long:  `dwsh commands`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	kubeConfigPathDefault = ""
	kubeConfigPath        = ""

	nameSpaceDefault = "default"
	namespace        = ""
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func getKubeConfigPath() string {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	return filepath.Join(home, ".kube", "config")
}

func init() {
	kubeConfigPathDefault = getKubeConfigPath()
	kubeConfigPath = kubeConfigPathDefault

	rootCmd.AddCommand(deploymentCmd)
	rootCmd.AddCommand(podCmd)
	rootCmd.AddCommand(podControllerCmd)
	rootCmd.AddCommand(podBotCmd)
}
