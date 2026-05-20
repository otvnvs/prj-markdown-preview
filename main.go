package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type markdownPreviewer struct {
	app.Compo
	markdownText string
}

func (m *markdownPreviewer) OnMount(ctx app.Context) {
	m.markdownText = ""
	m.updatePreview()
}

func (m *markdownPreviewer) updatePreview() {
	el := app.Window().GetElementByID("preview")
	if el.Truthy() {
		el.Set("innerHTML", m.renderMarkdown())
	}
}

func (m *markdownPreviewer) Render() app.UI {
	return app.Div().
		Class("markdown-previewer").
		Body(
			app.H1().Body(
				app.Text("Markdown Previewer"),
			),
			app.P().Body(
				app.Text("Type markdown on the left, see preview on the right"),
			),
			app.Div().
				Class("container").
				Body(
					app.Div().
						Class("editor-pane").
						Body(
							app.Textarea().
								ID("editor-field").
								Class("editor").
								Placeholder("Type your markdown here...").
								OnInput(m.onMarkdownInput),
						),
					app.Div().
						Class("preview-pane").
						Body(
							app.Div().
								ID("preview").
								Class("preview"),
						),
				),
			app.Div().
				Class("button-group").
				Body(
					app.Button().
						Class("btn clear-btn").
						OnClick(m.clearText).
						Body(
							app.Text("Clear"),
						),
				),
		)
}

func (m *markdownPreviewer) onMarkdownInput(ctx app.Context, e app.Event) {
	m.markdownText = ctx.JSSrc().Get("value").String()
	m.updatePreview()
}

func (m *markdownPreviewer) renderMarkdown() string {
	if m.markdownText == "" {
		return `<p class="empty">Start typing markdown...</p>`
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	var buf strings.Builder
	if err := md.Convert([]byte(m.markdownText), &buf); err != nil {
		return `<p class="error">Error: ` + err.Error() + `</p>`
	}
	return buf.String()
}

func (m *markdownPreviewer) clearText(ctx app.Context, e app.Event) {
	m.markdownText = ""

	textarea := app.Window().GetElementByID("editor-field")
	if textarea.Truthy() {
		textarea.Set("value", "")
	}

	m.updatePreview()
	ctx.Update()
}

var handler = &app.Handler{
	Name:        "Markdown Previewer",
	Description: "A live markdown previewer built with go-app",
	Styles: []string{
		"/web/style.css",
	},
}


func generate(dir string) error {
	fmt.Printf("Generating static site into %q...\n", dir)

	if err := app.GenerateStaticWebsite(dir, handler); err != nil {
		return fmt.Errorf("generate static website: %w", err)
	}

	// Copy wasm_exec.js from the Go installation
	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		// fall back to `go env GOROOT` output embedded at build time isn't
		// available here, so just warn and skip
		fmt.Println("Warning: GOROOT not set, skipping wasm_exec.js copy.")
		fmt.Println("Copy it manually: $(go env GOROOT)/misc/wasm/wasm_exec.js ->", dir)
	} else {
		src := filepath.Join(goroot, "misc", "wasm", "wasm_exec.js")
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read wasm_exec.js: %w", err)
		}
		dst := filepath.Join(dir, "wasm_exec.js")
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("write wasm_exec.js: %w", err)
		}
		fmt.Println("Copied wasm_exec.js")
	}

	// Copy web/ assets
	entries, err := os.ReadDir("web")
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read web/: %w", err)
	}
	webDst := filepath.Join(dir, "web")
	if err := os.MkdirAll(webDst, 0755); err != nil {
		return fmt.Errorf("mkdir web/: %w", err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join("web", e.Name()))
		if err != nil {
			return fmt.Errorf("read web/%s: %w", e.Name(), err)
		}
		if err := os.WriteFile(filepath.Join(webDst, e.Name()), data, 0644); err != nil {
			return fmt.Errorf("write web/%s: %w", e.Name(), err)
		}
		fmt.Printf("Copied web/%s\n", e.Name())
	}

	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  GOARCH=wasm GOOS=js go build -ldflags=\"-s -w\" -trimpath -o %s/app.wasm .\n", dir)
	fmt.Println("  Then commit the", dir, "folder and push.")
	return nil
}

func main() {
	app.Route("/", func() app.Composer { return &markdownPreviewer{} })
	app.RunWhenOnBrowser()

	generateFlag := flag.Bool("generate", false, "Generate static website instead of starting the server")
	dir := flag.String("dir", "docs", "Output directory for static generation")
	repo := flag.String("repo", "", "GitHub Pages repo name (e.g. prj-markdown-preview), sets asset base path")
	flag.Parse()

	if *generateFlag {
		if *repo != "" {
			handler.Resources = app.GitHubPages(*repo)
		}
		if err := generate(*dir); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.Handle("/", handler)
	fmt.Println("Serving on http://localhost:8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
