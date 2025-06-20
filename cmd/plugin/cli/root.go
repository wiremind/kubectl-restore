package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tj/go-spin"
	"github.com/wiremind/kubectl-restore/pkg/logger"
	"github.com/wiremind/kubectl-restore/pkg/plugin"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	KubernetesConfigFlags *genericclioptions.ConfigFlags
)

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "kubectl-restore",
		Short:         "",
		Long:          `.`,
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if err := viper.BindPFlags(cmd.Flags()); err != nil {
				panic(fmt.Errorf("failed to bind flags: %w", err))
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger()
			log.Info("")

			s := spin.New()
			finishedCh := make(chan bool, 1)
			namespaceName := make(chan string, 1)
			go func() {
				lastNamespaceName := ""
				for {
					select {
					case <-finishedCh:
						fmt.Printf("\r")
						return
					case n := <-namespaceName:
						lastNamespaceName = n
					case <-time.After(time.Millisecond * 100):
						if lastNamespaceName == "" {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s", s.Next())
						} else {
							fmt.Printf("\r  \033[36mSearching for namespaces\033[m %s (%s)", s.Next(), lastNamespaceName)
						}
					}
				}
			}()
			defer func() {
				finishedCh <- true
			}()

			if err := plugin.RunPlugin(KubernetesConfigFlags, namespaceName); err != nil {
				return errors.Unwrap(err)
			}

			log.Info("")

			return nil
		},
	}

	cobra.OnInitialize(initConfig)

	KubernetesConfigFlags = genericclioptions.NewConfigFlags(false)
	KubernetesConfigFlags.AddFlags(cmd.Flags())

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	cmd.AddCommand(databaseCmd)

	return cmd
}

func InitAndExecute() {
	if err := RootCmd().Execute(); err != nil {
		fmt.Println(err)
		osExit(1)
	}
}

func initConfig() {
	viper.AutomaticEnv()
}
