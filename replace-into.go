package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type fileParts struct {
	head []byte
	tail []byte
}

// ReplaceInFile writes the provided content in the specified file by replacing
// any existing content between two markers.
func ReplaceInFile(
	file string,
	content string,
	startMarker string,
	endMarker string,
) error {
	fp, err := appFS.OpenFile(file, os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer fp.Close()

	parts, err := getParts(fp, startMarker, endMarker)
	if err != nil {
		return err
	}

	_, err = fp.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = fp.Truncate(0)
	if err != nil {
		return err
	}

	doc := [][]byte{
		parts.head,
		[]byte(startMarker + "\n"),
		[]byte(content),
		[]byte("\n" + endMarker),
		parts.tail,
	}
	err = writeDoc(fp, doc)

	return err
}

// getParts splits a file into the header and tail parts denoted by the given
// markers.
func getParts(reader io.Reader, startMarker string, endMarker string) (*fileParts, error) {
	fileContents, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if bytes.Index(fileContents, []byte(startMarker)) == -1 {
		return nil, fmt.Errorf("unable to find start marker: `%s`", startMarker)
	}
	if bytes.Index(fileContents, []byte(endMarker)) == -1 {
		return nil, fmt.Errorf("unable to find end marker: `%s`", endMarker)
	}

	result := &fileParts{}
	parts := bytes.Split(fileContents, []byte(startMarker))
	result.head = parts[0]

	parts = bytes.Split(fileContents, []byte(endMarker))
	result.tail = parts[1]

	return result, nil
}

// writeDoc writes out the parts of a document into the provided
// writer.
func writeDoc(writer io.Writer, parts [][]byte) error {
	_, err := writer.Write(bytes.Join(parts, []byte("")))
	return err
}
