package exporter

import (
	"bytes"
	"fmt"
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

func TestBuildPDF_PageDimensionsMatchInputPixels(t *testing.T) {
	images := [][]byte{
		makeTestPNG(t, 100, 100, color.RGBA{255, 0, 0, 255}),
	}

	widthPx, heightPx := 960, 540
	pdfBytes, err := buildPDF(images, widthPx, heightPx)
	if err != nil {
		t.Fatalf("buildPDF failed: %v", err)
	}

	// fpdf writes page size as "/MediaBox [0 0 %.2f %.2f]" in points. Since
	// buildPDF uses "pt" units, the MediaBox values should equal widthPx and
	// heightPx exactly, in that order (width first, then height) — not
	// swapped.
	wantMediaBox := []byte(fmt.Sprintf("/MediaBox [0 0 %.2f %.2f]", float64(widthPx), float64(heightPx)))
	if !bytes.Contains(pdfBytes, wantMediaBox) {
		idx := bytes.Index(pdfBytes, []byte("/MediaBox"))
		got := "<not found>"
		if idx >= 0 {
			end := idx + 40
			if end > len(pdfBytes) {
				end = len(pdfBytes)
			}
			got = string(pdfBytes[idx:end])
		}
		t.Errorf("expected PDF to contain %q, got %q (page dims possibly swapped)", wantMediaBox, got)
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
