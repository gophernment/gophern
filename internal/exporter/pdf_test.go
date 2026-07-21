package exporter

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func makeTestPNG(t *testing.T, w, h int, c color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}
	return buf.Bytes()
}

func TestBuildPDF_OnePagePerImage(t *testing.T) {
	images := [][]byte{
		makeTestPNG(t, 100, 100, color.RGBA{255, 0, 0, 255}),
		makeTestPNG(t, 100, 100, color.RGBA{0, 255, 0, 255}),
		makeTestPNG(t, 100, 100, color.RGBA{0, 0, 255, 255}),
	}

	pdfBytes, err := buildPDF(images, 960, 540)
	if err != nil {
		t.Fatalf("buildPDF failed: %v", err)
	}

	if !bytes.HasPrefix(pdfBytes, []byte("%PDF-")) {
		t.Fatalf("output does not start with a PDF header, got: %q", pdfBytes[:min(20, len(pdfBytes))])
	}

	pageCount := bytes.Count(pdfBytes, []byte("/Type /Page/"))
	if pageCount == 0 {
		// fpdf may format the dict differently (e.g. "/Type/Page/" with no
		// space); fall back to counting "/Type /Page" as a substring that
		// also matches "/Type /Pages", then subtract that one Pages node.
		pageCount = bytes.Count(pdfBytes, []byte("/Type /Page")) - 1
	}
	if pageCount != len(images) {
		t.Errorf("expected %d pages, found %d", len(images), pageCount)
	}
}

func TestBuildPDF_EmptyImagesReturnsError(t *testing.T) {
	_, err := buildPDF(nil, 960, 540)
	if err == nil {
		t.Fatal("expected an error for empty image list, got nil")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
