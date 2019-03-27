package words

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/russross/blackfriday/v2"
)

var sep = []byte("---")

// TODO: allow to completely override CSS
const mainTemplate = `
<!doctype html>
<html>
<head>
  <meta charset="utf-8">
  <title>{{.Title}}</title>
	<style>
		body {
			margin-top: 5em;
			line-height:1.5;
			font-family: verdana, serif;
			font-size: 17px;
			width: 700px;
			max-width: 100%;
			margin-left: auto;
			margin-right: auto;
		}

		pre {
			background-color: #f2f2f2;
			padding: 1.5em;
			color: #424242;
		}

		a {
			text-decoration: none;
			font-weight: bold;
			color: royalblue;
		}
		a:hover {
			text-decoration: underline;
		}

		img {
			max-width: 100%;
		}
	</style>
</head>
<body>
  {{template "body"}}
</body>
</html>
`

const postTemplate = `
{{define "body"}}
<a href="..">← BACK</a>
{{template "content"}}
{{end}}
`

const indexTemplate = `
{{define "body"}}
{{template "blurb"}}
<h2>Posts</h2>
{{template "list"}}
{{end}}
`

func renderPost(p *post) ([]byte, error) {
	t := template.Must(template.New("page").Parse(mainTemplate))
	t.Parse(postTemplate)

	t.New("content").Parse(string(p.html))

	var out bytes.Buffer
	if err := t.Execute(&out, map[string]interface{}{
		"Title": p.Title,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func renderIndex(i index, posts []*post) ([]byte, error) {
	t := template.Must(template.New("page").Parse(mainTemplate))
	template.Must(t.Parse(indexTemplate))

	template.Must(
		t.New("blurb").Parse(
			string(blackfriday.Run(i.html)),
		),
	)

	var listBuf bytes.Buffer
	for _, p := range posts {
		fmt.Fprintf(
			&listBuf,
			`%s <a href="%s">%s</a><br />`,
			p.Date.Format("2006-01-02"), p.Slug, p.Title,
		)
	}
	template.Must(
		t.New("list").Parse(
			listBuf.String(),
		),
	)

	var out bytes.Buffer
	if err := t.Execute(&out, map[string]string{
		"Title": i.title,
	}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
