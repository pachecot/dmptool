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
	destDir  string
	withType bool
}

func (s *peHandler) Code(code *dmp.Code) {

	file := filepath.Join(s.destDir, code.Path+".pe")
	if s.withType {
		file = filepath.Join(filepath.Dir(file), code.Type, filepath.Base(file))
	}

	err := os.MkdirAll(filepath.Dir(file), os.ModeDir)
	if err != nil {
		fmt.Println("error creating directory: ", err)
		return
	}

	os.WriteFile(file, []byte(strings.Join(code.Lines, "\r\n")), os.ModePerm)
	if !code.Modified.IsZero() {
		os.Chtimes(file, code.Modified, code.Modified)
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
