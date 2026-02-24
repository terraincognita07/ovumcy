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

  var dayEditorAutoSaveTimers = new WeakMap();

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

  function dayEditorAutosaveFieldName(target) {
    if (!target || typeof target.getAttribute !== "function") {
      return "";
    }
    var name = String(target.getAttribute("name") || "").trim();
    if (name === "is_period" || name === "flow" || name === "symptom_ids" || name === "notes") {
      return name;
    }
    return "";
  }

  function submitDayEditorForm(form) {
    if (!form || !document.body.contains(form)) {
      return;
    }
    if (window.htmx && typeof window.htmx.trigger === "function") {
      window.htmx.trigger(form, "submit");
      return;
    }
    if (typeof form.requestSubmit === "function") {
      form.requestSubmit();
      return;
    }
    form.submit();
  }

  function queueDayEditorAutosave(form, delayMs) {
    if (!form) {
      return;
    }

    var wait = Number(delayMs);
    if (!Number.isFinite(wait) || wait < 0) {
      wait = 0;
    }

    var existingTimer = dayEditorAutoSaveTimers.get(form);
    if (existingTimer) {
      window.clearTimeout(existingTimer);
    }

    var timer = window.setTimeout(function () {
      dayEditorAutoSaveTimers.delete(form);
      if (form.classList.contains("htmx-request")) {
        queueDayEditorAutosave(form, 120);
        return;
      }
      submitDayEditorForm(form);
    }, wait);
    dayEditorAutoSaveTimers.set(form, timer);
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

    document.body.addEventListener("change", function (event) {
      var target = getEventTarget(event);
      if (!target || !target.closest) {
        return;
      }
      var fieldName = dayEditorAutosaveFieldName(target);
      if (!fieldName) {
        return;
      }
      var form = target.closest("form[data-day-editor-autosave]");
      if (!form) {
        return;
      }
      var delayMs = 0;
      if (fieldName === "symptom_ids") {
        delayMs = 120;
      } else if (fieldName === "notes") {
        delayMs = 220;
      }
      queueDayEditorAutosave(form, delayMs);
    });

    document.body.addEventListener("input", function (event) {
      var target = getEventTarget(event);
      if (!target || !target.closest || target.tagName !== "TEXTAREA") {
        return;
      }
      if (dayEditorAutosaveFieldName(target) !== "notes") {
        return;
      }
      var form = target.closest("form[data-day-editor-autosave]");
      if (!form) {
        return;
      }
      queueDayEditorAutosave(form, 700);
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
