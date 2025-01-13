package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
)

const publicDir = "./public"
const examplesDir = "./examples"

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
		if isFile(hashFilePath) {
			prevHash := readFile(hashFilePath)
			if hash != string(prevHash) {
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

func writeFile(path string, content []byte) {
	err := os.WriteFile(
		path,
		content,
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

func makeHash(content []byte) string {
	hasher := sha1.New()
	hasher.Write([]byte(content))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}
