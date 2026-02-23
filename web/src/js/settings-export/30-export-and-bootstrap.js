  function bindRangeInput(input, side, onRangeChanged) {
    input.addEventListener("input", function () {
      sanitizeDateInputValue(input);
      onRangeChanged(side);
    });
    input.addEventListener("blur", function () {
      onRangeChanged(side);
    });
  }

  function createExportHandler(context, rangeController) {
    return async function handleExport(event) {
      event.preventDefault();
      var link = event.currentTarget;
      var baseEndpoint = link.getAttribute("href");
      if (!baseEndpoint) {
        return;
      }

      if (!rangeController.validate("export")) {
        if (context.fromInput.validationMessage) {
          context.fromInput.reportValidity();
        } else if (context.toInput) {
          context.toInput.reportValidity();
        }

        if (typeof window.showToast === "function") {
          var message = context.invalidRangeMessage;
          if (context.fromInput.validationMessage) {
            message = context.fromInput.validationMessage;
          } else if (context.toInput.validationMessage) {
            message = context.toInput.validationMessage;
          }
          window.showToast(message, "error");
        }
        return;
      }

      var endpoint = rangeController.buildExportEndpoint(baseEndpoint);
      var type = (link.getAttribute("data-export-type") || "csv").toLowerCase();

      link.classList.add("btn-loading");
      link.setAttribute("aria-disabled", "true");

      try {
        var response = await fetch(endpoint, {
          credentials: "same-origin",
          headers: buildAcceptLanguageHeaders()
        });
        if (!response.ok) {
          throw new Error("request_failed");
        }

        var blob = await response.blob();
        var extension = type === "json" ? "json" : "csv";
        var fallbackName = "ovumcy-export." + extension;
        var filename = parseFilenameFromDisposition(response.headers.get("Content-Disposition") || "", fallbackName);

        var objectURL = URL.createObjectURL(blob);
        var downloadLink = document.createElement("a");
        downloadLink.href = objectURL;
        downloadLink.download = filename;
        document.body.appendChild(downloadLink);
        downloadLink.click();
        downloadLink.remove();
        window.setTimeout(function () {
          URL.revokeObjectURL(objectURL);
        }, DOWNLOAD_REVOKE_DELAY_MS);

        if (typeof window.showToast === "function") {
          window.showToast(context.successMessage, "success");
        }
      } catch {
        if (typeof window.showToast === "function") {
          window.showToast(context.failedMessage, "error");
        }
      } finally {
        link.classList.remove("btn-loading");
        link.removeAttribute("aria-disabled");
      }
    };
  }
  var section = document.querySelector("[data-export-section]");
  if (!section) {
    return;
  }

  var context = createContext(section);
  if (!context) {
    return;
  }

  var bounds = createBounds(context.rawMinDate, context.rawMaxDate);
  var rangeController = createDateRangeController(context, bounds);
  var summaryController = createSummaryController(context, bounds, rangeController);

  function onRangeChanged(side) {
    rangeController.validate(side);
    rangeController.updatePresetState();
    summaryController.scheduleRefresh();
  }

  var calendarController = createCalendarController(context, bounds, onRangeChanged);

  context.fromInput.title = context.openCalendarLabel;
  context.toInput.title = context.openCalendarLabel;
  if (context.calendarTitleToggle) {
    context.calendarTitleToggle.title = context.jumpTitle;
  }

  if (!bounds.hasBounds) {
    context.fromInput.disabled = true;
    context.toInput.disabled = true;
    context.fromInput.value = "";
    context.toInput.value = "";
    calendarController.disableControls();
    rangeController.updatePresetState();
    rangeController.setExportLinksDisabled(false);
  } else {
    rangeController.syncInitialRange();
    rangeController.updatePresetState();
    rangeController.setExportLinksDisabled(false);
    summaryController.scheduleRefresh();
  }

  bindRangeInput(context.fromInput, "from", onRangeChanged);
  bindRangeInput(context.toInput, "to", onRangeChanged);

  context.fromInput.addEventListener("focus", function () {
    calendarController.openCalendarForInput(context.fromInput);
  });
  context.fromInput.addEventListener("click", function () {
    calendarController.openCalendarForInput(context.fromInput);
  });

  context.toInput.addEventListener("focus", function () {
    calendarController.openCalendarForInput(context.toInput);
  });
  context.toInput.addEventListener("click", function () {
    calendarController.openCalendarForInput(context.toInput);
  });

  for (var presetIndex = 0; presetIndex < context.presetButtons.length; presetIndex++) {
    (function (button) {
      button.addEventListener("click", function () {
        var presetValue = button.getAttribute("data-export-preset") || "";
        if (rangeController.applyPreset(presetValue)) {
          summaryController.scheduleRefresh();
        }
      });
    })(context.presetButtons[presetIndex]);
  }

  if (context.calendarTitleToggle) {
    context.calendarTitleToggle.addEventListener("click", calendarController.toggleCalendarJump);
  }
  if (context.calendarMonth) {
    context.calendarMonth.addEventListener("change", calendarController.syncJumpControls);
  }
  if (context.calendarYear) {
    context.calendarYear.addEventListener("input", calendarController.syncJumpControls);
    context.calendarYear.addEventListener("keydown", calendarController.onYearKeydown);
  }
  if (context.calendarPrev) {
    context.calendarPrev.addEventListener("click", function () {
      calendarController.moveMonth(-1);
    });
  }
  if (context.calendarNext) {
    context.calendarNext.addEventListener("click", function () {
      calendarController.moveMonth(1);
    });
  }
  if (context.calendarApply) {
    context.calendarApply.addEventListener("click", calendarController.applyJumpSelection);
  }
  if (context.calendarClose) {
    context.calendarClose.addEventListener("click", calendarController.closeCalendar);
  }

  document.addEventListener("click", function (event) {
    if (!context.calendarPanel || context.calendarPanel.classList.contains("hidden")) {
      return;
    }
    var target = event.target;
    if (!target) {
      return;
    }
    if (context.calendarPanel.contains(target)) {
      return;
    }
    if (target === context.fromInput || target === context.toInput) {
      return;
    }
    calendarController.closeCalendar();
  });

  document.addEventListener("keydown", function (event) {
    if (event.key === "Escape") {
      calendarController.closeCalendar();
    }
  });

  var handleExport = createExportHandler(context, rangeController);
  for (var linkIndex = 0; linkIndex < context.links.length; linkIndex++) {
    context.links[linkIndex].addEventListener("click", handleExport);
  }
})();

