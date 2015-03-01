package ace

// github.com/yosssi/gcss binding for slurp.
// No Configuration required.

import (
	"bytes"
	html "html/template"
	"sync"

	"github.com/omeid/slurp"
	"github.com/omeid/slurp/tools/path"

	"github.com/yosssi/ace"
)

type Options ace.Options

func Compile(c *slurp.C, options Options, data interface{}) slurp.Stage {
	return func(in <-chan slurp.File, out chan<- slurp.File) {

		options := ace.Options(options)

		fs := []*ace.File{}

		var wg sync.WaitGroup
		defer wg.Wait() //Wait before all templates are executed.

		for file := range in {

			buf := new(bytes.Buffer)
			_, err := buf.ReadFrom(file.Reader)
			file.Close()
			if err != nil {
				c.Error(err)
				continue
			}

			s, err := file.Stat()
			if err != nil {
			  c.Error(err)
			  break
			}

			name := s.Name() //Probably filepath.Rel(file.Dir, file.Path) ??
			f := ace.NewFile(name, buf.Bytes())
			source := ace.NewSource(
				ace.NewFile("", nil),
				f,
				fs,
			)

			fs = append(fs, f)

			r, err := ace.ParseSource(source, &options)
			if err != nil {
				c.Error(err)
				continue
			}

			t, err := ace.CompileResultWithTemplate(html.New(name), r, &options)
			if err != nil {
				c.Error(err)
				continue
			}

			buf = new(bytes.Buffer)
			err = t.Execute(buf, data)
			if err != nil {
				c.Error(err)
				continue
			}

			file.Reader = buf
			file.FileInfo.SetSize(int64(buf.Len()))

			file, err = path.ReplaceExt(file, ".ace", ".html")
			if err != nil {
			  c.Error(err)
			  continue
			}

			out <- file

		}
	}
}
