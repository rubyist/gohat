package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	var heapUtilCmd = &cobra.Command{
		Use:   "heaputil",
		Short: "Heaputil is go heap dump reader",
		Long: `Heaputil can read and query go heap dump files.
Complete documentation is available at http://github.com/rubyist/heaputil`,
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	var objectsCommand = &cobra.Command{
		Use:   "objects",
		Short: "Dump a list of objects",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 1 {
				fmt.Println("Please specify a heap dump file")
				os.Exit(1)
			}
			heapFile, _ := NewHeapFile(args[0])
			objects := heapFile.Objects()
			for _, object := range objects {
				fmt.Printf("%x %s %d %d\n", object.Address, object.TypeName, object.Kind, object.Size)
			}
		},
	}
	heapUtilCmd.AddCommand(objectsCommand)

	var objectCommand = &cobra.Command{
		Use:   "object",
		Short: "Dump the contents of an object",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gonna dump an object", args)
		},
	}
	heapUtilCmd.AddCommand(objectCommand)

	heapUtilCmd.Execute()
}
