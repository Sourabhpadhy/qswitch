package cmd

import (
	"fmt"

	"qswitch/utils"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [flavour]",
	Short: "Switch to or apply a flavour",
	Long:  `Switch to a specific flavour or apply the current flavour's configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := utils.LoadConfig()
		currentFlag, _ := cmd.Flags().GetBool("current")
		debugFlag, _ := cmd.Flags().GetBool("debug")

		if currentFlag {
			currentFlavour := utils.ReadState()

			if currentFlavour == "" {
				fmt.Println("No flavour is currently set.")
				return
			}

			if !utils.IsValidFlavour(currentFlavour, config) {
				fmt.Println("No valid current flavour set.")
				return
			}

			if !utils.IsFlavourInstalled(currentFlavour, config) {
				fmt.Println("Current flavour is not installed:", currentFlavour)
				return
			}

			utils.ApplyFlavour(currentFlavour, config, debugFlag)
			fmt.Println("Applied current flavour:", currentFlavour)
			return
		}

		if len(args) != 1 {
			fmt.Println("Invalid usage. Use 'qswitch apply <flavour>' or 'qswitch apply --current'.")
			return
		}

		flavour := args[0]

		if !utils.IsValidFlavour(flavour, config) {
			fmt.Println("Unknown flavour:", flavour)
			fmt.Println("Run 'qswitch --help' to list flavours.")
			return
		}

		if !utils.IsFlavourInstalled(flavour, config) {
			fmt.Println("Flavour not installed:", flavour)
			fmt.Println("Keybind file not found in ~/.config/qswitch/keybinds/")
			return
		}

		current := utils.ReadState()
		if current == flavour {
			fmt.Println("Already running:", flavour)
			return
		}

		utils.WriteState(flavour)
		utils.ApplyFlavour(flavour, config, debugFlag)
		fmt.Println("Switched to", flavour)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().Bool("current", false, "Apply the current flavour")
	applyCmd.Flags().BoolP("debug", "d", false, "Print debug output")
}