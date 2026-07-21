package exporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCaptureSlides_OneImagePerSlide(t *testing.T) {
	if !chromeAvailable() {
		t.Skip("no local Chrome/Chromium found, skipping headless capture test")
	}

	dir := t.TempDir()
	htmlPath := filepath.Join(dir, "deck.html")
	html := `<!DOCTYPE html>
<html><head><style>
  :root { --slide-width: 200px; --slide-height: 100px; --scale: 1; }
  #slide-container { width: var(--slide-width); height: var(--slide-height); }
  .slide { display: none; width: 100%; height: 100%; }
  .slide.active { display: block; }
  .slide:nth-child(1) { background: red; }
  .slide:nth-child(2) { background: blue; }
</style></head>
<body>
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
</body></html>`
	if err := os.WriteFile(htmlPath, []byte(html), 0o644); err != nil {
		t.Fatalf("failed to write test html: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	images, err := captureSlides(ctx, htmlPath, 2, 200, 100)
	if err != nil {
		t.Fatalf("captureSlides failed: %v", err)
	}
	if len(images) != 2 {
		t.Fatalf("expected 2 images, got %d", len(images))
	}
	for i, img := range images {
		if len(img) == 0 {
			t.Errorf("image %d is empty", i)
		}
	}
}
