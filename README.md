# gohat

Gohat is a heap dump analyzer tool for Go heap dumps written with `runtime.WriteHeapDump()`

### Show the heap dump params
```
$ gohat params dumpfile.dump
Big Endian
Pointer Size: 8
Channel Header Size: 88
Heap Starting Address 2081a4000
Heap Ending Address: 2082a4000
Architecture: 54
GOEXPERIMENT:
nCPU: 8
```

### Show the memstats at the time of the dump
```
$ gohat params dumpfile.dump
Big Endian
Pointer Size: 8
Channel Header Size: 88
Heap Starting Address 2081a4000
Heap Ending Address: 2082a4000
Architecture: 54
GOEXPERIMENT:
nCPU: 8
$ gohat memstats dumpfile.dump
General statistics
Alloc: 174864
TotalAlloc: 174880
Sys: 4327672
Lookups: 6
Mallocs: 110
Frees: 1

Main allocation heap statistics
HeapAlloc: 174864
HeapSys: 1048576
HeapIdle: 737280
HeapInuse: 311296
HeapReleased: 0
HeapObjects: 109

Low-level fixed-size structure allocator statistics
StackInuse: 57344
StatckSys: 393216
MSpanInuse: 2944
MSpanSys: 16384
BuckHashSys: 1440424
GCSys: 1114112
OtherSys: 298576

Garbage collector statistics
NextGC: 160256
LastGC: 1408810400718324880
PauseTotalNs: 104638
NumGC: 1
Last GC Pauses:
[104638 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
```

### List all of the objects on the heap
```
$ gohat objects dumpfile.dump
00000002081c81e0 map.hdr[string]*unicode.RangeTable regular 48
00000002081de000 os.File regular 8
00000002081b2000 <unknown> conservatively scanned 1664
00000002081ce9c0 map.bucket[string]*unicode.RangeTable array 208
00000002081a6000 runtime.g regular 288
00000002081a8000 <unknown> regular 96
00000002081d2000 map.bucket[string]int array 416
00000002081c8150 map.hdr[string]*unicode.RangeTable regular 48
00000002081cc000 <unknown> regular 576
00000002081a4070 errors.errorString regular 16
000000c208499df0 string regular 16
...
```

### Show a single object
```
$ gohat object dumpfile.dump 000000c208499df0
000000c208499df0 regular 16 16
string

[128 8 73 8 194 0 0 0 54 0 0 0 0 0 0 0]

Field List:
String 0x0000000000000000  [128 8 73 8 194 0 0 0 54 0 0 0 0 0 0 0]
```

```
$ gohat object dumpfile.dump 000000c208490880 --string
000000c208490880 regular 64 64

TÜRKTRUST Elektronik Sertifika Hizmet Sağlayıcısı
```
