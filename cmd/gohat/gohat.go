package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/rubyist/gohat/pkg/heapfile"
	"github.com/spf13/cobra"
	"os"
	"sort"
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

	var allocsCommand = &cobra.Command{
		Use:   "allocs",
		Short: "Dump the alloc stack trace samples",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			allocs := heapFile.Allocs()
			fmt.Println(len(allocs), "alloc samples")
			for _, alloc := range allocs {
				obj := alloc.Object()
				if obj.Type == nil {
					fmt.Println("<unknown>")
				} else {
					fmt.Println(obj.Type.Name)
				}
				record := alloc.Profile()
				fmt.Printf("%x %d %d %d\n", record.Record, record.Size, record.Allocs, record.Frees)
				for _, frame := range record.Frames {
					fmt.Printf("\t%s   %s:%d\n", frame.Name, frame.File, frame.Line)
				}
				fmt.Println()
			}
		},
	}
	gohatCmd.AddCommand(allocsCommand)

	var dataBinary bool
	var dataCommand = &cobra.Command{
		Use:   "data",
		Short: "Dump the data segment",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			data := heapFile.DataSegment()
			if dataBinary {
				os.Stdout.Write([]byte(data.Content))
				return
			}

			fmt.Println("Address:", data.Address)
			for _, field := range data.Fields {
				if field.Kind == heapfile.FieldStr {
					fmt.Printf("%s %d %s\n", field.KindString(), field.Offset, field.Content)
				} else {
					fmt.Printf("%s %d %v\n", field.KindString(), field.Offset, []byte(field.Content))
				}
			}
			fmt.Println(len(data.Content))
		},
	}
	dataCommand.Flags().BoolVarP(&dataBinary, "binary", "b", false, "Dump data contents as binary")
	gohatCmd.AddCommand(dataCommand)

	var bssBinary bool
	var bssCommand = &cobra.Command{
		Use:   "bss",
		Short: "Dump the bss",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			data := heapFile.BSS()

			if bssBinary {
				os.Stdout.Write([]byte(data.Content))
				return
			}

			fmt.Println("Address:", data.Address)
			for _, field := range data.Fields {
				if field.Kind == heapfile.FieldStr {
					fmt.Printf("%s %d %s\n", field.KindString(), field.Offset, field.Content)
				} else {
					fmt.Printf("%s %d %v\n", field.KindString(), field.Offset, []byte(field.Content))
				}
			}
		},
	}
	bssCommand.Flags().BoolVarP(&bssBinary, "binary", "b", false, "Dump data contents as binary")
	gohatCmd.AddCommand(bssCommand)

	var goroutinesCommand = &cobra.Command{
		Use:   "goroutines",
		Short: "Dump goroutines",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			for _, g := range heapFile.Goroutines() {
				fmt.Printf("Goroutine %d\n", g.Id)
				fmt.Printf("\tAddress: %x\n", g.Address)
				fmt.Printf("\tTop of stack: %x\n", g.Top)
				fmt.Printf("\tCreator Location: %x\n", g.Location)
				if g.System {
					fmt.Println("\tSystem Started Go routine")
				}
				if g.Background {
					fmt.Println("\tBackground Go routine")
				}
				fmt.Printf("\tStatus: %s\n", g.Status())
				if reason := g.ReasonWaiting(); reason != "" {
					fmt.Printf("\tReason Waiting: %s\n", reason)
				}
				fmt.Printf("\tLast Started Waiting: %d\n", g.LastWaiting)
				fmt.Printf("\tCurrent Frame: %x\n", g.CurrentFrame)
				fmt.Printf("\tOS Thread %d\n", g.OSThread)
				fmt.Printf("\tTop Defer Record: %x\n", g.DeferRecord)
				fmt.Printf("\tTop Panic Record: %x\n", g.PanicRecord)
				fmt.Println("")
			}
		},
	}
	gohatCmd.AddCommand(goroutinesCommand)

	var histogramCommand = &cobra.Command{
		Use:   "histogram",
		Short: "Dump a histogram of object counts by type",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			counts := make(map[string]int, 0)
			for _, object := range heapFile.Objects() {
				if object.Type != nil {
					counts[object.Type.Name] += 1
				} else {
					counts["<unknown>"] += 1
				}
			}

			histogram := NewHistSorter(counts)
			histogram.Sort()
			for i := 0; i < len(histogram.Keys); i++ {
				fmt.Printf("%d\t%s\n", histogram.Vals[i], histogram.Keys[i])
			}
		},
	}
	gohatCmd.AddCommand(histogramCommand)

	var memProfCommand = &cobra.Command{
		Use:   "memprof",
		Short: "Dump the alloc/free profile records",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			memProf := heapFile.MemProf()
			for _, record := range memProf {
				fmt.Printf("%x %d %d %d\n", record.Record, record.Size, record.Allocs, record.Frees)
				for _, frame := range record.Frames {
					fmt.Printf("\t%s   %s:%d\n", frame.Name, frame.File, frame.Line)
				}
			}
		},
	}
	gohatCmd.AddCommand(memProfCommand)

	var memStatsCommand = &cobra.Command{
		Use:   "memstats",
		Short: "Dump the memstats",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

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

	var containsCommand = &cobra.Command{
		Use:   "contains",
		Short: "Find objects that point to an address",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			if len(args) != 2 {
				fmt.Println("contains <heap file> <address>")
				os.Exit(1)
			}

			addr, _ := strconv.ParseInt(args[1], 16, 64)

			// Check objects
			for _, object := range heapFile.Objects() {
				for _, child := range object.Children() {
					if uint64(addr) == child.Address {
						fmt.Printf("Found in object %x\n", object.Address)
						return
					}
				}
			}

			// Check data segment
			for _, object := range heapFile.DataSegment().Objects() {
				if uint64(addr) == object.Address {
					fmt.Printf("Found object in data segment\n")
					return
				}
			}

			// Check bss
			for _, object := range heapFile.BSS().Objects() {
				if uint64(addr) == object.Address {
					fmt.Printf("Found object in bss\n")
					return
				}
			}
		},
	}
	gohatCmd.AddCommand(containsCommand)

	var objectAsString bool
	var objectBinary bool
	var objectCommand = &cobra.Command{
		Use:   "object",
		Short: "Dump the contents of an object",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			if len(args) != 2 {
				fmt.Println("object <heap file> <address>")
				os.Exit(1)
			}

			addr, _ := strconv.ParseInt(args[1], 16, 64)
			object := heapFile.Object(addr)
			if object == nil {
				fmt.Println("Could not find object")
				return
			}

			if objectBinary {
				os.Stdout.Write([]byte(object.Content))
				return
			}

			// os.Stdout.Write([]byte(object.Content))
			// fmt.Println([]byte(object.Content))

			fmt.Printf("%016x %s %d %d\n", object.Address, object.Kind(), object.Size, len(object.Content))
			if object.Type != nil {
				fmt.Println(object.Type.Name)
			}

			if object.Type != nil && object.Type.Name == "string" {
				val := derefToString([]byte(object.Content), heapFile)
				if val != "" {
					fmt.Printf("Value: %s\n", val)
				}
			}
			fmt.Println("")
			if objectAsString {
				fmt.Printf("%v\n", []byte(object.Content))
			}
			// } else {
			// 	fmt.Println([]byte(object.Content))
			// }
			fmt.Println("")

			if object.Type != nil {
				fmt.Println("Field List:")
				var lastOffset uint64
				for idx, field := range object.Type.FieldList {
					if idx == len(object.Type.FieldList)-1 {
						data := []byte(object.Content)[lastOffset:]
						switch field.KindString() {
						case "Ptr   ":
							var ptraddr int64
							buf := bytes.NewReader(data)
							binary.Read(buf, binary.LittleEndian, &ptraddr)
							fmt.Printf("%s 0x%016x  %x\n", field.KindString(), field.Offset, ptraddr)
						case "String":
							// val := derefToString(data, heapFile)
							fmt.Printf("%s 0x%016x  \n", field.KindString(), field.Offset)
						default:
							fmt.Printf("%s 0x%016x  \n", field.KindString(), field.Offset)
						}
					} else {
						lastOffset = object.Type.FieldList[idx].Offset
						nextOffset := object.Type.FieldList[idx+1].Offset
						data := []byte(object.Content)[lastOffset:nextOffset]
						switch field.Kind {
						case heapfile.FieldPtr:
							var ptraddr int64
							buf := bytes.NewReader(data)
							binary.Read(buf, binary.LittleEndian, &ptraddr)
							fmt.Printf("%s 0x%016x  %x\n", field.KindString(), field.Offset, ptraddr)
						case heapfile.FieldStr:
							// val := derefToString(data, heapFile)
							fmt.Printf("%s 0x%016x  \n", field.KindString(), field.Offset)
						default:
							fmt.Printf("%s 0x%016x  \n", field.KindString(), field.Offset)
						}
					}
				}
			}

			fmt.Println("Children")
			for _, child := range object.Children() {
				fmt.Printf("%x\n", child.Address)
			}
		},
	}
	objectCommand.Flags().BoolVarP(&objectAsString, "string", "s", false, "Show object contents as a string")
	objectCommand.Flags().BoolVarP(&objectBinary, "binary", "b", false, "Dump the binary contents of the object")
	gohatCmd.AddCommand(objectCommand)

	var objectsCommand = &cobra.Command{
		Use:   "objects",
		Short: "Dump a list of objects",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			objects := heapFile.Objects()
			for _, object := range objects {
				typeName := "<unknown>"
				if object.Type != nil {
					typeName = object.Type.Name
				}
				fmt.Printf("%016x,%s,%s,%d\n", object.Address, typeName, object.Kind(), object.Size)
			}
		},
	}
	gohatCmd.AddCommand(objectsCommand)

	var paramsCommand = &cobra.Command{
		Use:   "params",
		Short: "Show the heap parameters",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

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
	gohatCmd.AddCommand(paramsCommand)

	var rootsCommand = &cobra.Command{
		Use:   "roots",
		Short: "dump roots",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			for _, root := range heapFile.Roots() {
				fmt.Printf("%x %s\n", root.Pointer, root.Description)
			}
		},
	}
	gohatCmd.AddCommand(rootsCommand)

	var fragmentCommand = &cobra.Command{
		Use:   "fragment",
		Short: "show unused address locations",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			totalFrag := uint64(0)

			objects := heapFile.Objects()
			addresses := make(uint64arr, 0, len(objects))
			for _, obj := range objects {
				addresses = append(addresses, obj.Address)
			}
			sort.Sort(addresses)

			firstAddr := addresses[0]
			lastAddr := addresses[len(addresses)-1]
			lastObject := heapFile.Object(int64(lastAddr))
			// lastAddr += uint64(heapFile.Object(int64(lastAddr)).Size)

			for i := 0; i < len(addresses)-1; i++ {
				addr := addresses[i]
				nextAddr := addresses[i+1]
				object := heapFile.Object(int64(addr))
				size := object.Size
				endAddr := addr + uint64(size)

				// fmt.Printf("%x %x %d (%x)\n", addr, nextAddr, size, endAddr)
				if endAddr != nextAddr {
					fragAmount := nextAddr - endAddr
					totalFrag += fragAmount
					fmt.Printf("%x - %x  (%d)\n", endAddr, nextAddr, fragAmount)
				}
			}

			// May be junk on the end
			params := heapFile.DumpParams()
			endAddr := lastAddr + uint64(lastObject.Size)
			// fmt.Printf("last addr: %x, endAddr: %x\n", lastAddr, endAddr)
			endCruft := params.EndAddress - endAddr
			if endCruft > 0 {
				totalFrag += endCruft
				fmt.Printf("%x - %x  (%d)\n", endAddr, params.EndAddress, endCruft)
			}

			fmt.Printf("Total bytes fragmented between %x and %x: %d\n", firstAddr, params.EndAddress, totalFrag)
		},
	}
	gohatCmd.AddCommand(fragmentCommand)

	type sameObject struct {
		Object      *heapfile.Object
		SameContent bool
	}
	var sameCommand = &cobra.Command{
		Use:   "same",
		Short: "find objects that are the same in two heap files",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile1 := verifyHeapDumpFile(args)

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

			same := make([]*sameObject, 0, len(heapObjects1))

			for _, obj := range heapObjects1 {
				if cmp := heapFile2.Object(int64(obj.Address)); cmp != nil {
					if cmp.TypeAddress == obj.TypeAddress &&
						cmp.Kind() == obj.Kind() &&
						// cmp.Content == obj.Content &&
						cmp.Size == obj.Size {
						same = append(same, &sameObject{obj, cmp.Content == obj.Content})
					}
				}
			}

			for _, same := range same {
				object := same.Object
				typeName := "unknown"
				if object.Type != nil {
					typeName = object.Type.Name
				}

				fmt.Printf("%x,%s,%d,%v\n", object.Address, typeName, object.Size, same.SameContent)
				//
				// fmt.Printf("%016x %s %d %d\n", object.Address, object.Kind(), object.Size, len(object.Content))
				// if object.Type != nil {
				// 	fmt.Println(object.Type.Name)
				// }
				// fmt.Println("")
				// if objectAsString {
				// 	fmt.Println(object.Content)
				// } else {
				// 	fmt.Println([]byte(object.Content))
				// }
				// fmt.Println("")
				//
				// if object.Type != nil {
				// 	fmt.Println("Field List:")
				// 	var lastOffset uint64
				// 	for idx, field := range object.Type.FieldList {
				// 		if idx == len(object.Type.FieldList)-1 {
				// 			fmt.Printf("%s 0x%016x  %v\n", field.Kind(), field.Offset, []byte(object.Content[lastOffset:]))
				// 		} else {
				// 			lastOffset = object.Type.FieldList[idx].Offset
				// 			nextOffset := object.Type.FieldList[idx+1].Offset
				// 			fmt.Printf("%s 0x%016x  %v\n", field.Kind(), field.Offset, []byte(object.Content[lastOffset:nextOffset]))
				// 		}
				// 	}
				// }
				// fmt.Printf("\n\n")
			}
		},
	}
	gohatCmd.AddCommand(sameCommand)

	var stackFramesCommand = &cobra.Command{
		Use:   "stackframes",
		Short: "Dump the stack frames",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			for _, frame := range heapFile.StackFrames() {
				fmt.Printf("%x %s\n", frame.StackPointer, frame.Name)
				for _, object := range frame.Objects() {
					fmt.Printf("\t%x\n", object.Address)
					for _, child := range object.Children() {
						fmt.Printf("\t\t%x\n", child.Address)
					}
				}
			}
		},
	}
	gohatCmd.AddCommand(stackFramesCommand)

	var garbageCommand = &cobra.Command{
		Use:   "garbage",
		Short: "Dump unreachable objects",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			trash := garbage(heapFile)
			fmt.Printf("Found %d unreachable objects\n", len(trash))
			for _, object := range trash {
				typeName := "unknown"
				if object.Type != nil {
					typeName = object.Type.Name
				}
				fmt.Printf("%x %s\n", object.Address, typeName)
			}
		},
	}
	gohatCmd.AddCommand(garbageCommand)

	var serverAddress string
	var serverCommand = &cobra.Command{
		Use:   "server",
		Short: "run the web interface",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)
			s := newGohatServer(serverAddress, heapFile)
			s.Run()
		},
	}
	gohatCmd.AddCommand(serverCommand)
	serverCommand.Flags().StringVarP(&serverAddress, "addr", "a", ":5150", "Address to listen on (default :5150)")

	var typeCommand = &cobra.Command{
		Use:   "type",
		Short: "Dump information about a type",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

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

	var typesCommand = &cobra.Command{
		Use:   "types",
		Short: "Dump all the types found in the heap",
		Run: func(cmd *cobra.Command, args []string) {
			heapFile := verifyHeapDumpFile(args)

			types := heapFile.Types()
			for _, t := range types {
				fmt.Printf("%016x %d %s\n", t.Address, len(t.FieldList), t.Name)
			}
		},
	}
	gohatCmd.AddCommand(typesCommand)

	gohatCmd.Execute()
}

func verifyHeapDumpFile(args []string) *heapfile.HeapFile {
	if len(args) < 1 {
		fmt.Println("heap file required")
		os.Exit(1)
	}
	heapFile, err := heapfile.New(args[0])
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	return heapFile
}

func derefToString(b []byte, heapFile *heapfile.HeapFile) string {
	var len int64
	var addr int64
	buf := bytes.NewReader(b[8:])
	binary.Read(buf, binary.LittleEndian, &len)
	buf = bytes.NewReader(b[:8])
	binary.Read(buf, binary.LittleEndian, &addr)
	if obj := heapFile.Object(addr); obj != nil {
		return obj.Content
	}
	return ""
}

type HistSorter struct {
	Keys []string
	Vals []int
}

func NewHistSorter(m map[string]int) *HistSorter {
	hs := &HistSorter{make([]string, 0, len(m)), make([]int, 0, len(m))}
	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

func (hs *HistSorter) Sort() {
	sort.Sort(hs)
}

func (hs *HistSorter) Len() int           { return len(hs.Vals) }
func (hs *HistSorter) Less(i, j int) bool { return hs.Vals[i] < hs.Vals[j] }
func (hs *HistSorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}

type uint64arr []uint64

func (a uint64arr) Len() int           { return len(a) }
func (a uint64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a uint64arr) Less(i, j int) bool { return a[i] < a[j] }

func garbage(heapFile *heapfile.HeapFile) []*heapfile.Object {
	objects := heapFile.Objects()
	seen := make(map[uint64]bool, len(objects))

	for _, object := range objects {
		seen[object.Address] = false
	}

	// Mark all the objects the stack frames (roots) point to
	for _, frame := range heapFile.StackFrames() {
		for _, object := range frame.Objects() {
			mark(object, &seen)
		}
	}

	// other roots
	for _, root := range heapFile.Roots() {
		if object := heapFile.Object(int64(root.Pointer)); object != nil {
			mark(object, &seen)
		}
	}

	// data segment
	for _, object := range heapFile.DataSegment().Objects() {
		mark(object, &seen)
	}

	// bss
	for _, object := range heapFile.BSS().Objects() {
		mark(object, &seen)
	}

	// finalizers
	for _, f := range heapFile.QueuedFinalizers() {
		o := heapFile.Object(int64(f.ObjectAddress))
		if o != nil {
			mark(o, &seen)

		}
	}
	for _, f := range heapFile.Finalizers() {
		o := heapFile.Object(int64(f.ObjectAddress))
		if o != nil {
			mark(o, &seen)

		}
	}

	trash := make([]*heapfile.Object, 0, len(objects))
	for addr, visited := range seen {
		if !visited {
			trash = append(trash, heapFile.Object(int64(addr)))
		}
	}

	return trash
}

func mark(object *heapfile.Object, seen *map[uint64]bool) {
	if seen := (*seen)[object.Address]; seen {
		return
	}

	(*seen)[object.Address] = true
	for _, child := range object.Children() {
		mark(child, seen)
	}
}

func dumpObject(heapFile *heapfile.HeapFile, objectAddr uint64, content []byte, depth int) {
	tabs := depth + 1
	params := heapFile.DumpParams()
	var addr uint64
	var lastIndex uint64 = 0
	contentLength := uint64(len(content))
	for i := params.PtrSize; i < contentLength+params.PtrSize; i += params.PtrSize {
		buf := bytes.NewReader([]byte(content[lastIndex:i]))
		binary.Read(buf, binary.LittleEndian, &addr)

		if addr >= params.StartAddress && addr <= params.EndAddress {
			if obj := heapFile.Object(int64(addr)); obj != nil {
				if addr == objectAddr {
					continue
				}
				for i := 0; i < tabs; i++ {
					fmt.Print(" ")
				}
				objectType := "<unknown>"
				if obj.Type != nil {
					objectType = obj.Type.Name
				}
				fmt.Printf("%x %s\n", addr, objectType)
				dumpObject(heapFile, obj.Address, []byte(obj.Content), tabs)
			}
		}

		lastIndex = i
	}
}
