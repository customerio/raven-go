package raven

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

// a
func trace() *Stacktrace {
	return NewStacktrace(0, 2, []string{thisPackage})
	// b
}

func init() {
	thisFile, thisPackage = derivePackage()
	functionNameTests = []FunctionNameTest{
		{0, thisPackage, "TestFunctionName"},
		{1, "testing", "tRunner"},
		{2, "runtime", "goexit"},
		{100, "", ""},
	}
}

type FunctionNameTest struct {
	skip int
	pack string
	name string
}

var (
	thisFile          string
	thisPackage       string
	functionNameTests []FunctionNameTest
)

func TestFunctionName(t *testing.T) {
	for _, test := range functionNameTests {
		pc, _, _, _ := runtime.Caller(test.skip)
		pack, name := functionName(runtime.FuncForPC(pc).Name())

		if pack != test.pack {
			t.Errorf("incorrect package; got %s, want %s", pack, test.pack)
		}
		if name != test.name {
			t.Errorf("incorrect function; got %s, want %s", name, test.name)
		}
	}
}

func TestStacktrace(t *testing.T) {
	st := trace()
	if st == nil || st.Frames == nil {
		t.Error("got nil stacktrace")
	}
	if len(st.Frames) == 0 {
		t.Error("got zero frames")
	}
}

func TestStacktraceFrame(t *testing.T) {
	st := trace()
	f := st.Frames[len(st.Frames)-1]
	_, filename, _, _ := runtime.Caller(0)
	runningInVendored := strings.Contains(filename, "vendor")

	if f.Filename != thisFile {
		t.Errorf("incorrect Filename; got %s, want %s", f.Filename, thisFile)
	}
	if !strings.HasSuffix(f.AbsolutePath, thisFile) {
		t.Error("incorrect AbsolutePath:", f.AbsolutePath)
	}
	if f.Function != "trace" {
		t.Error("incorrect Function:", f.Function)
	}
	if f.Module != thisPackage {
		t.Error("incorrect Module:", f.Module)
	}
	if f.Lineno != 16 {
		t.Error("incorrect Lineno:", f.Lineno)
	}
	if f.InApp != !runningInVendored {
		t.Error("expected InApp to be true")
	}
	if f.InApp && st.Culprit() != fmt.Sprintf("%s.trace", thisPackage) {
		t.Error("incorrect Culprit:", st.Culprit())
	}
}

func TestStacktraceContext(t *testing.T) {
	st := trace()
	f := st.Frames[len(st.Frames)-1]
	if f.ContextLine != "\treturn NewStacktrace(0, 2, []string{thisPackage})" {
		t.Errorf("incorrect ContextLine: %#v", f.ContextLine)
	}
	if len(f.PreContext) != 2 || f.PreContext[0] != "// a" || f.PreContext[1] != "func trace() *Stacktrace {" {
		t.Errorf("incorrect PreContext %#v", f.PreContext)
	}
	if len(f.PostContext) != 2 || f.PostContext[0] != "\t// b" || f.PostContext[1] != "}" {
		t.Errorf("incorrect PostContext %#v", f.PostContext)
	}
}

func traceErrorWithStack() error {
	return errors.WithStack(errors.New("oops"))
}

func TestStacktraceErrorsWithStack(t *testing.T) {
	err := traceErrorWithStack()
	st := GetOrNewStacktrace(err, 0, 0, []string{thisPackage})

	if st == nil || st.Frames == nil {
		t.Error("failed to get stacktrace from errors.WithStack() error")
	}
	if len(st.Frames) != 3 {
		t.Error("unexpected number of stack frames")
	}
	if st.Frames[0] == nil {
		t.Error("frame 0 is nil")
	}

	// 0: tRunner
	f := st.Frames[0]
	if f.Function != "tRunner" {
		t.Error("incorrect Function", f.Function)
	}

	// 1: TestStacktraceErrorsWithStack (this function)
	f = st.Frames[1]
	if f.Function != "TestStacktraceErrorsWithStack" {
		t.Error("incorrect Function", f.Function)
	}
	if f.Module != thisPackage {
		t.Error("incorrect Module:", f.Module)
	}

	// 2: traceErrorWithStack
	f = st.Frames[2]
	if f.Function != "traceErrorWithStack" {
		t.Error("incorrect Function", f.Function)
	}
	if f.Module != thisPackage {
		t.Error("incorrect Module:", f.Module)
	}

	// for i, frame := range st.Frames {
	// 	fmt.Fprintf(os.Stderr, "[%d]: %+v\n", i, frame)
	// }
}

func derivePackage() (file, pack string) {
	// Get file name by seeking caller's file name.
	_, callerFile, _, ok := runtime.Caller(1)
	if !ok {
		return
	}

	// Trim file name
	file = callerFile
	for _, dir := range build.Default.SrcDirs() {
		dir := dir + string(filepath.Separator)
		if trimmed := strings.TrimPrefix(callerFile, dir); len(trimmed) < len(file) {
			file = trimmed
		}
	}

	// Now derive package name
	dir := filepath.Dir(callerFile)

	dirPkg, err := build.ImportDir(dir, build.AllowBinary)
	if err != nil {
		return
	}

	pack = dirPkg.ImportPath
	return
}

// TestNewStacktrace_outOfBounds verifies that a context exceeding the number
// of lines in a file does not cause a panic.
func TestNewStacktrace_outOfBounds(t *testing.T) {
	st := NewStacktrace(0, 1000000, []string{thisPackage})
	f := st.Frames[len(st.Frames)-1]
	if f.ContextLine != "\tst := NewStacktrace(0, 1000000, []string{thisPackage})" {
		t.Errorf("incorrect ContextLine: %#v", f.ContextLine)
	}
}

func TestNewStacktrace_noFrames(t *testing.T) {
	st := NewStacktrace(999999999, 0, []string{})
	if st != nil {
		t.Errorf("expected st.Frames to be nil: %v", st)
	}
}

func TestFileContext(t *testing.T) {
	// reset the cache
	loader := &fsLoader{cache: make(map[string][][]byte)}

	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal("failed to create temporary directory:", err)
	}

	defer func() {
		err := os.RemoveAll(tempdir)
		if err != nil {
			fmt.Println("failed to remove temporary directory:", err)
		}
	}()

	okPath := filepath.Join(tempdir, "ok")
	missingPath := filepath.Join(tempdir, "missing")
	noPermissionPath := filepath.Join(tempdir, "noperms")

	err = ioutil.WriteFile(okPath, []byte("hello\nworld\n"), 0600)
	if err != nil {
		t.Fatal("failed writing file:", err)
	}
	err = ioutil.WriteFile(noPermissionPath, []byte("no access\n"), 0000)
	if err != nil {
		t.Fatal("failed writing file:", err)
	}

	tests := []struct {
		path          string
		expectedLines int
		expectedIndex int
	}{
		{okPath, 1, 0},
		{missingPath, 0, 0},
		{noPermissionPath, 0, 0},
	}
	for i, test := range tests {
		lines, index := loader.Load(test.path, 1, 0)
		if !(len(lines) == test.expectedLines && index == test.expectedIndex) {
			t.Errorf("%d: fileContext(%#v, 1, 0) = %v, %v; expected len()=%d, %d",
				i, test.path, lines, index, test.expectedLines, test.expectedIndex)
		}
		cacheLen := len(loader.cache)
		if cacheLen != i+1 {
			t.Errorf("%d: result was not cached; len=%d", i, cacheLen)
		}
	}
}
