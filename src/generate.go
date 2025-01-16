package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/russross/blackfriday/v2"
)

const publicDir = "./public"
const examplesDir = "./examples"
const templatesDir = "./templates"

type Lexer string

const (
	LexerJavascript Lexer = "javascript"
)

var docsRegex = regexp.MustCompile(`^(\s*(\/\/|#)\s|\s*\/\/$)`)

type Segment struct {
	Documentation string
	Code          string
}

type Example struct {
	ID       string
	Name     string
	Hash     string
	Segments []Segment
	Next     *string
	Previous *string
}

func makeExamples() []Example {
	examples := []Example{}
	entries := readDir(examplesDir)
	for i, entry := range entries {
		example := Example{}
		info, err := entry.Info()
		if err != nil {
			panic("not possible to get file info for example")
		}
		example.ID = info.Name()
		example.Name = info.Name()
		example.Next = getNextId(i, entries)
		example.Previous = getPreviousId(i, entries)
		jsFilePath := composePath(entry.Name(), "js")
		if isFile(jsFilePath) {
			segments := makeSegments(jsFilePath)
			example.Segments = segments
		}
		content := readFile(jsFilePath)
		hash := makeHash(content)
		example.Hash = hash
		examples = append(examples, example)
	}
	return examples
}

func makeTerminal(example Example) []byte {
	jsFilePath := composePath(example.ID, "js")
	command := exec.Command(
		"deno", "run",
		"--allow-read",
		jsFilePath,
	)
	output, err := command.CombinedOutput()
	if err != nil {
		panic("could not execute the deno command for the js example file")
	}
	prompt := []byte("$ deno run " + example.ID + ".js" + "\n")
	prompt = append(prompt, output...)
	return prompt
}

func makeHashFile(example Example) {
	shFilePath := composePath(example.ID, "sh")
	hashFilePath := composePath(example.ID, "hash")
	if isFile(hashFilePath) {
		prevHash := readFile(hashFilePath)
		if example.Hash != string(prevHash) {
			fmt.Printf("recreating hash file\n")
			terminal := makeTerminal(example)
			writeFile(
				shFilePath,
				[]byte(terminal),
			)
			writeFile(
				hashFilePath,
				[]byte(example.Hash),
			)
		} else {
			fmt.Printf("hash files match. Omitting creating files\n")
		}
	} else {
		fmt.Printf("creating fresh hash file\n")
		terminal := makeTerminal(example)
		writeFile(
			shFilePath,
			[]byte(terminal),
		)
		writeFile(
			hashFilePath,
			[]byte(example.Hash),
		)
	}
}

func makeCss() []byte {
	style := getStyle()
	formatter := html.New(html.WithClasses(true))
	var buffer bytes.Buffer
	err := formatter.WriteCSS(&buffer, style)
	if err != nil {
		panic("failed to generate CSS")
	}
	return buffer.Bytes()
}

func getStyle() *chroma.Style {
	return styles.Get("tokyonight-night")
}

func makeCodeBlock(code string) string {
	lexer := lexers.Get(string(LexerJavascript))
	formatter := html.New(html.WithClasses(true))
	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		panic("there was an error while tokenasing the code")
	}
	buffer := bytes.NewBuffer([]byte{})
	style := getStyle()
	err = formatter.Format(buffer, style, iterator)
	if err != nil {
		panic("there was an error while formatting the code")
	}
	return buffer.String()
}

type Row struct {
	Markdown  string
	CodeBlock string
}
type PageDataExample struct {
	Name     string
	Next     *string
	Previous *string
	Rows     []Row
}
type PageTemplate string

const (
	TemplateExample = "example"
	TemplateIndex   = "index"
)

func makeTemplate(t PageTemplate) *template.Template {
	s := string(t)
	tmpl := template.New(s + ".tmpl")
	return template.Must(tmpl.ParseFiles(templatesDir + "/" + s + ".tmpl"))
}

func makeFiles(examples []Example) {
	for _, example := range examples {
		var rows []Row
		description := Row{
			Markdown:  example.ID,
			CodeBlock: example.ID,
		}
		rows = append(rows, description)
		for _, segment := range example.Segments {
			markdown := makeMarkdown(segment.Documentation)
			codeBlock := makeCodeBlock(segment.Code)
			row := Row{
				Markdown:  markdown,
				CodeBlock: codeBlock,
			}
			rows = append(rows, row)
		}
		t := makeTemplate(TemplateExample)
		var buffer bytes.Buffer
		err := t.Execute(&buffer, PageDataExample{
			Name:     example.Name,
			Next:     example.Next,
			Previous: example.Previous,
			Rows:     rows,
		})
		if err != nil {
			fmt.Printf("%v\n", err.Error())
			panic("error while generating the page template")
		}
		writeFile(publicDir+"/"+example.ID+".html", buffer.Bytes())
		makeHashFile(example)
	}
}

func main() {
	if !isDir(examplesDir) {
		panic("examples path is not a dir")
	}
	examples := makeExamples()
	css := makeCss()
	writeFile(publicDir+"/code.css", css)
	makeFiles(examples)
}

func composePath(id string, extension string) string {
	return examplesDir + "/" + id + "/" + id + "." + extension
}

func readLines(path string) []string {
	content := readFile(path)
	return strings.Split(string(content), "\n")
}

func makeSegments(path string) []Segment {
	lines := readLines(path)
	nextSegment := Segment{}
	segments := []Segment{}
	for _, line := range lines {
		matchesDocs := docsRegex.MatchString(line)
		matchesCode := !matchesDocs
		matchesEmptyLine := strings.TrimSpace(line) == ""
		if matchesEmptyLine {
			segments = append(
				segments,
				nextSegment,
			)
			nextSegment = Segment{}
		}
		if matchesDocs {
			nextSegment.Documentation = strings.Trim(line, " /")
		}
		if matchesCode {
			nextCode := fmt.Sprintf("%v\n%v", nextSegment.Code, line)
			nextSegment.Code = strings.TrimSpace(nextCode)
		}
	}
	return segments
}

func makeMarkdown(src string) string {
	return string(blackfriday.Run([]byte(src)))
}

func writeFile(path string, bytes []byte) {
	err := os.WriteFile(
		path,
		bytes,
		0644,
	)
	if err != nil {
		panic("could not write the prompt text for the js example file")
	}
}

func isDir(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return stat.IsDir()
}

func readDir(path string) []os.DirEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		panic("reading the examples dir failed")
	}
	return entries
}

func readFile(path string) []byte {
	content, err := os.ReadFile(path)
	if err != nil {
		panic("not possible to find the previous hash file")
	}
	return content
}

func isFile(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

func makeHash(bytes []byte) string {
	hasher := sha1.New()
	hasher.Write([]byte(bytes))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

func makeShellLexer() chroma.Lexer {
	config := &chroma.Config{}
	lexer := chroma.MustNewLexer(
		config,
		func() chroma.Rules {
			return chroma.Rules{
				"root": {
					{
						Pattern: `#.*\n`,
						Type:    chroma.Comment,
						Mutator: nil,
					},
					{
						Pattern: `*`,
						Type:    chroma.GenericOutput,
						Mutator: nil,
					},
				},
			}
		},
	)
	return lexer
}

func getPreviousId(index int, entries []fs.DirEntry) *string {
	if index == 0 {
		return nil
	}
	id := entries[index-1].Name()
	return &id
}

func getNextId(index int, entries []fs.DirEntry) *string {
	if index >= len(entries)-1 {
		return nil
	}
	id := entries[index+1].Name()
	return &id
}

func makeName(id string) string {
	firstLetter := id[0:1]
	rest := id[1:]
	return fmt.Sprintf("%v%v", strings.ToUpper(firstLetter), strings.ToLower(rest))
}
