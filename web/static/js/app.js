(function () {
  "use strict";

  var PASSWORD_HIDE_ICON = "\u{1F648}";
  var PASSWORD_SHOW_ICON = "\u{1F441}";
  var TOAST_VISIBLE_MS = 2200;
  var TOAST_EXIT_MS = 220;
  var STATUS_CLEAR_MS = 2000;
  var DOWNLOAD_REVOKE_MS = 500;

  function onDocumentReady(callback) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", callback);
      return;
    }
    callback();
  }

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
    } catch (_) {
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
    applyHTMLLanguage(readCookie("lume_lang") || document.documentElement.getAttribute("lang"));

    var links = document.querySelectorAll("a.lang-link");
    for (var index = 0; index < links.length; index++) {
      var link = links[index];
      var updatedHref = withCurrentNextPath(link.getAttribute("href"));
      if (updatedHref) {
        link.setAttribute("href", updatedHref);
      }
    }

    document.addEventListener("click", function (event) {
      var target = event.target;
      if (!target || !target.closest) {
        return;
      }

      var link = target.closest("a.lang-link");
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
      var target = event.target;
      if (!target || !target.closest) {
        return;
      }

      var link = target.closest("a[data-auth-switch]");
      if (!link) {
        return;
      }

      if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
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

  function updatePasswordToggleLabel(button, isVisible) {
    var showLabel = button.getAttribute("data-show-label") || "Show password";
    var hideLabel = button.getAttribute("data-hide-label") || "Hide password";
    button.setAttribute("aria-label", isVisible ? hideLabel : showLabel);
    button.textContent = isVisible ? PASSWORD_HIDE_ICON : PASSWORD_SHOW_ICON;
  }

  function attachPasswordToggles(root) {
    var scope = root && root.querySelectorAll ? root : document;
    var buttons = scope.querySelectorAll("[data-password-toggle]");

    for (var index = 0; index < buttons.length; index++) {
      var button = buttons[index];
      if (button.dataset.passwordToggleBound === "1") {
        continue;
      }

      var field = button.parentElement ? button.parentElement.querySelector("input[type='password'], input[type='text']") : null;
      if (!field) {
        continue;
      }

      button.dataset.passwordToggleBound = "1";
      updatePasswordToggleLabel(button, field.type === "text");

      button.addEventListener("click", (function (input, toggleButton) {
        return function () {
          var reveal = input.type === "password";
          input.type = reveal ? "text" : "password";
          updatePasswordToggleLabel(toggleButton, reveal);
        };
      })(field, button));
    }
  }

  function initPasswordToggles() {
    attachPasswordToggles(document);
    document.body.addEventListener("htmx:afterSwap", function (event) {
      var target = event && event.detail ? event.detail.target : null;
      attachPasswordToggles(target || document);
    });
  }

  function initLoginValidation() {
    var form = document.getElementById("login-form");
    if (!form) {
      return;
    }

    var requiredMessage = form.getAttribute("data-required-message") || "Please fill out this field.";
    var emailMessage = form.getAttribute("data-email-message") || "Please enter a valid email address.";

    function updateValidityMessage(input) {
      if (!input || typeof input.setCustomValidity !== "function") {
        return;
      }

      input.setCustomValidity("");
      if (!input.validity) {
        return;
      }

      if (input.validity.valueMissing) {
        input.setCustomValidity(requiredMessage);
        return;
      }
      if (input.type === "email" && input.validity.typeMismatch) {
        input.setCustomValidity(emailMessage);
      }
    }

    var fields = form.querySelectorAll("input[required]");
    for (var index = 0; index < fields.length; index++) {
      fields[index].addEventListener("invalid", function () {
        updateValidityMessage(this);
      });
      fields[index].addEventListener("input", function () {
        this.setCustomValidity("");
      });
      fields[index].addEventListener("blur", function () {
        updateValidityMessage(this);
      });
    }
  }

  function initConfirmModal() {
    var modal = document.getElementById("confirm-modal");
    var messageNode = document.getElementById("confirm-modal-message");
    var cancelButton = document.getElementById("confirm-modal-cancel");
    var acceptButton = document.getElementById("confirm-modal-accept");
    if (!modal || !messageNode || !cancelButton || !acceptButton) {
      return;
    }

    var pendingResolve = null;

    function closeConfirm(accepted) {
      if (!pendingResolve) {
        return;
      }
      var resolve = pendingResolve;
      pendingResolve = null;
      modal.classList.add("hidden");
      modal.setAttribute("aria-hidden", "true");
      resolve(accepted);
    }

    function openConfirm(question, acceptLabel) {
      if (pendingResolve) {
        pendingResolve(false);
        pendingResolve = null;
      }

      messageNode.textContent = question || "";
      cancelButton.textContent = document.body.getAttribute("data-confirm-cancel") || "Cancel";
      acceptButton.textContent = acceptLabel || document.body.getAttribute("data-confirm-delete") || "Delete";
      modal.classList.remove("hidden");
      modal.setAttribute("aria-hidden", "false");
      cancelButton.focus();

      return new Promise(function (resolve) {
        pendingResolve = resolve;
      });
    }

    cancelButton.addEventListener("click", function () {
      closeConfirm(false);
    });

    acceptButton.addEventListener("click", function () {
      closeConfirm(true);
    });

    modal.addEventListener("click", function (event) {
      if (event.target === modal) {
        closeConfirm(false);
      }
    });

    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        closeConfirm(false);
      }
    });

    document.body.addEventListener("htmx:confirm", function (event) {
      if (!event || !event.detail || !event.detail.question) {
        return;
      }

      var source = event.detail.elt || event.target;
      if (!source || !source.getAttribute) {
        return;
      }

      var acceptLabel = source.getAttribute("data-confirm-accept") || "";
      event.preventDefault();
      openConfirm(event.detail.question, acceptLabel).then(function (confirmed) {
        if (confirmed) {
          event.detail.issueRequest(true);
        }
      });
    });

    document.addEventListener("submit", function (event) {
      var form = event.target;
      if (!form || !form.matches || !form.matches("form[data-confirm]")) {
        return;
      }

      if (form.dataset.confirmBypass === "1") {
        form.dataset.confirmBypass = "";
        return;
      }

      event.preventDefault();
      openConfirm(form.getAttribute("data-confirm") || "", form.getAttribute("data-confirm-accept") || "").then(function (confirmed) {
        if (!confirmed) {
          return;
        }
        form.dataset.confirmBypass = "1";
        if (typeof form.requestSubmit === "function") {
          form.requestSubmit();
          return;
        }
        form.submit();
      });
    });
  }

  function renderErrorStatus(target, text) {
    target.textContent = "";
    var block = document.createElement("div");
    block.className = "status-error";
    block.textContent = text;
    target.appendChild(block);
  }

  function createToastStack() {
    var stack = document.createElement("div");
    stack.className = "toast-stack";
    document.body.appendChild(stack);
    return stack;
  }

  function initToastAPI() {
    var stack = null;

    function getStack() {
      if (stack) {
        return stack;
      }
      stack = createToastStack();
      return stack;
    }

    window.showToast = function (message, kind) {
      if (!message) {
        return;
      }

      var container = getStack();
      var toast = document.createElement("div");
      toast.className = (kind === "error" ? "status-error" : "status-ok") + " reveal";
      toast.textContent = message;
      container.appendChild(toast);

      window.setTimeout(function () {
        toast.classList.add("toast-exit");
        window.setTimeout(function () {
          toast.remove();
        }, TOAST_EXIT_MS);
      }, TOAST_VISIBLE_MS);
    };
  }

  function initHTMXHooks() {
    document.body.addEventListener("htmx:configRequest", function (event) {
      var tokenMeta = document.querySelector('meta[name="csrf-token"]');
      if (!tokenMeta || !event || !event.detail) {
        return;
      }

      var token = tokenMeta.getAttribute("content");
      if (!token) {
        return;
      }

      event.detail.parameters = event.detail.parameters || {};
      event.detail.parameters.csrf_token = token;
      event.detail.headers = event.detail.headers || {};
      event.detail.headers["X-CSRF-Token"] = token;
    });

    document.body.addEventListener("htmx:beforeRequest", function (event) {
      var source = event && event.target && event.target.closest ? event.target : null;
      var form = source ? source.closest("form[data-save-feedback]") : null;
      if (!form) {
        return;
      }

      var button = form.querySelector("[data-save-button]");
      if (!button) {
        return;
      }

      button.disabled = true;
      button.setAttribute("aria-busy", "true");
      button.classList.add("btn-loading");
    });

    document.body.addEventListener("htmx:afterRequest", function (event) {
      var source = event && event.target && event.target.closest ? event.target : null;
      var form = source ? source.closest("form[data-save-feedback]") : null;
      if (!form) {
        return;
      }

      var button = form.querySelector("[data-save-button]");
      if (!button) {
        return;
      }

      button.disabled = false;
      button.removeAttribute("aria-busy");
      button.classList.remove("btn-loading");
    });

    document.body.addEventListener("htmx:afterSwap", function (event) {
      var target = event && event.detail ? event.detail.target : null;
      if (!target || !target.classList || !target.classList.contains("save-status")) {
        return;
      }

      var successNode = target.querySelector(".status-ok");
      if (!successNode) {
        return;
      }

      var dayEditor = document.getElementById("day-editor");
      var form = target.closest("form[data-save-feedback]");
      if (dayEditor && form && form.closest("#day-editor")) {
        if (window.htmx && typeof window.htmx.trigger === "function") {
          window.htmx.trigger(document.body, "calendar-day-updated");
        }

        var postPath = form.getAttribute("hx-post") || "";
        var match = postPath.match(/\/api\/days\/(\d{4}-\d{2}-\d{2})$/);
        if (match && window.htmx && typeof window.htmx.ajax === "function") {
          window.htmx.ajax("GET", "/calendar/day/" + match[1], { target: "#day-editor", swap: "innerHTML" });
        }
      }

      window.setTimeout(function () {
        if (target.querySelector(".status-ok")) {
          target.textContent = "";
        }
      }, STATUS_CLEAR_MS);
    });

    document.body.addEventListener("htmx:responseError", function (event) {
      var target = event && event.detail ? event.detail.target : null;
      if (!target || !target.classList || !target.classList.contains("save-status")) {
        return;
      }

      var xhr = event.detail.xhr;
      var responseText = xhr && typeof xhr.responseText === "string" ? xhr.responseText : "";
      if (responseText && responseText.indexOf("status-error") !== -1) {
        target.innerHTML = responseText;
        return;
      }

      var fallback = document.body.getAttribute("data-request-failed") || "Request failed. Please try again.";
      renderErrorStatus(target, fallback);
    });
  }

  function writeTextToClipboard(text) {
    if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
      return navigator.clipboard.writeText(text);
    }

    return new Promise(function (resolve, reject) {
      var textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.setAttribute("readonly", "readonly");
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.select();

      try {
        var copied = document.execCommand("copy");
        document.body.removeChild(textarea);
        if (copied) {
          resolve();
          return;
        }
      } catch (_) {
        document.body.removeChild(textarea);
      }

      reject(new Error("copy_failed"));
    });
  }

  window.calendarView = function (config) {
    var safeConfig = config || {};
    return {
      selectedDate: safeConfig.selectedDate || "",
      selectDay: function (value) {
        this.selectedDate = value || "";
        if (!this.selectedDate || !window.history || typeof window.history.replaceState !== "function") {
          return;
        }

        try {
          var currentURL = new URL(window.location.href);
          currentURL.searchParams.set("day", this.selectedDate);
          var nextPath = currentURL.pathname + currentURL.search + currentURL.hash;
          window.history.replaceState({}, "", nextPath);
        } catch (_) {}
      }
    };
  };

  window.onboardingFlow = function (config) {
    var safeConfig = config || {};
    var lang = safeConfig.lang || "en";

    return {
      step: 0,
      minDate: safeConfig.minDate || "",
      maxDate: safeConfig.maxDate || "",
      selectedDate: safeConfig.lastPeriodStart || "",
      periodStatus: safeConfig.onboardingPeriodStatus || "",
      periodEndDate: safeConfig.onboardingPeriodEnd || "",
      cycleLength: Number(safeConfig.cycleLength || 28),
      periodLength: Number(safeConfig.periodLength || 5),
      autoPeriodFill: safeConfig.autoPeriodFill !== false,
      dayOptions: [],
      endDayOptions: [],
      init: function () {
        this.dayOptions = this.buildDayOptions(this.minDate, this.maxDate, lang);
        this.onStartDateChanged();
      },
      setStartDate: function (value) {
        this.selectedDate = value || "";
        this.onStartDateChanged();
      },
      setPeriodStatus: function (value) {
        this.periodStatus = value || "";
        if (this.periodStatus !== "finished") {
          this.periodEndDate = "";
        }
        this.refreshEndDayOptions();
      },
      onStartDateChanged: function () {
        if (!this.selectedDate) {
          this.periodStatus = "";
          this.periodEndDate = "";
          this.endDayOptions = [];
          return;
        }
        this.refreshEndDayOptions();
      },
      refreshEndDayOptions: function () {
        if (!this.selectedDate) {
          this.endDayOptions = [];
          this.periodEndDate = "";
          return;
        }
        this.endDayOptions = this.buildDayOptions(this.selectedDate, this.maxDate, lang);
        if (!this.isDateWithinRange(this.periodEndDate, this.selectedDate, this.maxDate)) {
          this.periodEndDate = "";
        }
      },
      isDateWithinRange: function (value, minRaw, maxRaw) {
        if (!value) {
          return false;
        }
        var current = this.parseDate(value);
        var min = this.parseDate(minRaw);
        var max = this.parseDate(maxRaw);
        if (!current || !min || !max) {
          return false;
        }
        return current >= min && current <= max;
      },
      buildDayOptions: function (minDateRaw, maxDateRaw, locale) {
        var minDate = this.parseDate(minDateRaw);
        var maxDate = this.parseDate(maxDateRaw);
        if (!minDate || !maxDate || minDate > maxDate) {
          return [];
        }

        var result = [];
        var formatter = new Intl.DateTimeFormat(locale || "en", {
          day: "numeric",
          month: "short"
        });

        for (var cursor = new Date(maxDate); cursor >= minDate; cursor.setDate(cursor.getDate() - 1)) {
          var current = new Date(cursor);
          result.push({
            value: this.formatDateValue(current),
            label: formatter.format(current)
          });
        }
        return result;
      },
      parseDate: function (value) {
        if (!value) {
          return null;
        }
        var parsed = new Date(value + "T00:00:00");
        if (isNaN(parsed.getTime())) {
          return null;
        }
        return parsed;
      },
      formatDateValue: function (value) {
        var year = value.getFullYear();
        var month = String(value.getMonth() + 1).padStart(2, "0");
        var day = String(value.getDate()).padStart(2, "0");
        return year + "-" + month + "-" + day;
      }
    };
  };

  window.recoveryCodeTools = function () {
    return {
      copied: false,
      downloaded: false,
      downloadFailed: false,
      copyCode: function () {
        var code = this.$refs.code ? this.$refs.code.textContent.trim() : "";
        if (!code) {
          return;
        }

        var self = this;
        writeTextToClipboard(code).then(function () {
          self.copied = true;
          self.downloaded = false;
          self.downloadFailed = false;
          window.setTimeout(function () {
            self.copied = false;
          }, STATUS_CLEAR_MS);
        }).catch(function () {
          self.downloadFailed = true;
          window.setTimeout(function () {
            self.downloadFailed = false;
          }, STATUS_CLEAR_MS);
        });
      },
      downloadCode: function () {
        var code = this.$refs.code ? this.$refs.code.textContent.trim() : "";
        if (!code) {
          return;
        }

        var self = this;
        this.copied = false;
        this.downloaded = false;
        this.downloadFailed = false;

        try {
          var content = "Lume recovery code\n\n" + code + "\n\nStore this code offline and private.";
          var blob = new Blob([content], { type: "text/plain;charset=utf-8" });
          var objectURL = URL.createObjectURL(blob);
          var link = document.createElement("a");
          link.href = objectURL;
          link.download = "lume-recovery-code.txt";
          document.body.appendChild(link);
          link.click();
          link.remove();

          window.setTimeout(function () {
            URL.revokeObjectURL(objectURL);
          }, DOWNLOAD_REVOKE_MS);

          self.downloaded = true;
          window.setTimeout(function () {
            self.downloaded = false;
          }, STATUS_CLEAR_MS);
        } catch (_) {
          self.downloadFailed = true;
          window.setTimeout(function () {
            self.downloadFailed = false;
          }, STATUS_CLEAR_MS);
        }
      }
    };
  };

  onDocumentReady(function () {
    initAuthPanelTransitions();
    initLanguageSwitcher();
    initPasswordToggles();
    initLoginValidation();
    initConfirmModal();
    initToastAPI();
    initHTMXHooks();
  });
})();
