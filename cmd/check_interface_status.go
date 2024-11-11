package cmd

import (
	"github.com/inexio/thola/internal/device"
	"github.com/inexio/thola/internal/request"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	addDeviceFlags(checkInterfaceStatusCMD)
	addInterfaceOptionsFlags(checkInterfaceStatusCMD)
	checkCMD.AddCommand(checkInterfaceStatusCMD)

	checkInterfaceStatusCMD.Flags().Bool("oper-status", true, "Check oper status of interface")
	checkInterfaceStatusCMD.Flags().Bool("admin-status", false, "Check admin status of interface")
	checkInterfaceStatusCMD.Flags().String("status", "up", "Desired interface state")
}

var checkInterfaceStatusCMD = &cobra.Command{
	Use:   "interface-status",
	Short: "Check operational or admin status of interfaces",
	Long:  "Check operational or admin status of interfaces.",
	Run: func(cmd *cobra.Command, args []string) {
		operStatus, err := cmd.Flags().GetBool("oper-status")
		if err != nil {
			log.Fatal().Err(err).Msg("oper-status needs to be a boolean")
		}

		adminStatus, err := cmd.Flags().GetBool("admin-status")
		if err != nil {
			log.Fatal().Err(err).Msg("admin-status needs to be a boolean")
		}

		desiredState, err := cmd.Flags().GetString("status")
		if err != nil {
			log.Fatal().Err(err).Msg("status needs to be a string")
		}

		r := request.CheckInterfaceStatusRequest{
			OperStatus:         operStatus,
			AdminStatus:        adminStatus,
			DesiredState:       device.Status(desiredState),
			InterfaceOptions:   getInterfaceOptions(),
			CheckDeviceRequest: getCheckDeviceRequest(args[0]),
		}

		handleRequest(&r)
	},
}
