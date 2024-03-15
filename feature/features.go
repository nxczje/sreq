package feature

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/ffuf/pencode/pkg/pencode"
)

type File struct {
	Name string
	Body string
}

// Create a zip file
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

// Using
// prefix_length := 3
//
// feature.Itertools(prefix_length, feature.Char)
//
// Itertools return all possible combination of characters
func Itertools(prefix_length int, char []rune) []string {
	result := []string{}
	current := ""
	return itertoolwork(prefix_length, char, result, current)
}

// itertoolwork is a helper function for Itertools
func itertoolwork(prefix_length int, char []rune, result []string, current string) []string {
	if len(current) == prefix_length {
		return append(result, current)
	}
	for _, r := range char {
		result = itertoolwork(prefix_length, char, result, current+string(r))
	}
	return result
}

// SingleLine remove all \t and \n
func SingleLine(data string) string {
	result := strings.ReplaceAll(data, "\t", "")
	result = strings.ReplaceAll(result, "\n", "")
	return result
}

// WriteFile write data to new file
func WriteFile(filename string, data []byte) error {
	err := os.WriteFile(filename, data, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// ReadFile read file and return string
func ReadFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	dataret := []string{}
	for scanner.Scan() {
		dataret = append(dataret, scanner.Text())
	}
	return dataret, nil
}

// AppendFile append data to file
func AppendFile(filename string, data []byte) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	file.Write(data)
}

// For example it return like
//
//	map[string]interface {}{
//	    "message": []interface {}{
//	        map[string]interface {}{
//	            "content":   "3",
//	        },
//	    },
//	}
//
// xx, _ = datatemp["message"].([]interface{})[0].(map[string]interface{})["content"].(string)
func ReadJson(data string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// EscapeChar escape char
func EscapeChar(data string, char ...string) string {
	for _, c := range char {
		data = strings.ReplaceAll(data, c, "\\"+c)
	}
	return data
}

// DelItemArr delete item in array
func DelItemArr(arr []string, f func(string) bool) []string {
	newarr := make([]string, 0)
	for i, v := range arr {
		check := f(arr[i])
		if check {
			newarr = append(newarr, v)
		}
	}
	return newarr
}

// Str = "123" -> 123
func Str2Int(data string) int {
	result := 0
	for _, c := range data {
		result = result*10 + int(c-'0')
	}
	return result
}

// "123" -> []rune{1, 2, 3} -> string(rune) = "123"
func Str2Rune(data string) []rune {
	return []rune(data)
}

// Using pencode to encode/decode data
//
// Supported encoder: b64encode, hexencode, htmlescape, jsonescape, unicodeencode, urlencode, urlencodeall ,xmlencode, utf16, utf16be, xmlescape
//
// Supported decoder: b64decode, hexdecode, htmlunescape, jsonunescape, unicodedecode, urldecode, xmlunescape
//
// Supported hash: md5, sha1, sha224 ,sha256, sha384 ,sha512
//
// Other: lower, upper
func Pencode(listEncoder []string, data string) string {
	chain := pencode.NewChain()
	err := chain.Initialize(listEncoder)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	output, err := chain.Encode([]byte(data))
	if err != nil {
		panic(err)
	}
	return string(output)
}
