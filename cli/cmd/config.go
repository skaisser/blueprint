package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value from blueprint/.config.yml (supports dot notation)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		data, err := os.ReadFile("blueprint/.config.yml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot read blueprint/.config.yml: %v\n", err)
			os.Exit(1)
		}

		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid YAML: %v\n", err)
			os.Exit(1)
		}

		value, found := getNestedKey(raw, key)
		if !found {
			fmt.Fprintf(os.Stderr, "Error: key '%s' not found\n", key)
			os.Exit(1)
		}

		fmt.Println(value)
	},
}

// getNestedKey retrieves a value from a nested map using dot notation.
func getNestedKey(m map[string]interface{}, key string) (interface{}, bool) {
	parts := strings.SplitN(key, ".", 2)

	val, ok := m[parts[0]]
	if !ok {
		return nil, false
	}

	// If no more nesting, return the value
	if len(parts) == 1 {
		return val, true
	}

	// Try to descend into nested map
	nested, ok := val.(map[string]interface{})
	if !ok {
		return nil, false
	}

	return getNestedKey(nested, parts[1])
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
