package exporter

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"

	"github.com/go-pdf/fpdf"
)

// buildPDF assembles one PNG image per PDF page. Each page's size (in
// points) matches widthPx/heightPx 1:1, so the PDF page's aspect ratio
// matches the captured slide's aspect ratio exactly.
func buildPDF(images [][]byte, widthPx, heightPx int) ([]byte, error) {
	if len(images) == 0 {
		return nil, fmt.Errorf("buildPDF: no images to assemble")
	}

	pageW := float64(widthPx)
	pageH := float64(heightPx)

	// Use "P" (portrait) orientation even for landscape-shaped inputs: fpdf
	// treats the Size passed to NewCustom/AddPageFormat as portrait-relative
	// and swaps Wd/Ht internally when orientation is "L", which would turn a
	// 960x540 page into a 540x960 one. With "P", fpdf uses Wd/Ht exactly as
	// given, so the resulting page is precisely pageW x pageH.
	pdf := fpdf.NewCustom(&fpdf.InitType{
		OrientationStr: "P",
		UnitStr:        "pt",
		SizeStr:        "",
		Size:           fpdf.SizeType{Wd: pageW, Ht: pageH},
	})

	for i, imgBytes := range images {
		if _, _, err := image.DecodeConfig(bytes.NewReader(imgBytes)); err != nil {
			return nil, fmt.Errorf("buildPDF: slide %d: invalid image: %w", i+1, err)
		}

		imgName := fmt.Sprintf("slide-%d", i)
		pdf.RegisterImageOptionsReader(imgName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(imgBytes))
		pdf.AddPageFormat("P", fpdf.SizeType{Wd: pageW, Ht: pageH})
		pdf.ImageOptions(imgName, 0, 0, pageW, pageH, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
	}

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("buildPDF: %w", err)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("buildPDF: failed to write output: %w", err)
	}
	return buf.Bytes(), nil
}
