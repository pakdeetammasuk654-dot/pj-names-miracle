/**
 * Responsive Navigation for ชื่อดี.com
 */
(function () {
  function initNav() {
    var header = document.querySelector('.site-header');
    if (!header) return;

    var container = header.querySelector('.header-container');
    if (!container) return;

    var navLinks = header.querySelector('.nav-links');
    if (!navLinks) return;

    // Check if toggle button already exists, if not create one
    var toggleBtn = header.querySelector('.nav-toggle');
    if (!toggleBtn) {
      toggleBtn = document.createElement('button');
      toggleBtn.className = 'nav-toggle';
      toggleBtn.type = 'button';
      toggleBtn.setAttribute('aria-label', 'เปิดเมนูนำทาง');
      toggleBtn.setAttribute('aria-expanded', 'false');
      toggleBtn.setAttribute('aria-controls', 'siteNav');
      toggleBtn.innerHTML =
        '<span class="nav-toggle-bar"></span>' +
        '<span class="nav-toggle-bar"></span>' +
        '<span class="nav-toggle-bar"></span>';

      container.appendChild(toggleBtn);
    }

    function toggleMenu(forceClose) {
      var isOpen = forceClose ? false : !navLinks.classList.contains('is-open');
      navLinks.classList.toggle('is-open', isOpen);
      toggleBtn.classList.toggle('is-active', isOpen);
      toggleBtn.setAttribute('aria-expanded', isOpen ? 'true' : 'false');
      toggleBtn.setAttribute('aria-label', isOpen ? 'ปิดเมนูนำทาง' : 'เปิดเมนูนำทาง');
    }

    toggleBtn.addEventListener('click', function (e) {
      e.stopPropagation();
      toggleMenu();
    });

    // Close menu when clicking outside
    document.addEventListener('click', function (e) {
      if (navLinks.classList.contains('is-open')) {
        if (!header.contains(e.target)) {
          toggleMenu(true);
        }
      }
    });

    // Close on Escape
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape' && navLinks.classList.contains('is-open')) {
        toggleMenu(true);
      }
    });

    // Close when clicking any nav link
    var links = navLinks.querySelectorAll('a');
    for (var i = 0; i < links.length; i++) {
      links[i].addEventListener('click', function () {
        toggleMenu(true);
      });
    }
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initNav);
  } else {
    initNav();
  }
})();
