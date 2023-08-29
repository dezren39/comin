package cmd

import (
	"github.com/nlewo/comin/config"
	"github.com/nlewo/comin/deploy"
	"github.com/nlewo/comin/http"
	"github.com/nlewo/comin/inotify"
	"github.com/nlewo/comin/poller"
	"github.com/nlewo/comin/state"
	"github.com/nlewo/comin/worker"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var configFilepath string
var dryRun bool

var pollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Poll a repository and deploy the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := config.Read(configFilepath)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

		stateManager, err := state.New(filepath.Join(config.StateFilepath))
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

		deployer, err := deploy.NewDeployer(dryRun, config, stateManager)
		if err != nil {
			logrus.Error(err)
			os.Exit(1)
		}

		wk := worker.NewWorker(deployer.Deploy)

		go poller.Poller(wk, config.Remotes)
		// FIXME: the state should be available from somewhere else...
		go http.Run(wk, config.Webhook, stateManager)
		go inotify.Run(wk, config.Inotify)
		wk.Run()
	},
}

func init() {
	pollCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "dry-run mode")
	pollCmd.PersistentFlags().StringVarP(&configFilepath, "config", "", "", "the configuration file path")
	pollCmd.MarkPersistentFlagRequired("config")
	rootCmd.AddCommand(pollCmd)
}
