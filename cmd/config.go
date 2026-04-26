package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bikramdhoju/ecsctl/internal/config"
)

var configCmd = &cobra.Command{
	Use:               "config",
	Short:             "Manage ecsctl configuration and contexts",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return initConfig() },
}

var configViewCmd = &cobra.Command{
	Use:   "view",
	Short: "Print the current config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, entry := range globalCfg.Contexts {
			marker := "  "
			if entry.Name == globalCfg.CurrentContext {
				marker = "* "
			}
			fmt.Printf("%s%s\n", marker, entry.Name)
			fmt.Printf("    cluster:     %s\n", entry.Context.Cluster)
			fmt.Printf("    region:      %s\n", entry.Context.Region)
			if entry.Context.AWSProfile != "" {
				fmt.Printf("    aws-profile: %s\n", entry.Context.AWSProfile)
			}
			if entry.Context.RoleARN != "" {
				fmt.Printf("    role-arn:    %s\n", entry.Context.RoleARN)
			}
		}
		return nil
	},
}

var configGetContextsCmd = &cobra.Command{
	Use:   "get-contexts",
	Short: "List all configured contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(globalCfg.Contexts) == 0 {
			fmt.Println("No contexts configured.")
			return nil
		}
		fmt.Printf("%-3s %-30s %-50s %-15s %s\n", "", "NAME", "CLUSTER", "REGION", "PROFILE")
		for _, entry := range globalCfg.Contexts {
			current := " "
			if entry.Name == globalCfg.CurrentContext {
				current = "*"
			}
			fmt.Printf("%-3s %-30s %-50s %-15s %s\n",
				current,
				entry.Name,
				entry.Context.Cluster,
				entry.Context.Region,
				entry.Context.AWSProfile,
			)
		}
		return nil
	},
}

var configCurrentContextCmd = &cobra.Command{
	Use:   "current-context",
	Short: "Show the active context name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if globalCfg.CurrentContext == "" {
			fmt.Println("(none)")
			return nil
		}
		fmt.Println(globalCfg.CurrentContext)
		return nil
	},
}

var configUseContextCmd = &cobra.Command{
	Use:   "use-context <name>",
	Short: "Switch the active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		found := false
		for _, entry := range globalCfg.Contexts {
			if entry.Name == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("context %q not found; run: ecsctl config get-contexts", name)
		}
		globalCfg.CurrentContext = name
		if err := config.Save(globalCfg, cfgPath); err != nil {
			return err
		}
		fmt.Printf("Switched to context %q.\n", name)
		return nil
	},
}

var (
	setContextCluster string
	setContextRegion  string
	setContextProfile string
	setContextRole    string
	setContextOutput  string
)

var configSetContextCmd = &cobra.Command{
	Use:   "set-context <name>",
	Short: "Create or update a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Start from existing context if it exists, apply only what was explicitly set
		existing := config.Context{}
		for _, entry := range globalCfg.Contexts {
			if entry.Name == name {
				existing = entry.Context
				break
			}
		}

		if cmd.Flags().Changed("cluster") {
			existing.Cluster = setContextCluster
		}
		if cmd.Flags().Changed("region") {
			existing.Region = setContextRegion
		}
		if cmd.Flags().Changed("profile") {
			existing.AWSProfile = setContextProfile
		}
		if cmd.Flags().Changed("role-arn") {
			existing.RoleARN = setContextRole
		}
		if cmd.Flags().Changed("output") {
			existing.Output = setContextOutput
		}

		if existing.Cluster == "" {
			return fmt.Errorf("--cluster is required")
		}
		if existing.Region == "" {
			return fmt.Errorf("--region is required")
		}

		globalCfg.SetContext(name, existing)
		if globalCfg.CurrentContext == "" {
			globalCfg.CurrentContext = name
		}

		if err := config.Save(globalCfg, cfgPath); err != nil {
			return err
		}
		fmt.Printf("Context %q saved.\n", name)
		return nil
	},
}

var configDeleteContextCmd = &cobra.Command{
	Use:   "delete-context <name>",
	Short: "Remove a context from the config",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !globalCfg.DeleteContext(name) {
			return fmt.Errorf("context %q not found", name)
		}
		if globalCfg.CurrentContext == name {
			globalCfg.CurrentContext = ""
		}
		if err := config.Save(globalCfg, cfgPath); err != nil {
			return err
		}
		fmt.Printf("Context %q deleted.\n", name)
		return nil
	},
}

func init() {
	configSetContextCmd.Flags().StringVar(&setContextCluster, "cluster", "", "cluster name or ARN")
	configSetContextCmd.Flags().StringVar(&setContextRegion, "region", "", "AWS region")
	configSetContextCmd.Flags().StringVar(&setContextProfile, "profile", "", "AWS credentials profile (~/.aws/credentials)")
	configSetContextCmd.Flags().StringVar(&setContextRole, "role-arn", "", "IAM role ARN to assume")
	configSetContextCmd.Flags().StringVar(&setContextOutput, "output", "table", "default output format: table|wide|json|yaml")

	configCmd.AddCommand(
		configViewCmd,
		configGetContextsCmd,
		configCurrentContextCmd,
		configUseContextCmd,
		configSetContextCmd,
		configDeleteContextCmd,
	)
	rootCmd.AddCommand(configCmd)
}
