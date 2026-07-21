package exporter

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
)

// chromeAvailable reports whether a local Chrome/Chromium install can be
// launched headlessly. Used to skip PDF export (and its tests) with a clear
// message instead of failing deep inside chromedp on machines without a
// browser installed.
func chromeAvailable() bool {
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	err := chromedp.Run(ctx, chromedp.Navigate("about:blank"))
	return err == nil
}

// captureStyleOverride forces a fixed, unscaled render at widthPx/heightPx
// and hides on-screen navigation chrome, so the screenshot shows only slide
// content at full resolution regardless of what device pixel size the
// live-view scale-to-fit logic would otherwise pick.
const captureStyleOverrideTemplate = `
:root { --slide-width: %dpx !important; --slide-height: %dpx !important; --scale: 1 !important; }
* { transition: none !important; }
.nav-buttons, .progress-bar-container, .slide-number { display: none !important; }
`

// captureSlides renders each slide of the deck at htmlPath (a file:// path
// already positioned next to its asset/ folder) to a full-resolution PNG at
// widthPx x heightPx, in slide order. slideCount must match the number of
// ".slide" elements the page renders.
func captureSlides(ctx context.Context, htmlPath string, slideCount, widthPx, heightPx int) ([][]byte, error) {
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, chromedp.DefaultExecAllocatorOptions[:]...)
	defer cancel()

	browserCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	styleOverride := fmt.Sprintf(captureStyleOverrideTemplate, widthPx, heightPx)

	images := make([][]byte, 0, slideCount)

	tasks := chromedp.Tasks{
		chromedp.Navigate("file://" + htmlPath),
		chromedp.EmulateViewport(int64(widthPx), int64(heightPx)),
		chromedp.WaitVisible("#slide-container", chromedp.ByID),
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				var style = document.createElement('style');
				style.textContent = %q;
				document.head.appendChild(style);
			})();
		`, styleOverride), nil),
	}
	if err := chromedp.Run(browserCtx, tasks); err != nil {
		return nil, fmt.Errorf("captureSlides: failed to load %s: %w", htmlPath, err)
	}

	for i := 0; i < slideCount; i++ {
		var buf []byte
		err := chromedp.Run(browserCtx, chromedp.Tasks{
			chromedp.Evaluate(fmt.Sprintf(`window.goToSlide(%d)`, i), nil),
			chromedp.Screenshot("#slide-container", &buf, chromedp.NodeVisible, chromedp.ByID),
		})
		if err != nil {
			return nil, fmt.Errorf("captureSlides: failed to capture slide %d: %w", i+1, err)
		}
		images = append(images, buf)
	}

	return images, nil
}
