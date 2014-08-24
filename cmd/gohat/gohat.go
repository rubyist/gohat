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

	var dumpParamsCommand = &cobra.Command{
		Use:   "params",
		Short: "Show the heap parameters",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			dumpParams := heapFile.DumpParams()
			if dumpParams.BigEndian {
				fmt.Println("Big Endian")
			} else {
				fmt.Println("Little Endian")
			}
			fmt.Println("Pointer Size:", dumpParams.PtrSize)
			fmt.Println("Channel Header Size:", dumpParams.ChHdrSize)
			fmt.Printf("Heap Starting Address %02x\n", dumpParams.StartAddress)
			fmt.Printf("Heap Ending Address: %02x\n", dumpParams.EndAddress)
			fmt.Println("Architecture:", dumpParams.Arch)
			fmt.Println("GOEXPERIMENT:", dumpParams.GoExperiment)
			fmt.Println("nCPU:", dumpParams.NCPU)
		},
	}
	gohatCmd.AddCommand(dumpParamsCommand)

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
			fmt.Println("General statistics")
			fmt.Println("Alloc:", memstats.Alloc)
			fmt.Println("TotalAlloc:", memstats.TotalAlloc)
			fmt.Println("Sys:", memstats.Sys)
			fmt.Println("Lookups:", memstats.Lookups)
			fmt.Println("Mallocs:", memstats.Mallocs)
			fmt.Println("Frees:", memstats.Frees)
			fmt.Println("")
			fmt.Println("Main allocation heap statistics")
			fmt.Println("HeapAlloc:", memstats.HeapAlloc)
			fmt.Println("HeapSys:", memstats.HeapSys)
			fmt.Println("HeapIdle:", memstats.HeapIdle)
			fmt.Println("HeapInuse:", memstats.HeapInuse)
			fmt.Println("HeapReleased:", memstats.HeapReleased)
			fmt.Println("HeapObjects:", memstats.HeapObjects)
			fmt.Println("")
			fmt.Println("Low-level fixed-size structure allocator statistics")
			fmt.Println("StackInuse:", memstats.StackInuse)
			fmt.Println("StatckSys:", memstats.StackSys)
			fmt.Println("MSpanInuse:", memstats.MSpanInuse)
			fmt.Println("MSpanSys:", memstats.MSpanSys)
			fmt.Println("BuckHashSys:", memstats.BuckHashSys)
			fmt.Println("GCSys:", memstats.GCSys)
			fmt.Println("OtherSys:", memstats.OtherSys)
			fmt.Println("")
			fmt.Println("Garbage collector statistics")
			fmt.Println("NextGC:", memstats.NextGC)
			fmt.Println("LastGC:", memstats.LastGC)
			fmt.Println("PauseTotalNs:", memstats.PauseTotalNs)
			fmt.Println("NumGC:", memstats.NumGC)
			fmt.Println("Last GC Pauses:")
			fmt.Println(memstats.PauseNs)
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
				fmt.Printf("%016x %s %s %d\n", object.Address, typeName, object.Kind(), object.Size)
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

			fmt.Printf("%016x %s %d %d\n", object.Address, object.Kind(), object.Size, len(object.Content))
			if object.Type != nil {
				fmt.Println(object.Type.Name)
			}
			fmt.Println("")
			fmt.Println(object.Content)
			fmt.Println("")

			if object.Type != nil {
				fmt.Println("Field List:")
				var lastOffset uint64
				for idx, field := range object.Type.FieldList {
					if idx == len(object.Type.FieldList)-1 {
						fmt.Printf("%s 0x%016x  %v\n", field.Kind(), field.Offset, []byte(object.Content[lastOffset:]))
					} else {
						lastOffset = object.Type.FieldList[idx].Offset
						nextOffset := object.Type.FieldList[idx+1].Offset
						fmt.Printf("%s 0x%016x  %v\n", field.Kind(), field.Offset, []byte(object.Content[lastOffset:nextOffset]))
					}
				}
			}
		},
	}
	gohatCmd.AddCommand(objectCommand)

	var sameCommand = &cobra.Command{
		Use:   "same",
		Short: "find objects that are the same in two heap files",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile1, err := verifyHeapDumpFile(args)
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			if len(args) != 2 {
				fmt.Println("same <heap file> <heap file>")
				os.Exit(1)
			}

			heapFile2, err := heapfile.New(args[1])
			if err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			heapObjects1 := heapFile1.Objects()

			same := make([]uint64, 0, len(heapObjects1))

			for _, obj := range heapObjects1 {
				if cmp := heapFile2.Object(int64(obj.Address)); cmp != nil {
					if cmp.TypeAddress == obj.TypeAddress &&
						cmp.Kind() == obj.Kind() &&
						cmp.Content == obj.Content &&
						cmp.Size == obj.Size {
						same = append(same, cmp.Address)
					}
				}
			}

			for _, addr := range same {
				object := heapFile2.Object(int64(addr))
				fmt.Printf("%016x %d %d %d\n", object.Address, object.Kind, object.Size, len(object.Content))
				fmt.Println([]byte(object.Content))
				fmt.Println(object.Content)
				if object.Type != nil {
					fmt.Println("Type: ", object.Type.Name)
					fmt.Println("Field List:")
					for _, field := range object.Type.FieldList {
						fmt.Printf("%s 0x%016x\n", field.Kind(), field.Offset)
					}
				}
				fmt.Printf("\n\n")
			}
		},
	}
	gohatCmd.AddCommand(sameCommand)

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
				fmt.Printf("%016x %d %s\n", t.Address, len(t.FieldList), t.Name)
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
			fmt.Printf("%016x %d %s\n", t.Address, len(t.FieldList), t.Name)
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
