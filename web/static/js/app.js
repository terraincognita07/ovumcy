(function () {
  "use strict";

  var PASSWORD_HIDE_ICON = "\u{1F648}";
  var PASSWORD_SHOW_ICON = "\u{1F441}";
  var TOAST_VISIBLE_MS = 2200;
  var TOAST_EXIT_MS = 220;
  var STATUS_CLEAR_MS = 2000;
  var DOWNLOAD_REVOKE_MS = 500;

  function getEventTarget(event) {
    return event && event.target ? event.target : null;
  }

  function closestFromEvent(event, selector) {
    var target = getEventTarget(event);
    if (!target || !target.closest) {
      return null;
    }
    return target.closest(selector);
  }

  function isPrimaryClick(event) {
    return !!event && event.button === 0 && !event.metaKey && !event.ctrlKey && !event.shiftKey && !event.altKey;
  }

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

  function loginPasswordDraftStorage() {
    if (!window.sessionStorage) {
      return null;
    }
    return window.sessionStorage;
  }

  function readLoginPasswordDraft(storage, key) {
    if (!storage || !key) {
      return "";
    }
    try {
      return String(storage.getItem(key) || "");
    } catch {
      return "";
    }
  }

  function writeLoginPasswordDraft(storage, key, value) {
    if (!storage || !key) {
      return;
    }
    try {
      storage.setItem(key, String(value || ""));
    } catch {
      // Ignore session storage write failures (privacy mode, quota, etc.).
    }
  }

  function clearLoginPasswordDraft(storage, key) {
    if (!storage || !key) {
      return;
    }
    try {
      storage.removeItem(key);
    } catch {
      // Ignore session storage cleanup failures.
    }
  }

  function isTruthyDataValue(raw) {
    var normalized = String(raw || "").trim().toLowerCase();
    return normalized === "1" || normalized === "true" || normalized === "yes";
  }

  function focusLoginPasswordField(input) {
    if (!input || typeof input.focus !== "function") {
      return;
    }
    input.focus();

    if (typeof input.setSelectionRange !== "function") {
      return;
    }
    var end = String(input.value || "").length;
    input.setSelectionRange(end, end);
  }

  function initLoginPasswordPersistence() {
    var form = document.getElementById("login-form");
    if (!form) {
      return;
    }

    var passwordField = document.getElementById("login-password");
    if (!passwordField) {
      return;
    }

    var storage = loginPasswordDraftStorage();
    var storageKey = form.getAttribute("data-password-draft-key") || "lume_login_password_draft";
    var hasError = isTruthyDataValue(form.getAttribute("data-login-has-error"));

    function persistPasswordDraft() {
      writeLoginPasswordDraft(storage, storageKey, passwordField.value);
    }

    passwordField.addEventListener("input", persistPasswordDraft);
    form.addEventListener("submit", persistPasswordDraft);

    if (!hasError) {
      clearLoginPasswordDraft(storage, storageKey);
      return;
    }

    var draft = readLoginPasswordDraft(storage, storageKey);
    if (draft) {
      passwordField.value = draft;
    }
    focusLoginPasswordField(passwordField);
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

  function getSaveFeedbackFormFromEvent(event) {
    var target = getEventTarget(event);
    if (!target || !target.closest) {
      return null;
    }
    return target.closest("form[data-save-feedback]");
  }

  function setSaveButtonState(form, isBusy) {
    if (!form) {
      return;
    }
    var button = form.querySelector("[data-save-button]");
    if (!button) {
      return;
    }

    button.disabled = isBusy;
    if (isBusy) {
      button.setAttribute("aria-busy", "true");
      button.classList.add("btn-loading");
      return;
    }
    button.removeAttribute("aria-busy");
    button.classList.remove("btn-loading");
  }

  function scheduleClearSuccessStatus(target) {
    window.setTimeout(function () {
      if (target.querySelector(".status-ok")) {
        target.textContent = "";
      }
    }, STATUS_CLEAR_MS);
  }

  function maybeRefreshDayEditor(target) {
    var dayEditor = document.getElementById("day-editor");
    var form = target.closest("form[data-save-feedback]");
    if (!dayEditor || !form || !form.closest("#day-editor")) {
      return;
    }

    if (window.htmx && typeof window.htmx.trigger === "function") {
      window.htmx.trigger(document.body, "calendar-day-updated");
    }

    var postPath = form.getAttribute("hx-post") || "";
    var match = postPath.match(/\/api\/days\/(\d{4}-\d{2}-\d{2})$/);
    if (match && window.htmx && typeof window.htmx.ajax === "function") {
      window.htmx.ajax("GET", "/calendar/day/" + match[1], { target: "#day-editor", swap: "innerHTML" });
    }
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
      setSaveButtonState(getSaveFeedbackFormFromEvent(event), true);
    });

    document.body.addEventListener("htmx:afterRequest", function (event) {
      setSaveButtonState(getSaveFeedbackFormFromEvent(event), false);
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

      maybeRefreshDayEditor(target);
      scheduleClearSuccessStatus(target);
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
      } catch {
        document.body.removeChild(textarea);
      }

      reject(new Error("copy_failed"));
    });
  }

  function parseDateValue(value) {
    if (!value) {
      return null;
    }
    var parsed = new Date(value + "T00:00:00");
    if (isNaN(parsed.getTime())) {
      return null;
    }
    return parsed;
  }

  function formatDateValue(value) {
    var year = value.getFullYear();
    var month = String(value.getMonth() + 1).padStart(2, "0");
    var day = String(value.getDate()).padStart(2, "0");
    return year + "-" + month + "-" + day;
  }

  function buildDayOptions(minDateRaw, maxDateRaw, locale) {
    var minDate = parseDateValue(minDateRaw);
    var maxDate = parseDateValue(maxDateRaw);
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
        value: formatDateValue(current),
        label: formatter.format(current)
      });
    }
    return result;
  }

  function isDateWithinRange(value, minRaw, maxRaw) {
    if (!value) {
      return false;
    }
    var current = parseDateValue(value);
    var min = parseDateValue(minRaw);
    var max = parseDateValue(maxRaw);
    if (!current || !min || !max) {
      return false;
    }
    return current >= min && current <= max;
  }

  function setTimedFlag(target, key, timeoutMs) {
    target[key] = true;
    window.setTimeout(function () {
      target[key] = false;
    }, timeoutMs);
  }

  function getRecoveryCodeText(refs) {
    var node = refs && refs.code ? refs.code : null;
    return node ? String(node.textContent || "").trim() : "";
  }

  function collectCheckedSymptomLabels(scope) {
    if (!scope || !scope.querySelectorAll) {
      return [];
    }

    var checked = scope.querySelectorAll("input[name='symptom_ids']:checked");
    var labels = [];
    for (var index = 0; index < checked.length; index++) {
      var label = String(checked[index].dataset.symptomLabel || "").trim();
      if (label) {
        labels.push(label);
      }
    }
    return labels;
  }

  window.appShellState = function () {
    return {
      mobileMenu: false,
      toggleMobileMenu: function () {
        this.mobileMenu = !this.mobileMenu;
      }
    };
  };

  window.settingsCycleForm = function (config) {
    var safeConfig = config || {};
    return {
      cycleLength: Number(safeConfig.cycleLength || 28),
      periodLength: Number(safeConfig.periodLength || 5),
      autoPeriodFill: !!safeConfig.autoPeriodFill
    };
  };

  window.dayEditorForm = function (config) {
    var safeConfig = config || {};
    return {
      isPeriod: !!safeConfig.isPeriod
    };
  };

  window.dashboardTodayEditor = function (config) {
    var safeConfig = config || {};

    return {
      isPeriod: !!safeConfig.isPeriod,
      activeSymptoms: [],
      notesPreview: "",
      syncSymptoms: function () {
        this.activeSymptoms = collectCheckedSymptomLabels(this.$root);
      },
      hasActiveSymptoms: function () {
        return this.activeSymptoms.length > 0;
      },
      hasNotesPreview: function () {
        return String(this.notesPreview || "").trim().length > 0;
      },
      init: function () {
        var notesField = this.$root ? this.$root.querySelector("#today-notes") : null;
        this.notesPreview = notesField ? String(notesField.value || "") : "";
        this.syncSymptoms();
      }
    };
  };

  window.calendarView = function (config) {
    var safeConfig = config || {};
    return {
      selectedDate: safeConfig.selectedDate || "",
      isSelectedDay: function (value) {
        return this.selectedDate === String(value || "");
      },
      selectDayFromEvent: function (event) {
        var target = event && event.currentTarget ? event.currentTarget : null;
        if (!target || typeof target.getAttribute !== "function") {
          return;
        }
        this.selectDay(target.getAttribute("data-day"));
      },
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
        } catch {
          // Ignore malformed URLs and keep current location unchanged.
        }
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
      clearStepStatuses: function () {
        var statusIDs = ["onboarding-step1-status", "onboarding-step2-status", "onboarding-step3-status"];
        for (var index = 0; index < statusIDs.length; index++) {
          var node = document.getElementById(statusIDs[index]);
          if (node) {
            node.textContent = "";
          }
        }
      },
      init: function () {
        this.dayOptions = buildDayOptions(this.minDate, this.maxDate, lang);
        this.onStartDateChanged();
      },
      goToStep: function (value) {
        var nextStep = Number(value);
        if (!Number.isFinite(nextStep)) {
          return;
        }
        this.clearStepStatuses();
        this.step = nextStep;
      },
      begin: function () {
        this.goToStep(1);
      },
      onStepOneSaved: function (event) {
        this.advanceAfterSuccessfulRequest(event, 2);
      },
      onStepTwoSaved: function (event) {
        this.advanceAfterSuccessfulRequest(event, 3);
      },
      advanceAfterSuccessfulRequest: function (event, targetStep) {
        if (!event || !event.detail || !event.detail.successful) {
          return;
        }
        this.goToStep(targetStep);
      },
      setStartDate: function (value) {
        this.selectedDate = value || "";
        this.onStartDateChanged();
      },
      setPeriodEndDate: function (value) {
        this.periodEndDate = value || "";
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
        this.endDayOptions = buildDayOptions(this.selectedDate, this.maxDate, lang);
        if (!isDateWithinRange(this.periodEndDate, this.selectedDate, this.maxDate)) {
          this.periodEndDate = "";
        }
      }
    };
  };

  window.recoveryCodeTools = function () {
    return {
      copied: false,
      downloaded: false,
      downloadFailed: false,
      copyCode: function () {
        var code = getRecoveryCodeText(this.$refs);
        if (!code) {
          return;
        }

        var self = this;
        writeTextToClipboard(code).then(function () {
          self.copied = true;
          self.downloaded = false;
          self.downloadFailed = false;
          setTimedFlag(self, "copied", STATUS_CLEAR_MS);
        }).catch(function () {
          setTimedFlag(self, "downloadFailed", STATUS_CLEAR_MS);
        });
      },
      downloadCode: function () {
        var code = getRecoveryCodeText(this.$refs);
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

          setTimedFlag(self, "downloaded", STATUS_CLEAR_MS);
        } catch {
          setTimedFlag(self, "downloadFailed", STATUS_CLEAR_MS);
        }
      }
    };
  };

  onDocumentReady(function () {
    initAuthPanelTransitions();
    initLanguageSwitcher();
    initPasswordToggles();
    initLoginValidation();
    initLoginPasswordPersistence();
    initConfirmModal();
    initToastAPI();
    initHTMXHooks();
  });
})();
