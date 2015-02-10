package ace

// github.com/yosssi/gcss binding for slurp.
// No Configuration required.

import (
	"bytes"
	html "html/template"
	"strings"
	"sync"

	"github.com/omeid/slurp"
	"github.com/omeid/slurp/stages/template"
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
				c.Println(err)
				break
			}

			name := file.Stat.Name() //Probably filepath.Rel(file.Dir, file.Path) ??
			f := ace.NewFile(name, buf.Bytes())
			source := ace.NewSource(
				ace.NewFile("", nil),
				f,
				fs,
			)

			fs = append(fs, f)

			r, err := ace.ParseSource(source, &options)
			if err != nil {
				c.Println(err)
				break
			}

			t, err := ace.CompileResultWithTemplate(html.New(name), r, &options)
			if err != nil {
				c.Println(err)
				break
			}

			path := strings.TrimSuffix(file.Path, ".ace") + ".html"

			stat := slurp.FileInfoFrom(file.Stat)

			stat.SetName(path)
			file.Path = path

			file.Reader = template.NewTemplateReadCloser(c, wg, t, data)
			out <- file
		}
	}
}
