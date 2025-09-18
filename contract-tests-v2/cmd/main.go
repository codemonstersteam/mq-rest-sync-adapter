package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	version = "0.1.0"
	appName = "contract-validator"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Инструмент валидации контрактов AsyncAPI 3.0",
	Long: `contract-validator - инструмент для валидации совместимости
контрактов между потребителем и поставщиком по спецификациям AsyncAPI 3.0.

Инструмент проверяет совместимость каналов, протоколов и структур сообщений
для обеспечения корректного взаимодействия между сервисами.`,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Показать версию приложения",
	Long:  "Показать текущую версию contract-validator",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", appName, version)
	},
}

func init() {
	// Добавляем команды к root команде
	rootCmd.AddCommand(versionCmd)
}

// getRootCmd возвращает корневую команду для использования в тестах
func getRootCmd() *cobra.Command {
	return rootCmd
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}