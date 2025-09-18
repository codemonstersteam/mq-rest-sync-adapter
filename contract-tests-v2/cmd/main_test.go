package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// executeCommand executes a cobra command and returns the output
func executeCommand(cmd *cobra.Command, args ...string) string {
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)
	cmd.Execute()
	return buf.String()
}

func TestCLI_VersionCommand(t *testing.T) {
	// GIVEN: CLI приложение
	cmd := getRootCmd()

	// WHEN: запускаем команду version
	output := executeCommand(cmd, "version")

	// THEN: возвращается версия приложения
	assert.Contains(t, output, "contract-validator version")
}

func TestCLI_HelpCommand(t *testing.T) {
	// GIVEN: CLI приложение
	cmd := getRootCmd()

	// WHEN: запускаем команду help
	output := executeCommand(cmd, "help")

	// THEN: содержит описание доступных команд
	assert.Contains(t, output, "contract-validator")
	assert.Contains(t, output, "version")
}

func TestCLI_RootCommandWithoutArgs(t *testing.T) {
	// GIVEN: CLI приложение
	cmd := getRootCmd()

	// WHEN: запускаем без аргументов
	output := executeCommand(cmd)

	// THEN: показывается help
	assert.Contains(t, strings.ToLower(output), "usage")
}