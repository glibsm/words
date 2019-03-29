package highlight

import (
	"io"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/russross/blackfriday/v2"
)

type Renderer struct {
	blackfriday.Renderer
	Formatter html.Formatter
}

// New ...
func New() *Renderer {
	return &Renderer{
		Renderer: blackfriday.NewHTMLRenderer(
			blackfriday.HTMLRendererParameters{
				Flags: blackfriday.CommonHTMLFlags,
			},
		),
	}
}

// ...
func (r *Renderer) highlight(
	w io.Writer, text []byte, data blackfriday.CodeBlockData,
) error {
	var lexer chroma.Lexer

	if len(data.Info) > 0 {
		// try to see the defined lexer
		lexer = lexers.Get(string(data.Info))
	}

	if lexer == nil {
		// else try to default it
		lexer = lexers.Analyse(string(text))
	}

	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := lexer.Tokenise(nil, string(text))
	if err != nil {
		return err
	}
	return r.Formatter.Format(w, styles.Get("autumn"), iterator)
}

// RenderNode satisfies the Renderer interface
func (r *Renderer) RenderNode(
	w io.Writer, node *blackfriday.Node, entering bool,
) blackfriday.WalkStatus {
	switch node.Type {
	case blackfriday.CodeBlock:
		if err := r.highlight(w, node.Literal, node.CodeBlockData); err != nil {
			// fall through to default blackfriday
			return r.Renderer.RenderNode(w, node, entering)
		}
		return blackfriday.SkipChildren
	default:
		// fall through to default blackfriday
		return r.Renderer.RenderNode(w, node, entering)
	}
}
