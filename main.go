package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

type Page struct {
	Title      string
	Desc       string
	Date       *time.Time
	Content    template.HTML
	InlineCSS  template.HTML
	InlineJS   template.HTML
	AssetsPath string
}

type Metadata struct {
	Title  string     `yaml:"title"`
	Desc   string     `yaml:"desc"`
	Date   *time.Time `yaml:"date,omitempty"`
	Status bool       `yaml:"status"`
	CSS    string     `yaml:"css,omitempty"`
	JS     string     `yaml:"js,omitempty"`
}

var markdownParser = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokailight"),
			highlighting.WithFormatOptions(html.WithLineNumbers(false)),
		),
	),
	goldmark.WithRendererOptions(gmhtml.WithUnsafe()),
)

func main() {
	srcDir := "src"
	distDir := "dist"

	err := prepareDirectories(distDir)
	if err != nil {
		log.Fatalf("Error setting up directories: %v", err)
	}

	err = copyAssets(filepath.Join(srcDir, "assets"), distDir)
	if err != nil {
		log.Printf("Error copying assets: %v", err)
	}

	files, err := collectMarkdownFiles(srcDir)
	if err != nil {
		log.Fatalf("Error collecting markdown files: %v", err)
	}

	processFilesConcurrently(files, srcDir, distDir)
}

func prepareDirectories(distDir string) error {
	err := os.RemoveAll(distDir)
	if err != nil {
		return fmt.Errorf("failed to clear destination directory %s: %w", distDir, err)
	}

	err = os.MkdirAll(distDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", distDir, err)
	}

	return nil
}

func collectMarkdownFiles(srcDir string) ([]string, error) {
	var files []string

	err := filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		files = append(files, path)
		return nil
	})

	return files, err
}

func processFilesConcurrently(files []string, srcDir, distDir string) {
	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			err := processMarkdownFile(file, srcDir, distDir)
			if err != nil {
				log.Printf("Error processing file %s: %v", file, err)
			}
		}(file)
	}
	wg.Wait()
}

func generateHighlightCSS(styleName string) (template.HTML, error) {
	style := styles.Get(styleName)
	if style == nil {
		return "", fmt.Errorf("style %s not found", styleName)
	}

	var buf bytes.Buffer
	err := html.New().WriteCSS(&buf, style)
	if err != nil {
		return "", fmt.Errorf("error writing CSS: %w", err)
	}

	return template.HTML(buf.String()), nil
}

func copyAssets(src, dest string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, strings.TrimPrefix(path, src))

		if info.IsDir() {
			err = os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating directory %s: %w", destPath, err)
			}
			return nil
		}

		return copyFile(path, destPath)
	})
}

func copyFile(srcPath, destPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("error copying file from %s to %s: %w", srcPath, destPath, err)
	}

	return nil
}

func processMarkdownFile(mdPath, srcDir, distDir string) error {
	content, err := os.ReadFile(mdPath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", mdPath, err)
	}

	var metadata Metadata
	body, err := frontmatter.Parse(bytes.NewReader(content), &metadata)
	if err != nil {
		return fmt.Errorf("failed to parse front matter for %s: %w", mdPath, err)
	}

	if !metadata.Status {
		return nil
	}

	metadata.Title = getDefaultTitle(mdPath, metadata.Title)
	htmlContent, err := markdownToHTML(body)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML for %s: %w", mdPath, err)
	}

	inlineCSS, err := generateHighlightCSS("monokai")
	if err != nil {
		return fmt.Errorf("failed to generate CSS: %w", err)
	}

	inlineJS := template.HTML(metadata.JS)

	page := createPage(metadata, htmlContent, inlineCSS, inlineJS)

	return renderHTMLPage(page, mdPath, srcDir, distDir)
}

func getDefaultTitle(mdPath, title string) string {
	if title == "" {
		return strings.TrimSuffix(filepath.Base(mdPath), ".md")
	}
	return title
}

func createPage(metadata Metadata, content string, inlineCSS, inlineJS template.HTML) Page {
	return Page{
		Title:      metadata.Title,
		Desc:       metadata.Desc,
		Date:       metadata.Date,
		Content:    template.HTML(content),
		InlineCSS:  inlineCSS,
		InlineJS:   inlineJS,
		AssetsPath: "./",
	}
}

func renderHTMLPage(page Page, mdPath, srcDir, distDir string) error {
	layoutTemplatePath := filepath.Join(srcDir, "layout.html")
	layoutTemplate, err := template.ParseFiles(layoutTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to parse layout template %s: %w", layoutTemplatePath, err)
	}

	var output bytes.Buffer
	err = layoutTemplate.Execute(&output, page)
	if err != nil {
		return fmt.Errorf("failed to execute layout template for page %s: %w", page.Title, err)
	}

	return saveHTMLFile(output.Bytes(), mdPath, srcDir, distDir)
}

func saveHTMLFile(data []byte, mdPath, srcDir, distDir string) error {
	relativePath := strings.TrimPrefix(mdPath, srcDir)
	htmlPath := filepath.Join(distDir, strings.TrimSuffix(relativePath, ".md")+".html")

	err := os.MkdirAll(filepath.Dir(htmlPath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create directory for output file %s: %w", htmlPath, err)
	}

	err = os.WriteFile(htmlPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write HTML output to %s: %w", htmlPath, err)
	}

	return nil
}

func markdownToHTML(content []byte) (string, error) {
	var buf bytes.Buffer
	err := markdownParser.Convert(content, &buf)
	if err != nil {
		return "", fmt.Errorf("failed to convert markdown content: %w", err)
	}
	return buf.String(), nil
}
