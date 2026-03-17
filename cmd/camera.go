package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cameraCmd)
	cameraCmd.AddCommand(cameraSnapshotCmd)

	cameraSnapshotCmd.Flags().StringP("output-file", "O", "", "Output file path (default: <entity_id>.jpg)")
}

var cameraCmd = &cobra.Command{
	Use:   "camera",
	Short: "Manage Home Assistant cameras",
	Long: `Access camera entities for snapshots.

Examples:
  hactl camera snapshot camera.front_door
  hactl camera snapshot camera.front_door -O snapshot.jpg`,
}

var cameraSnapshotCmd = &cobra.Command{
	Use:   "snapshot <camera_entity_id>",
	Short: "Download a camera snapshot",
	Long: `Download the current snapshot from a camera entity.

Examples:
  hactl camera snapshot camera.front_door
  hactl camera snapshot camera.front_door -O /tmp/snapshot.jpg`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		entityID := args[0]
		outputFile, _ := cmd.Flags().GetString("output-file")
		if outputFile == "" {
			outputFile = entityID + ".jpg"
		}

		client, err := newAPIClient()
		if err != nil {
			return err
		}

		if err := client.GetToFile("/api/camera_proxy/"+entityID, outputFile); err != nil {
			return err
		}

		printer.Success(fmt.Sprintf("Snapshot saved to %s", outputFile))
		return nil
	},
}
