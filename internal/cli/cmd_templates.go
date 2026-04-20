package cli

import (
	"fmt"

	"github.com/paulplee/dockyard/internal/template"
	"github.com/spf13/cobra"
)

func newTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List available templates",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			names, err := template.List()
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Println("(no templates embedded)")
				return nil
			}
			for _, n := range names {
				m, err := template.LoadManifest(n)
				if err != nil {
					fmt.Printf("  %-12s  (error: %v)\n", n, err)
					continue
				}
				fmt.Printf("  %-12s  %s\n", n, m.Description)
			}
			return nil
		},
	}
}
