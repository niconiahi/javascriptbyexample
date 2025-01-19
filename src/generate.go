package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
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
const examplesFilePath = "./examples.txt"

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
	ID          string
	Name        string
	Hash        string
	Description string
	Console     string
	Segments    []Segment
	Next        *string
	Previous    *string
}

func makeDescription(path string) string {
	lines := readLines(path)
	return strings.Trim(lines[0], " /")
}

func makeExamples() []Example {
	examples := []Example{}
	ids := readLines(examplesFilePath)
	entries := readDir(examplesDir)
	if len(ids) != len(entries) {
		panic("the amount of examples in the txt should match the amount of examples in examples directory")
	}
	for i, id := range ids {
		jsFilePath := composePath(id, "js")
		content := readFile(jsFilePath)
		example := Example{}
		example.ID = id
		example.Name = makeName(example.ID)
		example.Hash = makeHash(content)
		example.Description = makeDescription(jsFilePath)
		example.Next = getNextId(i, ids)
		example.Previous = getPreviousId(i, ids)
		if isFile(jsFilePath) {
			segments := makeSegments(jsFilePath)
			example.Segments = segments
		}
		examples = append(examples, example)
	}
	return examples
}

func formatConsole(console []byte) string {
	lexer := ConsoleLexer
	formatter := html.New(html.WithClasses(true))
	iterator, err := lexer.Tokenise(nil, string(console))
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

func makeConsole(example Example) []byte {
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
	console := []byte("$ deno run " + example.ID + ".js" + "\n")
	console = append(console, output...)
	return console
}

func makeHashFile(example Example) {
	shFilePath := composePath(example.ID, "sh")
	hashFilePath := composePath(example.ID, "hash")
	if isFile(hashFilePath) {
		prevHash := readFile(hashFilePath)
		if example.Hash != string(prevHash) {
			fmt.Printf("recreating hash file\n")
			console := makeConsole(example)
			writeFile(
				shFilePath,
				[]byte(console),
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
		console := makeConsole(example)
		writeFile(
			shFilePath,
			[]byte(console),
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
	Segments []Row
	Console  Row
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
		var segments []Row
		description := Row{
			Markdown:  makeMarkdown(example.Description),
			CodeBlock: "",
		}
		segments = append(segments, description)
		for _, segment := range example.Segments {
			markdown := makeMarkdown(segment.Documentation)
			codeBlock := makeCodeBlock(segment.Code)
			row := Row{
				Markdown:  markdown,
				CodeBlock: codeBlock,
			}
			segments = append(segments, row)
		}
		console := Row{
			Markdown:  "",
			CodeBlock: formatConsole(makeConsole(example)),
		}
		t := makeTemplate(TemplateExample)
		var buffer bytes.Buffer
		err := t.Execute(&buffer, PageDataExample{
			Name:     example.Name,
			Next:     example.Next,
			Previous: example.Previous,
			Segments: segments,
			Console:  console,
		})
		if err != nil {
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
	return strings.Split(strings.TrimSpace(string(content)), "\n")
}

func makeSegments(path string) []Segment {
	lines := readLines(path)
	lines = lines[2:]
	nextSegment := Segment{}
	segments := []Segment{}
	for i, line := range lines {
		matchesDocs := docsRegex.MatchString(line)
		matchesCode := !matchesDocs
		matchesEmptyLine := strings.TrimSpace(line) == ""
		if matchesEmptyLine {
			segments = append(
				segments,
				nextSegment,
			)
			nextSegment = Segment{}
			continue
		}
		if matchesDocs {
			nextSegment.Documentation = strings.Trim(line, " /")
			continue
		}
		if matchesCode {
			nextCode := fmt.Sprintf("%v\n%v", nextSegment.Code, line)
			nextSegment.Code = strings.TrimSpace(nextCode)
			if i == len(lines)-1 {
				segments = append(
					segments,
					nextSegment,
				)
			}
			continue
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
		panic("could not write the bytes at the targeted path")
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

func getPreviousId(index int, entries []string) *string {
	if index == 0 {
		return nil
	}
	id := entries[index-1]
	return &id
}

func getNextId(index int, entries []string) *string {
	if index >= len(entries)-1 {
		return nil
	}
	id := entries[index+1]
	return &id
}

func makeName(id string) string {
	return strings.ReplaceAll(id, "-", " ")
}

var ConsoleLexer = chroma.MustNewLexer(
	&chroma.Config{
		Name: "ConsoleOutput",
	},
	func() chroma.Rules {
		return chroma.Rules{
			"root": {
				// the search starts with the capture of "$", after that it's the command
				{Pattern: `^\$`, Type: chroma.GenericError, Mutator: chroma.Push("command")},
			},
			"command": {
				// after line jump, it's all output
				{Pattern: `\n`, Type: chroma.Text, Mutator: chroma.Push("output")},

				// command style until line jump
				{Pattern: `[^\n]+`, Type: chroma.StringSymbol, Mutator: nil},
			},
			"output": {
				// style every line until the end as output
				{Pattern: `[^\n]+$\n?`, Type: chroma.GenericOutput, Mutator: nil},
			},
		}
	},
)
