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

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"

	// "github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/russross/blackfriday/v2"
)

const publicDir = "./public"
const examplesDir = "./examples"

type Lexer string

const (
	Javascript Lexer = "javascript"
	Console    Lexer = "console"
)

var docsRegex = regexp.MustCompile(`^(\s*(\/\/|#)\s|\s*\/\/$)`)

type Segment struct {
	Documentation string
	CodeLines     []string
}

func main() {
	if !isDir(examplesDir) {
		panic("examples path is not a dir")
	}
	examples := readDir(examplesDir)
	for _, example := range examples {
		info, err := example.Info()
		if err != nil {
			panic("not possible to get file info for example")
		}

		jsFilePath := examplesDir + "/" + example.Name() + "/" + info.Name() + ".js"
		shFilePath := examplesDir + "/" + example.Name() + "/" + info.Name() + ".sh"
		hashFilePath := examplesDir + "/" + example.Name() + "/" + info.Name() + ".hash"

		command := exec.Command(
			"deno", "run",
			"--allow-read",
			jsFilePath,
		)
		output, err := command.CombinedOutput()
		if err != nil {
			panic("could not execute the deno command for the js example file")
		}

		prompt := []byte("$ deno run " + info.Name() + ".js" + "\n")
		content := append(prompt, output...)
		hash := makeHash(content)
		if isFile(jsFilePath) {
			segments := makeSegments(jsFilePath)
			for _, segment := range segments {
				markdown := makeMarkdown(segment.Documentation)
				fmt.Printf("%v \n", markdown)
				codeBlock := strings.Join(segment.CodeLines, "\n")
				// fmt.Printf("codeblock %v\n", codeBlock)
				// lexer := makeShellLexer()
				// lexer = chroma.Coalesce(lexer)
				lexer := lexers.Get(string(Javascript))
				style := styles.Get("monokai")
				if style == nil {
					panic("there is no style for the given name")
				}
				formatter := html.New(html.WithClasses(true))
				iterator, err := lexer.Tokenise(nil, codeBlock)
				// fmt.Printf("iterator %v\n", iterator)
				if err != nil {
					panic("there was an error while tokenasing the code")
				}
				buffer := bytes.NewBuffer([]byte{})
				err = formatter.Format(buffer, style, iterator)
				if err != nil {
					panic("there was an error while formatting the code")
				}
				fmt.Printf("formatted %v\n", buffer.String())
			}
		}
		if isFile(hashFilePath) {
			prevHash := readFile(hashFilePath)
			if hash != prevHash {
				fmt.Printf("recreating hash file\n")
				writeFile(
					shFilePath,
					[]byte(content),
				)
				writeFile(
					hashFilePath,
					[]byte(hash),
				)
			} else {
				fmt.Printf("hash files match. Omitting creating files\n")
			}
		} else {
			fmt.Printf("creating fresh hash file\n")
			writeFile(
				shFilePath,
				[]byte(content),
			)
			writeFile(
				hashFilePath,
				[]byte(hash),
			)
		}
	}
}

// func getLexer(path string) string {
// 	if strings.HasSuffix(path, ".js") {
// 		return "javascript"
// 	} else {
// 		return "console"
// 	}
// }

func readLines(path string) []string {
	content := readFile(path)
	return strings.Split(content, "\n")
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
			segments = append(segments, nextSegment)
			nextSegment = Segment{}
		}
		if matchesDocs {
			nextSegment.Documentation = strings.Trim(line, " /")
		}
		if matchesCode {
			nextSegment.CodeLines = append(nextSegment.CodeLines, line)
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

func readFile(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		panic("not possible to find the previous hash file")
	}
	return string(content)
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
