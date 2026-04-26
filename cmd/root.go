package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	awsclient "github.com/bikramdhoju/ecsctl/internal/aws"
	"github.com/bikramdhoju/ecsctl/internal/config"
	"github.com/bikramdhoju/ecsctl/internal/output"
)

var (
	cfgPath         string
	ctxOverride     string
	clusterOverride string
	regionOverride  string
	profileOverride string
	outputFlag      string

	globalCfg     *config.EcsConfig
	globalClients *awsclient.Clients
	globalCtx     *config.Context
	globalPrinter *output.Printer
)

var rootCmd = &cobra.Command{
	Use:   "ecsctl",
	Short: "kubectl-like CLI for AWS ECS",
	Long:  `ecsctl — manage AWS ECS clusters, services, tasks, and containers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", config.DefaultConfigPath, "path to ecsctl config file")
	rootCmd.PersistentFlags().StringVar(&ctxOverride, "context", "", "override active context")
	rootCmd.PersistentFlags().StringVarP(&clusterOverride, "cluster", "c", "", "override cluster from context")
	rootCmd.PersistentFlags().StringVar(&regionOverride, "region", "", "override AWS region")
	rootCmd.PersistentFlags().StringVar(&profileOverride, "profile", "", "override AWS profile (~/.aws/credentials)")
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "output format: table|wide|json|yaml")
}

// initConfig loads config and resolves the active context.
// Called by commands that need config but NOT AWS credentials (e.g. ecsctl config *).
func initConfig() error {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}
	globalCfg = cfg

	if ctxOverride != "" {
		globalCfg.CurrentContext = ctxOverride
	}

	return nil
}

// initClients loads config, resolves context, validates credentials, and builds AWS clients.
// Called by commands that need AWS access.
func initClients(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}

	activeCtx, err := globalCfg.ActiveContext()
	if err != nil {
		return err
	}

	// Apply CLI overrides on top of the resolved context
	if clusterOverride != "" {
		activeCtx.Cluster = clusterOverride
	}
	if regionOverride != "" {
		activeCtx.Region = regionOverride
	}
	globalCtx = activeCtx

	clients, err := awsclient.NewClients(context.Background(), activeCtx, profileOverride)
	if err != nil {
		return err
	}

	if err := clients.ValidateCredentials(context.Background()); err != nil {
		return err
	}

	globalClients = clients

	fmt, err := output.ParseFormat(outputFlag)
	if err != nil {
		return err
	}
	if outputFlag == "" && globalCtx.Output != "" {
		fmt, _ = output.ParseFormat(globalCtx.Output)
	}
	globalPrinter = output.New(fmt)

	return nil
}

func mustCluster() (string, error) {
	if globalClients.Cluster == "" {
		return "", fmt.Errorf("cluster not set; use --cluster or configure it in context")
	}
	return globalClients.Cluster, nil
}
