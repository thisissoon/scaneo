package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

var (
	testFiles = []string{
		"testdata/declarations.go",
		"testdata/methods.go",
		"testdata/types.go",
		"testdata/visibility.go",
	}

	testFilesLen = len(testFiles)

	fileStructsMap = map[string][]structToken{
		"testdata/visibility.go": []structToken{
			{
				Name: "Exported",
				Fields: []fieldToken{
					{Name: "A", Type: "int"},
					{Name: "B", Type: "int"},
				},
			},
			{
				Name: "unexported",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "int"},
				},
			},
			{
				Name: "ExAndUn",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "int"},
				},
			},
			{
				Name: "unAndEx",
				Fields: []fieldToken{
					{Name: "A", Type: "int"},
					{Name: "B", Type: "int"},
				},
			},
		},
		"testdata/declarations.go": []structToken{
			{
				Name: "t0",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "bool"},
				},
			},
			{
				Name: "t1",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "bool"},
				},
			},
			{
				Name: "t2",
				Fields: []fieldToken{
					{Name: "a", Type: "string"},
					{Name: "b", Type: "byte"},
				},
			},
			{
				Name: "t3",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "int"},
					{Name: "c", Type: "int"},
					{Name: "d", Type: "bool"},
					{Name: "e", Type: "bool"},
					{Name: "f", Type: "bool"},
				},
			},
			{
				Name: "t4",
				Fields: []fieldToken{
					{Name: "a", Type: "int"},
					{Name: "b", Type: "bool"},
				},
			},
		},
		"testdata/methods.go": []structToken{
			{
				Name: "Post",
				Fields: []fieldToken{
					{Name: "ID", Type: "int"},
					{Name: "SemURL", Type: "string"},
					{Name: "Created", Type: "time.Time"},
					{Name: "Modified", Type: "time.Time"},
					{Name: "Published", Type: "pq.NullTime"},
					{Name: "Draft", Type: "bool"},
					{Name: "Title", Type: "string"},
					{Name: "Body", Type: "string"},
				},
			},
		},
		"testdata/types.go": []structToken{
			{
				Name: "boolean",
				Fields: []fieldToken{
					{Name: "a", Type: "bool"},
				},
			},
			{
				Name: "numerics",
				Fields: []fieldToken{
					{Name: "a", Type: "uint8"},
					{Name: "b", Type: "uint16"},
					{Name: "c", Type: "uint32"},
					{Name: "d", Type: "uint64"},
					{Name: "e", Type: "int8"},
					{Name: "f", Type: "int16"},
					{Name: "g", Type: "int32"},
					{Name: "h", Type: "int64"},
					{Name: "i", Type: "float32"},
					{Name: "j", Type: "float64"},
					{Name: "k", Type: "complex64"},
					{Name: "l", Type: "complex128"},
					{Name: "m", Type: "byte"},
					{Name: "n", Type: "rune"},
					{Name: "o", Type: "uint"},
					{Name: "p", Type: "int"},
					{Name: "q", Type: "uintptr"},
				},
			},
			{
				Name: "str",
				Fields: []fieldToken{
					{Name: "a", Type: "string"},
				},
			},
			{
				Name: "structs",
				Fields: []fieldToken{
					{Name: "a", Type: "sql.NullString"},
				},
			},
			{
				Name: "slices",
				Fields: []fieldToken{
					{Name: "a", Type: "[]bool"},
					{Name: "b", Type: "[]time.Time"},
					{Name: "c", Type: "[]*byte"},
					{Name: "d", Type: "[]*sql.NullString"},
				},
			},
			{
				Name: "pointers",
				Fields: []fieldToken{
					{Name: "a", Type: "*bool"},
					{Name: "b", Type: "*time.Time"},
					{Name: "c", Type: "*[]byte"},
					{Name: "d", Type: "*[]sql.NullString"},
				},
			},
		},
	}
)

func TestFindFiles(t *testing.T) {
	var noPaths []string
	files, err := findFiles(noPaths)
	if err == nil {
		t.Error("no file paths passed")
		t.Error("should be error")
		t.FailNow()
	}

	badPaths := []string{"doesnt/exist", "not/here.txt"}
	files, err = findFiles(badPaths)
	if err == nil {
		t.Error("passed non-existent file paths")
		t.Error("should be error")
		t.FailNow()
	}

	inputPaths := []string{"testdata/", testFiles[3]}
	files, err = findFiles(inputPaths)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if testFilesLen != len(files) {
		t.Error("unexpected file count")
		t.Errorf("expected: %d; found: %d\n", testFilesLen, len(files))
		t.FailNow()
	}

	sort.Strings(files)

	for i := range files {
		filename := filepath.Base(files[i])
		testFilename := filepath.Base(testFiles[i])

		if testFilename != filename {
			t.Error("unexpected filename")
			t.Errorf("expected: %s; found: %s\n", testFilename, filename)
			t.Error("files:", files)
			t.Error("testFiles:", testFiles)
			t.FailNow()
		}
	}
}

func TestWhitelist(t *testing.T) {
	whitelist := "Exported,unexported"
	expectedToks := 2

	toks, err := parseCode(testFiles[3], whitelist)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if expectedToks != len(toks) {
		t.Error("unexpected struct tokens length")
		t.Errorf("expected: %d; found: %d\n", expectedToks, len(toks))
	}
}

func TestParseCode(t *testing.T) {
	var noFilter string

	var noSource string
	if _, err := parseCode(noSource, noFilter); err == nil {
		t.Error("no source file path passed")
		t.Error("should be error")
		t.FailNow()
	}

	for fPath, structToks := range fileStructsMap {
		// get all struct tokens for a given file
		toks, err := parseCode(fPath, noFilter)
		if err != nil {
			t.Error(err)
			t.FailNow()
		}

		if len(structToks) != len(toks) {
			t.Error("unexpected struct tokens length")
			t.Errorf("expected: %d; found: %d\n", len(structToks), len(toks))
			t.FailNow()
		}

		for i := range toks {
			if structToks[i].Name != toks[i].Name {
				t.Error("unexpected struct name")
				t.Errorf("expected: %s; found: %s\n", structToks[i].Name, toks[i].Name)
				t.FailNow()
			}

			if len(structToks[i].Fields) != len(toks[i].Fields) {
				t.Error("unexpected struct fields length")
				t.Error("file:", fPath)
				t.Error("struct:", structToks[i].Name)
				t.Errorf("expected: %d; found: %d\n", len(structToks[i].Fields), len(toks[i].Fields))
				t.Error("expected:", structToks[i].Fields)
				t.Error("found:", toks[i].Fields)
				t.FailNow()
			}

			for j := range toks[i].Fields {
				if structToks[i].Fields[j].Name != toks[i].Fields[j].Name {
					t.Error("unexpected struct field name")
					t.Error("file:", fPath)
					t.Error("struct:", structToks[i].Name)
					t.Errorf("expected: %s; found: %s\n", structToks[i].Fields[j].Name, toks[i].Fields[j].Name)
					t.FailNow()
				}

				if structToks[i].Fields[j].Type != toks[i].Fields[j].Type {
					t.Error("unexpected struct field type")
					t.Error("file:", fPath)
					t.Error("struct:", structToks[i].Name)
					t.Error("field:", structToks[i].Fields[j].Name)
					t.Errorf("expected: %s; found: %s\n", structToks[i].Fields[j].Type, toks[i].Fields[j].Type)
					t.FailNow()
				}
			}
		}
	}
}

func TestGenFile(t *testing.T) {

	toks := fileStructsMap[testFiles[3]][:2]

	expectedFuncNames := []string{
		"scanExported",
		"scanExporteds",
		"scanUnexported",
		"scanUnexporteds",
	}

	tt := []struct {
		name          string
		outFile       bool
		tokens        []structToken
		unexport      bool
		funcs         bool
		pkgImport     string
		assert        func(*testing.T, error)
		expectedFuncs []string
	}{
		{
			"no tokens",
			true,
			[]structToken{},
			true,
			false,
			"",
			func(t *testing.T, err error) {
				if err == nil {
					t.Error("no struct tokens passed")
					t.Error("should be error")
					t.FailNow()
				}
			},
			expectedFuncNames,
		},
		{
			"no output file",
			false,
			toks,
			true,
			false,
			"",
			func(t *testing.T, err error) {
				if err == nil {
					t.Error("no output file path passed")
					t.Error("should be error")
					t.FailNow()
				}
			},
			expectedFuncNames,
		},
		{
			"scan funcs unexported",
			true,
			toks,
			true,
			false,
			"",
			func(t *testing.T, err error) {
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			expectedFuncNames,
		},
		{
			"sql helper funcs",
			true,
			toks,
			true,
			true,
			"",
			func(t *testing.T, err error) {
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			[]string{
				"scanExported",
				"scanExporteds",
				"SelectExported",
				"SelectExporteds",
				"sliceExported",
				"InsertExported",
				"UpdateExported",
				"scanUnexported",
				"scanUnexporteds",
				"SelectUnexported",
				"SelectUnexporteds",
				"sliceUnexported",
				"InsertUnexported",
				"UpdateUnexported",
			},
		},
		{
			"pkg import",
			true,
			toks,
			true,
			false,
			"testsvc/storage/user",
			func(t *testing.T, err error) {
				if err != nil {
					t.Error(err)
					t.FailNow()
				}
			},
			expectedFuncNames,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var outFile string
			if tc.outFile {
				outFile = filepath.Join(
					os.TempDir(),
					fmt.Sprintf("scaneo-test-%d", time.Now().UnixNano()))
			}
			// genFile(file, package, unexport, tokens, funcs)
			err := genFile(outFile, "testing", tc.unexport, tc.tokens, tc.funcs, tc.pkgImport)
			defer os.Remove(outFile) // comment this line to examine generated code

			tc.assert(t, err)
			if err != nil {
				return
			}

			fset := token.NewFileSet()
			astf, err := parser.ParseFile(fset, outFile, nil, 0)
			if err != nil {
				t.Error(err)
				t.FailNow()
			}

			scanFuncs := make([]string, 0, len(toks))
			for _, dec := range astf.Decls {
				funcDecl, isFuncDecl := dec.(*ast.FuncDecl)
				if !isFuncDecl {
					continue
				}

				scanFuncs = append(scanFuncs, funcDecl.Name.String())
			}

			if len(tc.expectedFuncs) != len(scanFuncs) {
				t.Error("unexpected number of scan functions found")
				t.Errorf("expected: %d; found: %d\n", len(tc.expectedFuncs), len(scanFuncs))
				t.FailNow()
			}

			for i := range tc.expectedFuncs {
				if tc.expectedFuncs[i] != scanFuncs[i] {
					t.Error("unexpected scan function found")
					t.Errorf("expected: %s; found: %s\n", tc.expectedFuncs[i], scanFuncs[i])
				}
			}
		})
	}
}
