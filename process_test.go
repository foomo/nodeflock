package nodeflock

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"testing"
	"time"
)

// __DIR__
func GetCurrentDir() string {
	_, filename, _, _ := runtime.Caller(1)
	return path.Dir(filename)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
func TestProcessSimple(t *testing.T) {
	var channel = make(chan int)
	sourceFile := path.Join(GetCurrentDir(), "nodeflock-process", "demo.js")
	sourceFileInfo, _ := os.Stat(sourceFile)

	p, err := newProcess(sourceFile, sourceFileInfo.ModTime().UnixNano(), 0, channel)
	panicOnErr(err)
	testName := "Hänsi"
	for i := 0; i < 10000; i++ {
		funcResult := ""
		result, err := p.callJS(&funcResult, "Demo.Simple", fmt.Sprint(testName, " ", i))
		panicOnErr(err)
		if len(result.Error) > 0 {
			t.Fatal("unexpected nodejs error", result.Error)
		}
		if funcResult != fmt.Sprint("simple name ", testName, " ", i) {
			t.Fatal("unexpected result from js call")
		}
	}
}

type complexResult struct {
	Size int
	Date time.Time
}

// func TestProcessComplex(t *testing.T) {
// 	funcResult := &complexResult{}
// 	p, err := newProcess(path.Join(GetCurrentDir(), "nodeflock-process", "demo.js"))
// 	panicOnErr(err)
// 	result, err := p.callJS(funcResult, "Complex", 31)
// 	panicOnErr(err)
// 	t.Log(result, funcResult.Date, funcResult.Size)

// }
