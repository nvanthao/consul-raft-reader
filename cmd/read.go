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

// readCmd represents the read command
var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read Raft log at given index",
	Long:  "Read Raft log at given index",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires path to raft.db file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		index, _ := cmd.Flags().GetInt("index")

		store, err := app.NewStore(filePath)
		if err != nil {
			log.Fatalf("unable to read Raft file: %s", err)
		}

		store.Read(uint64(index))

	},
	Example: "read --index 1 raft.db",
}

func init() {
	rootCmd.AddCommand(readCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// readCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	readCmd.Flags().Int("index", 0, "Index of log to read")
}
