package commands

import (
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdPEF() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pef",
		Short: "pef",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdPEFCapabilities())
	cmd.AddCommand(NewCmdPEFStatus())
	cmd.AddCommand(NewCmdPEFFilter())

	return cmd
}

func NewCmdPEFCapabilities() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities",
		Short: "capabilities",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetPEFCapabilities()
			if err != nil {
				CheckErr(fmt.Errorf("GetPEFCapabilities failed, err: %s", err))
			}

			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdPEFStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "status",
		Run: func(cmd *cobra.Command, args []string) {

			{
				res, err := client.GetLastProcessedEventId()
				if err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %s", err))
				}
				fmt.Println(res.Format())
			}

			{
				param := &ipmi.PEFConfigParam_Control{}
				if err := client.GetPEFConfigForParameter(param, 0, 0); err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %s", err))
				}
				fmt.Println(param.Format())
			}

			{
				param := &ipmi.PEFConfigParam_ActionGlobalControl{}
				if err := client.GetPEFConfigForParameter(param, 0, 0); err != nil {
					CheckErr(fmt.Errorf("GetLastProcessedEventId failed, err: %s", err))
				}
				fmt.Println(param.Format())
			}

		},
	}
	return cmd
}

func NewCmdPEFFilter() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "filter",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				cmd.Help()
			}
		},
	}
	cmd.AddCommand(NewCmdPEFFilterList())
	return cmd
}

func NewCmdPEFFilterList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {

			var numberOfEventFilters uint8

			{
				param := &ipmi.PEFConfigParam_NumberOfEventFilters{}
				if err := client.GetPEFConfigForParameter(param, 0, 0); err != nil {
					CheckErr(fmt.Errorf("get number of event filters failed, err: %s", err))
				}
				numberOfEventFilters = param.Value
			}

			var eventFilters = make([]*ipmi.PEFEventFilter, numberOfEventFilters)
			for i := uint8(1); i <= numberOfEventFilters; i++ {
				param := &ipmi.PEFConfigParam_EventFilter{}
				if err := client.GetPEFConfigForParameter(param, i, 0); err != nil {
					CheckErr(fmt.Errorf("get event filter entry %d failed, err: %s", i, err))
				}
				eventFilters[i-1] = param.Entry
			}

			fmt.Println(ipmi.FormatEventFilters(eventFilters))
		},
	}
	return cmd
}
