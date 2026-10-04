package main

import (
	"os"

	"k8s.io/component-base/cli"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/gorizond/tiered-node-scheduler/pkg/plugins/fallback"
)

func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(fallback.Name, fallback.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
