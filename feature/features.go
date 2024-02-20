package feature

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"os"
	"strings"

	"github.com/kr/pretty"
)

type File struct {
	Name string
	Body string
}

func ZipFile(destinationZip string, sourceFile ...File) error {
	// Create a buffer to store the ZIP file content
	var buf bytes.Buffer

	// Create a new ZIP writer
	zw := zip.NewWriter(&buf)

	// Add files to the ZIP archive
	fileContents := make(map[string]string)
	for _, file := range sourceFile {
		fileContents[file.Name] = file.Body
	}

	for fileName, content := range fileContents {
		f, err := zw.Create(fileName)
		if err != nil {
			return err
		}
		_, err = f.Write([]byte(content))
		if err != nil {
			return err
		}
	}

	// Close the ZIP writer
	err := zw.Close()
	if err != nil {
		return err
	}

	// Write the ZIP file to disk
	err = os.WriteFile(destinationZip, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func Itertools(prefix_length int, char []rune) []string {
	result := []string{}
	current := ""
	return itertoolwork(prefix_length, char, result, current)
}

func itertoolwork(prefix_length int, char []rune, result []string, current string) []string {
	if len(current) == prefix_length {
		return append(result, current)
	}
	for _, r := range char {
		result = itertoolwork(prefix_length, char, result, current+string(r))
	}
	return result
}

func CalcMd5(data string) string {
	hasher := md5.New()
	hasher.Write([]byte(data))
	rs := hasher.Sum(nil)
	return hex.EncodeToString(rs)
}

func CalcSHA1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return pretty.Sprintf("%x", bs)
}

func CalcSHA512(data string) string {
	h := sha512.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return pretty.Sprintf("%x", bs)
}

func CalcSHA256(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)
	return pretty.Sprintf("%x", bs)
}

func SingleLine(data string) string {
	result := strings.ReplaceAll(data, "\t", "")
	result = strings.ReplaceAll(result, "\n", "")
	return result
}

func WriteFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
