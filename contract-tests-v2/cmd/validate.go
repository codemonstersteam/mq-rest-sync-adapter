package main

import (
	"errors"
	"fmt"

	"github.com/codemonsters/contract-tests-v2/validator"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate [config-file]",
	Short: "Валидация контрактов AsyncAPI 3.0",
	Long: `Валидация совместимости контрактов между потребителем и поставщиком
по спецификациям AsyncAPI 3.0.

Команда принимает путь к файлу конфигурации contract-tests.yaml и проверяет
совместимость каналов, протоколов и структур сообщений.

Пример использования:
  contract-validator validate ./contract-tests.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
	configPath := args[0]

	// Проверяем, что файл конфигурации указан
	if configPath == "" {
		return errors.New("config file required")
	}

	// Создаем валидатор и запускаем валидацию
	contractValidator := validator.NewContractValidator()
	result, err := contractValidator.Validate(configPath)

	if err != nil {
		// Пока что просто возвращаем ошибку, форматирование добавим в следующих шагах
		return fmt.Errorf("validation failed: %w", err)
	}

	// Пока что простой вывод успеха, красивое форматирование добавим в следующих шагах
	fmt.Fprintf(cmd.OutOrStdout(), "✅ Контракты совместимы\n")
	fmt.Fprintf(cmd.OutOrStdout(), "Потребитель: %s\n", result.ConsumerChannelName)

	return nil
}

func init() {
	// Добавляем validate команду к root команде
	rootCmd.AddCommand(validateCmd)
}

// getValidateCmd возвращает validate команду для использования в тестах
func getValidateCmd() *cobra.Command {
	return validateCmd
}