document.addEventListener('DOMContentLoaded', () => {
  const container = document.getElementById('slide-container');
  if (!container) return;

  const slides = Array.from(container.querySelectorAll('.slide'));
  const btnPrev = document.getElementById('btn-prev');
  const btnNext = document.getElementById('btn-next');
  const progressBar = document.querySelector('.progress-bar');
  const slideNum = document.querySelector('.slide-number');

  let currentIndex = 0;

  // 16:9 Aspect Ratio Scaling
  function updateScale() {
    const targetWidth = 960;
    const targetHeight = 540;
    const windowWidth = window.innerWidth;
    const windowHeight = window.innerHeight;

    const scaleX = windowWidth / targetWidth;
    const scaleY = windowHeight / targetHeight;

    // Scale down if the screen is smaller than the target slide viewport,
    // and keep it aligned to fit within the viewport bounds.
    // Multiply by 0.95 to give a 5% margin to prevent slide edges from sticking to screen boundaries.
    const scale = Math.min(scaleX, scaleY) * 0.95;

    // Apply scale to the CSS custom property on the document root
    document.documentElement.style.setProperty('--scale', Math.max(scale, 0.1).toString());
  }

  // Go to a specific slide index and update CSS transition classes
  function goToSlide(index) {
    if (index < 0 || index >= slides.length) return;

    currentIndex = index;

    slides.forEach((slide, i) => {
      slide.classList.remove('active', 'past', 'future');
      if (i === currentIndex) {
        slide.classList.add('active');
      } else if (i < currentIndex) {
        slide.classList.add('past');
      } else {
        slide.classList.add('future');
      }
    });

    // Update progress bar width
    if (progressBar && slides.length > 0) {
      const progress = ((currentIndex + 1) / slides.length) * 100;
      progressBar.style.width = `${progress}%`;
    }

    // Update current/total slide text
    if (slideNum && slides.length > 0) {
      slideNum.textContent = `${currentIndex + 1} / ${slides.length}`;
    }

    // Toggle navigation button disabled states at boundaries
    if (btnPrev) {
      btnPrev.disabled = (currentIndex === 0);
    }
    if (btnNext) {
      btnNext.disabled = (currentIndex === slides.length - 1);
    }

    // Dispatch custom event to allow external hooks (e.g., SSE connection synchronizers) to listen to state changes
    document.dispatchEvent(new CustomEvent('slidechange', { 
      detail: { index: currentIndex } 
    }));
  }

  function nextSlide() {
    if (currentIndex < slides.length - 1) {
      goToSlide(currentIndex + 1);
    }
  }

  function prevSlide() {
    if (currentIndex > 0) {
      goToSlide(currentIndex - 1);
    }
  }

  // Bind Keyboard Navigation
  window.addEventListener('keydown', (e) => {
    const activeEl = document.activeElement;

    // Ignore keypresses if the focus is on input fields, textareas, select dropdowns, or contenteditable elements
    if (activeEl && (
        activeEl.tagName === 'INPUT' || 
        activeEl.tagName === 'TEXTAREA' || 
        activeEl.tagName === 'SELECT' || 
        activeEl.isContentEditable)) {
      return;
    }

    // If a button is focused, only ignore Spacebar to prevent double activation, but allow arrows
    if (activeEl && activeEl.tagName === 'BUTTON' && e.key === ' ') {
      return;
    }

    switch (e.key) {
      case 'ArrowRight':
      case ' ': // Spacebar
      case 'PageDown':
        e.preventDefault();
        nextSlide();
        break;
      case 'ArrowLeft':
      case 'PageUp':
        e.preventDefault();
        prevSlide();
        break;
    }
  });

  // Bind Navigation Buttons clicks
  if (btnPrev) {
    btnPrev.addEventListener('click', prevSlide);
  }
  if (btnNext) {
    btnNext.addEventListener('click', nextSlide);
  }

  // Bind scale updater to window resize
  window.addEventListener('resize', updateScale);

  // Run initial setups
  updateScale();
  goToSlide(0);

  let isHandlingSSE = false;

  // Hot-reload (triggered when the source markdown file changes) does a
  // full page reload, which would otherwise always snap back to slide 0
  // and lose the scroll position inside a long slide. Stash both in
  // sessionStorage right before reloading and restore them once the new
  // page's slide index is confirmed (per-tab only, not synced across
  // viewers — the SSE slide index is what stays synced across viewers).
  function saveScrollState() {
    const activeSlide = slides[currentIndex];
    sessionStorage.setItem('gophern:scrollIndex', String(currentIndex));
    sessionStorage.setItem('gophern:scrollTop', String(activeSlide ? activeSlide.scrollTop : 0));
  }

  function restoreScrollState(index) {
    const savedIndex = sessionStorage.getItem('gophern:scrollIndex');
    const savedScroll = sessionStorage.getItem('gophern:scrollTop');
    if (savedIndex === null || savedScroll === null) return;
    if (parseInt(savedIndex, 10) !== index) return;
    const activeSlide = slides[index];
    if (activeSlide) {
      activeSlide.scrollTop = parseInt(savedScroll, 10);
    }
  }

  if (window.location.protocol.startsWith('http')) {
    // Connect to Server-Sent Events stream
    const eventSource = new EventSource('/events');
    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.reload) {
          saveScrollState();
          window.location.reload();
          return;
        }
        const idx = data.slide;
        if (!isNaN(idx)) {
          if (idx !== currentIndex) {
            isHandlingSSE = true;
            goToSlide(idx);
            isHandlingSSE = false;
          }
          restoreScrollState(idx);
        }
      } catch (err) {
        console.error("Failed to parse SSE JSON payload:", err);
      }
    };
    eventSource.onerror = (err) => {
      console.error("SSE connection dropped, retrying...", err);
    };

    // Listen to slidechange event to synchronize other clients
    document.addEventListener('slidechange', (e) => {
      if (!isHandlingSSE) {
        const idx = e.detail.index;
        fetch('/api/slide', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ index: idx })
        }).catch(err => console.error('Failed to broadcast slide change:', err));
      }
    });
  }

  // Expose slide control functions globally for external integrations (such as Server-Sent Events controllers)
  window.goToSlide = goToSlide;
  window.getCurrentIndex = () => currentIndex;
  window.nextSlide = nextSlide;
  window.prevSlide = prevSlide;
});
