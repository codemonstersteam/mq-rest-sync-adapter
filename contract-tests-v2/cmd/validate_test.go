package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateCommand_WithConfigFile(t *testing.T) {
	// GIVEN: существует файл contract-tests.yaml
	configFile := "testdata/cmd/valid-contract-tests.yaml"

	// WHEN: запускаем validate с файлом конфигурации
	rootCommand := getRootCmd()
	output := executeCommand(rootCommand, "validate", configFile)

	// THEN: команда выполняется без ошибок парсинга аргументов
	// На данном этапе мы ожидаем, что команда не падает на парсинге аргументов
	// Ошибки валидации контрактов - это нормально, главное чтобы аргументы парсились
	assert.NotContains(t, output, "required flag")
	assert.NotContains(t, output, "unknown command")
	assert.NotContains(t, output, "accepts 1 arg(s), received")
}

func TestValidateCommand_MissingConfigFile(t *testing.T) {
	// GIVEN: команда validate без аргументов

	// WHEN: запускаем validate без аргументов используя executeCommand helper
	rootCommand := getRootCmd()
	output := executeCommand(rootCommand, "validate")

	// THEN: возвращается ошибка с понятным сообщением в выводе
	assert.Contains(t, output, "accepts 1 arg(s), received 0")
}

func TestValidateCommand_NonExistentConfigFile(t *testing.T) {
	// GIVEN: несуществующий файл конфигурации
	nonExistentFile := "path/to/non/existent/config.yaml"

	// WHEN: запускаем validate с несуществующим файлом
	cmd := getValidateCmd()
	cmd.SetArgs([]string{nonExistentFile})
	err := cmd.Execute()

	// THEN: возвращается ошибка о том, что файл не найден
	assert.Error(t, err)
	// Пока что проверяем, что хотя бы не падаем с паникой
	assert.NotNil(t, err)
}