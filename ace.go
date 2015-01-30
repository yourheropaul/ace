package ace

// github.com/yosssi/gcss binding for slurp.
// No Configuration required.

import (
	"bytes"
	html "html/template"
	"sync"

	"github.com/omeid/slurp/s"
	"github.com/omeid/slurp/template"
	"github.com/yosssi/ace"
)

func Compile(c *s.C, data interface{}) s.Job {
	return func(in <-chan s.File, out chan<- s.File) {

		options := ace.Options{}
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

			file.Reader = template.NewTemplateReadCloser(c, wg, t, data)
			out <- file
		}
	}
}
