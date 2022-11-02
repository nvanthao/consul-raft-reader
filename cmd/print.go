/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"log"

	"github.com/nvanthao/consul-raft-reader/app"
	"github.com/spf13/cobra"
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print current Raft logs",
	Long:  "Print current Raft logs",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires path to raft.db file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		s, _ := cmd.Flags().GetInt("start")
		e, _ := cmd.Flags().GetInt("end")

		start := uint64(s)
		end := uint64(e)

		store, err := app.NewStore(filePath)
		if err != nil {
			log.Fatalf("unable to read Raft file: %s", err)
		}

		if start < store.FirstIndex || end > store.LastIndex {
			log.Fatalf("allowed range: [%d, %d]", store.FirstIndex, store.LastIndex)
		}

		// default to print all logs
		if end == 0 {
			end = store.LastIndex
		}

		store.Print(start, end)
	},
	Example: "print --start 1 --end 10 raft.db",
}

func init() {
	rootCmd.AddCommand(printCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// printCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	printCmd.Flags().Int("start", 1, "Start index")
	printCmd.Flags().Int("end", 0, "End index")
}
