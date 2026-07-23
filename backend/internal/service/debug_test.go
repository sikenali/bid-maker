package service

import (
	"fmt"
	"os"
	"testing"

	"github.com/unidoc/unioffice/v2/document"
)

func TestDebugDocx(t *testing.T) {
	doc := document.New()
	p := doc.AddParagraph()
	r := p.AddRun()
	r.AddText("第一章 总则")

	tmpFile, _ := os.CreateTemp("", "debug-*.docx")
	err := doc.SaveToFile(tmpFile.Name())
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if err != nil {
		t.Logf("SaveToFile failed (likely license): %v", err)
		return
	}

	doc2, err := document.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	defer doc2.Close()

	paras := doc2.Paragraphs()
	fmt.Printf("Round-trip paragraphs: %d\n", len(paras))
	for _, p := range paras {
		text := ""
		for _, r := range p.Runs() {
			text += r.Text()
		}
		fmt.Printf("  [%s]\n", text)
	}
}
