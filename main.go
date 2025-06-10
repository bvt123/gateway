package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/centralmind/gateway/cli"
	_ "github.com/centralmind/gateway/connectors/bigquery"
	_ "github.com/centralmind/gateway/connectors/clickhouse"
	_ "github.com/centralmind/gateway/connectors/elasticsearch"
	_ "github.com/centralmind/gateway/connectors/mongodb"
	_ "github.com/centralmind/gateway/connectors/mssql"
	_ "github.com/centralmind/gateway/connectors/mysql"
	_ "github.com/centralmind/gateway/connectors/oracle"
	_ "github.com/centralmind/gateway/connectors/postgres"
	_ "github.com/centralmind/gateway/connectors/snowflake"
	_ "github.com/centralmind/gateway/connectors/sqlite"
	_ "github.com/centralmind/gateway/plugins/api_keys"
	_ "github.com/centralmind/gateway/plugins/lru_cache"
	_ "github.com/centralmind/gateway/plugins/lua_rls"
	_ "github.com/centralmind/gateway/plugins/oauth"
	_ "github.com/centralmind/gateway/plugins/otel"
	_ "github.com/centralmind/gateway/plugins/pii_remover"
	_ "github.com/centralmind/gateway/providers/anthropic"
	_ "github.com/centralmind/gateway/providers/bedrock"
	_ "github.com/centralmind/gateway/providers/openai"
)

func main() {
	rootCommand := &cobra.Command{
		Use:          "gateway",
		Short:        "gateway cli",
		Example:      "./gateway help",
		SilenceUsage: true,
	}
	var logLevel string
	rootCommand.PersistentFlags().StringVar(&logLevel, "log-level", "info", "logging level (trace, debug, info, warn, error)")
	rootCommand.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return fmt.Errorf("invalid log level %s", logLevel)
		}
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(level)
		return nil
	}
	cli.RegisterCommand(rootCommand, cli.StartCommand())
	cli.RegisterCommand(rootCommand, cli.Connectors())
	cli.RegisterCommand(rootCommand, cli.Plugins())
	cli.RegisterCommand(rootCommand, cli.Discover())
	cli.RegisterCommand(rootCommand, cli.Connection())
	cli.RegisterCommand(rootCommand, cli.GenerateReadmeCommand())
	err := rootCommand.Execute()
	if err != nil {
		os.Exit(1)
	}
}
