package main

import (
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

func main() {
	var gohatCmd = &cobra.Command{
		Use:   "gohat",
		Short: "gohat is go heap dump analysis tool",
		Long: `Gohat can read and query go heap dump files.
Complete documentation is available at http://github.com/rubyist/gohat`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	var memStatsCommand = &cobra.Command{
		Use:   "memstats",
		Short: "Dump the memstats",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			memstats := heapFile.MemStats()
			fmt.Println(memstats)
		},
	}
	gohatCmd.AddCommand(memStatsCommand)

	var objectsCommand = &cobra.Command{
		Use:   "objects",
		Short: "Dump a list of objects",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			objects := heapFile.Objects()
			for _, object := range objects {
				typeName := "<unknown>"
				if object.Type != nil {
					typeName = object.Type.Name
				}
				fmt.Printf("%x %s %d %d\n", object.Address, typeName, object.Kind, object.Size)
			}
		},
	}
	gohatCmd.AddCommand(objectsCommand)

	var objectCommand = &cobra.Command{
		Use:   "object",
		Short: "Dump the contents of an object",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if len(args) != 2 {
				fmt.Println("object <heap file> <address>")
				os.Exit(1)
			}

			addr, _ := strconv.ParseInt(args[1], 16, 64)
			object := heapFile.Object(addr)
			fmt.Println(object.Content)
		},
	}
	gohatCmd.AddCommand(objectCommand)

	var typesCommand = &cobra.Command{
		Use:   "types",
		Short: "Dump all the types found in the heap",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			types := heapFile.Types()
			for _, t := range types {
				fmt.Printf("%x %d %s\n", t.Address, len(t.FieldList), t.Name)
			}
		},
	}
	gohatCmd.AddCommand(typesCommand)

	var typeCommand = &cobra.Command{
		Use:   "type",
		Short: "Dump information about a type",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if len(args) != 2 {
				fmt.Println("type <heap file> <address>")
				os.Exit(1)
			}

			addr, _ := strconv.ParseInt(args[1], 16, 64)

			t := heapFile.Type(addr)
			fmt.Printf("%x %d %s\n", t.Address, len(t.FieldList), t.Name)
			for _, field := range t.FieldList {
				fmt.Println(field)
			}
		},
	}
	gohatCmd.AddCommand(typeCommand)

	gohatCmd.Execute()
}

func verifyHeapDumpFile(args []string) (*heapfile.HeapFile, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("heap file required")
	}
	heapFile, err := heapfile.New(args[0])
	return heapFile, err
}
