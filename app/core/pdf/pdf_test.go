package pdf

import (
	"bytes"
	"testing"
)

// Fonts used to be loaded from the build-time source path, which does not exist
// inside a container image. Embedding them keeps PDF output working on Cloud Run.
func TestNewPDFRendersThaiTextWithEmbeddedFonts(t *testing.T) {
	doc := NewPDF()
	doc.AddPage()
	doc.SetFont(FontFamily, "B", HeaderSize)
	doc.Cell(0, 10, "รายงานการขายยา")
	doc.SetFont(FontFamily, "I", FontSize)
	doc.Cell(0, 10, "ทดสอบ")

	var buffer bytes.Buffer
	if err := doc.Output(&buffer); err != nil {
		t.Fatalf("expected pdf output, got error: %v", err)
	}
	if buffer.Len() == 0 {
		t.Fatal("expected non-empty pdf output")
	}
}

func TestEmbeddedFontsArePresent(t *testing.T) {
	for name, font := range map[string][]byte{
		"regular": fontRegular,
		"bold":    fontBold,
		"italic":  fontItalic,
	} {
		if len(font) == 0 {
			t.Fatalf("%s font was not embedded", name)
		}
	}
}
