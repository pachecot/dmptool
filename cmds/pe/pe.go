package pe

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tpacheco/dmptool/dmp"
)

type peHandler struct {
	dmp.EmptyHandler
	destDir  string
	withType bool
}

func isCodeType(typeName string) bool {
	switch typeName {
	case
		"Program",
		"InfinityProgram",
		"InfinityFunction":
		// ok
		return true
	default:
		return false
	}
}

func (s *peHandler) handleCode(obj *dmp.Object) {

	code, ok := obj.Properties["ByteCode"]
	if !ok {
		return
	}

	file := filepath.Join(s.destDir, obj.Path+".pe")
	if s.withType {
		file = filepath.Join(filepath.Dir(file), obj.Type, filepath.Base(file))
	}

	dirName := filepath.Dir(file)
	err := os.MkdirAll(dirName, os.ModeDir)
	if err != nil {
		fmt.Printf("error creating directory %s: %e\n", dirName, err)
		return
	}

	os.WriteFile(file, []byte(code), os.ModePerm)

	if mod, ok := obj.Properties["Modified"]; ok {
		modTime, err := dmp.ParseTime(mod)
		if err == nil {
			os.Chtimes(file, modTime, modTime)
		}
	}
}

func (s *peHandler) Object(obj *dmp.Object) {
	if isCodeType(obj.Type) {
		s.handleCode(obj)
	}
}

type Command struct {
	FileName   string
	OutDir     string
	TypeFolder bool
}

func (cmd *Command) Execute() {

	h := &peHandler{
		destDir:  cmd.OutDir,
		withType: cmd.TypeFolder,
	}

	dmp.ParseFile(cmd.FileName, h)
}
