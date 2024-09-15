package pe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
)

type peHandler struct {
	dmp.EmptyHandler
	destDir    string
	flatten    bool
	separator  string
	withType   bool
	typePrefix string
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

func (s *peHandler) filePath(obj *dmp.Object) string {

	fileName := obj.Path + ".pe"

	if s.withType {
		dir, name := filepath.Split(fileName)
		fileName = filepath.Join(dir, s.typePrefix+obj.Type, name)
	}

	if s.flatten {
		fileName = strings.Join(
			filepath.SplitList(fileName),
			s.separator,
		)
	}

	return filepath.Join(s.destDir, fileName)
}

func (s *peHandler) handleCode(obj *dmp.Object) {

	code, ok := obj.Properties["ByteCode"]
	if !ok {
		return
	}

	file := s.filePath(obj)
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
	Flatten    bool
	FlattenSep string
	TypePrefix string
}

func (cmd *Command) Execute() {

	h := &peHandler{
		destDir:    cmd.OutDir,
		withType:   cmd.TypeFolder,
		flatten:    cmd.Flatten,
		typePrefix: cmd.TypePrefix,
		separator:  cmd.FlattenSep,
	}

	dmp.ParseFile(cmd.FileName, h)
}
