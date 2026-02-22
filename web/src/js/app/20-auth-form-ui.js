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
