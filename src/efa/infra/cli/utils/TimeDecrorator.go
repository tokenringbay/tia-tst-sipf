package utils

import (
	"fmt"
	"github.com/spf13/cobra"
	"time"
)

//RunE type definition for a command
type RunE func(cmd *cobra.Command, args []string) error

//TimedRunE provides a Time Decorator over CLI commands
func TimedRunE(f RunE) RunE {
	return func(cmd *cobra.Command, args []string) error {

		defer func(t time.Time) {
			fmt.Printf("--- Time Elapsed: %v ---\n", time.Since(t))
		}(time.Now())

		return f(cmd, args)
	}
}
