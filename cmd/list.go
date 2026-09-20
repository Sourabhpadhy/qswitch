/*
Copyright © 2025 Manpreet Singh <mannuvilasara@gmail.com>
*/
package cmd

import (
	"encoding/json"
	"fmt"

	"qswitch/utils"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available flavours",
	Long:  `List all available flavours configured in the config.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := utils.LoadConfig()
		status, _ := cmd.Flags().GetBool("status")
		var allFlavours []string
		seen := make(map[string]bool)
		for _, f := range config.Flavours {
			if !seen[f] {
				seen[f] = true
				allFlavours = append(allFlavours, f)
			}
		}
		for k := range config.Keybinds {
			if !seen[k] {
				seen[k] = true
				allFlavours = append(allFlavours, k)
			}
		}

		if status {
			type FlavourStatus struct {
				Name      string `json:"name"`
				Installed bool   `json:"installed"`
				Icon      string `json:"icon,omitempty"`
				Color     string `json:"color,omitempty"`
			}
			var statuses []FlavourStatus
			for _, f := range allFlavours {
				icon := utils.GetAssetPath(f)
				color := ""
				if icon != "" {
					color = utils.ExtractDominantColor(icon)
				}
				statuses = append(statuses, FlavourStatus{
					Name:      f,
					Installed: utils.IsFlavourInstalled(f, config),
					Icon:      icon,
					Color:     color,
				})
			}
			jsonData, _ := json.Marshal(statuses)
			fmt.Println(string(jsonData))
		} else {
			for _, f := range allFlavours {
				fmt.Println(f)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("status", "s", false, "Show installation status in JSON format")
}
