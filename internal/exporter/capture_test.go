package exporter

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// This synthetic deck deliberately mirrors the REAL page templates'
// cascade structure (web/templates/export.html, presentation.html):
// --slide-width/--slide-height are set as an inline style on <body>, not on
// :root/<html>. #slide-container is a child of body and reads the vars via
// inheritance. Only --scale is left to the real :root/document.documentElement
// convention (app.js sets it there), which this test doesn't need to exercise.
//
// A synthetic page that instead declared --slide-width/--slide-height on
// :root (as the previous version of this test did) could not have caught the
// capture-mode override targeting the wrong element, because body has no own
// declaration to shadow a :root rule in that case. The base size below
// (800x400) is deliberately different from the requested capture size
// (200x100) so a failed override is visible as a wrong-sized screenshot
// instead of an accidental match.
const captureTestBaseWidth = 800
const captureTestBaseHeight = 400

func writeCaptureTestDeck(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "deck.html")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html><head><style>
  #slide-container { width: var(--slide-width); height: var(--slide-height); }
  .slide { display: none; width: 100%%; height: 100%%; }
  .slide.active { display: block; }
  .slide:nth-child(1) { background: red; }
  .slide:nth-child(2) { background: blue; }
</style></head>
<body style="--slide-width: %dpx; --slide-height: %dpx;">
  <div id="slide-container">
    <div class="slide active">One</div>
    <div class="slide">Two</div>
  </div>
  <script>
    window.goToSlide = function(i) {
      document.querySelectorAll('.slide').forEach(function(s, idx) {
        s.classList.toggle('active', idx === i);
      });
    };
  </script>
</body></html>`, captureTestBaseWidth, captureTestBaseHeight)
	if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
		t.Fatalf("failed to write test html: %v", err)
	}
	return htmlPath
}

func TestCaptureSlides_OneImagePerSlide(t *testing.T) {
	if !chromeAvailable() {
		t.Skip("no local Chrome/Chromium found, skipping headless capture test")
	}

	htmlPath := writeCaptureTestDeck(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	const wantWidth, wantHeight = 200, 100

	images, err := captureSlides(ctx, htmlPath, 2, wantWidth, wantHeight)
	if err != nil {
		t.Fatalf("captureSlides failed: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	for i, img := range images {
		if len(img) == 0 {
			t.Errorf("image %d is empty", i)
			continue
		}
		cfg, err := png.DecodeConfig(bytes.NewReader(img))
		if err != nil {
			t.Errorf("image %d: failed to decode PNG: %v", i, err)
			continue
		}
		// This is the crux of the capture-mode override bug: the real page
		// templates set --slide-width/--slide-height as an inline style on
		// <body>, so the capture override must target body (not :root) to
		// actually win the cascade. If it targets :root only, the page keeps
		// rendering at its base size (captureTestBaseWidth x
		// captureTestBaseHeight here) instead of the requested capture
		// resolution, and this assertion catches it.
		if cfg.Width != wantWidth || cfg.Height != wantHeight {
			t.Errorf("image %d: got %dx%d, want %dx%d (capture-mode override failed to force fixed dimensions)",
				i, cfg.Width, cfg.Height, wantWidth, wantHeight)
		}
	}
}
