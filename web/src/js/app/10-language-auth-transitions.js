  function parseLanguage(raw) {
    if (!raw) {
      return "";
    }
    var normalized = String(raw).trim().toLowerCase().replace(/_/g, "-");
    if (!normalized) {
      return "";
    }
    if (normalized.indexOf("-") !== -1) {
      normalized = normalized.split("-")[0];
    }
    if (normalized !== "en" && normalized !== "ru") {
      return "";
    }
    return normalized;
  }

  function readCookie(name) {
    var cookies = document.cookie ? document.cookie.split(";") : [];
    for (var index = 0; index < cookies.length; index++) {
      var part = cookies[index].trim();
      if (part.indexOf(name + "=") !== 0) {
        continue;
      }
      return decodeURIComponent(part.substring(name.length + 1));
    }
    return "";
  }

  function languageFromHref(href) {
    if (!href) {
      return "";
    }
    var match = href.match(/\/lang\/([^/?#]+)/i);
    if (!match || !match[1]) {
      return "";
    }
    return match[1];
  }

  function withCurrentNextPath(href) {
    if (!href) {
      return href;
    }
    try {
      var url = new URL(href, window.location.origin);
      var nextPath = window.location.pathname + window.location.search;
      url.searchParams.set("next", nextPath);
      return url.pathname + url.search + url.hash;
    } catch {
      return href;
    }
  }

  function applyHTMLLanguage(raw) {
    var lang = parseLanguage(raw);
    if (!lang) {
      return;
    }
    document.documentElement.setAttribute("lang", lang);
  }

  function initLanguageSwitcher() {
    applyHTMLLanguage(readCookie("ovumcy_lang") || document.documentElement.getAttribute("lang"));

    var links = document.querySelectorAll("a.lang-link");
    for (var index = 0; index < links.length; index++) {
      var link = links[index];
      var updatedHref = withCurrentNextPath(link.getAttribute("href"));
      if (updatedHref) {
        link.setAttribute("href", updatedHref);
      }
    }

    document.addEventListener("click", function (event) {
      var link = closestFromEvent(event, "a.lang-link");
      if (!link) {
        return;
      }

      var updatedHref = withCurrentNextPath(link.getAttribute("href"));
      if (updatedHref) {
        link.setAttribute("href", updatedHref);
      }
      applyHTMLLanguage(languageFromHref(updatedHref || link.getAttribute("href")));
    });
  }

  function initAuthPanelTransitions() {
    var panel = document.querySelector("[data-auth-panel]");
    if (!panel) {
      return;
    }

    var prefersReducedMotion = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (!prefersReducedMotion) {
      panel.style.opacity = "0";
      panel.style.transform = "translateY(8px)";
      panel.style.transition = "opacity 180ms ease, transform 180ms ease";
      window.requestAnimationFrame(function () {
        panel.style.opacity = "1";
        panel.style.transform = "translateY(0)";
      });
    }

    document.addEventListener("click", function (event) {
      var link = closestFromEvent(event, "a[data-auth-switch]");
      if (!link) {
        return;
      }

      if (event.defaultPrevented || !isPrimaryClick(event)) {
        return;
      }
      if (link.getAttribute("target") === "_blank") {
        return;
      }

      var href = (link.getAttribute("href") || "").trim();
      if (!href || prefersReducedMotion) {
        return;
      }

      event.preventDefault();
      panel.style.pointerEvents = "none";
      panel.style.opacity = "0";
      panel.style.transform = "translateY(-6px)";
      window.setTimeout(function () {
        window.location.href = href;
      }, 140);
    });
  }

