  function createContext(section) {
    var locale = (document.documentElement.getAttribute("lang") || "").toLowerCase().indexOf("ru") === 0 ? "ru-RU" : "en-US";
    var monthFormatter = new Intl.DateTimeFormat(locale, { month: "long", year: "numeric" });
    var weekdayFormatter = new Intl.DateTimeFormat(locale, { weekday: "short" });
    var monthNameFormatter = new Intl.DateTimeFormat(locale, { month: "long" });

    var context = {
      section: section,
      rawMinDate: readTextAttribute(section, "data-export-min", ""),
      rawMaxDate: readTextAttribute(section, "data-export-max", ""),
      successMessage: readTextAttribute(section, "data-export-success", "Data exported successfully"),
      failedMessage: readTextAttribute(section, "data-export-failed", "Failed to export data"),
      invalidRangeMessage: readTextAttribute(section, "data-export-invalid-range", "End date must be on or after start date"),
      invalidDateMessage: readTextAttribute(section, "data-export-invalid-date", "Use YYYY-MM-DD"),
      openCalendarLabel: readTextAttribute(section, "data-export-open-calendar", "Open calendar"),
      jumpTitle: readTextAttribute(section, "data-export-jump-title", "Choose month and year"),
      summaryTotalTemplate: readTextAttribute(section, "data-export-summary-total-template", "Total entries: %d"),
      summaryRangeTemplate: readTextAttribute(section, "data-export-summary-range-template", "Date range: %s to %s"),
      summaryRangeEmpty: readTextAttribute(section, "data-export-summary-range-empty", "Date range: -"),
      links: section.querySelectorAll("a[data-export-link]"),
      presetButtons: section.querySelectorAll("button[data-export-preset]"),
      fromInput: section.querySelector("input[data-export-from]"),
      toInput: section.querySelector("input[data-export-to]"),
      summaryTotalNode: section.querySelector("[data-export-summary-total]"),
      summaryRangeNode: section.querySelector("[data-export-summary-range]"),
      calendarPanel: section.querySelector("[data-export-calendar-panel]"),
      calendarTitle: section.querySelector("[data-export-calendar-title]"),
      calendarTitleToggle: section.querySelector("[data-export-calendar-title-toggle]"),
      calendarActive: section.querySelector("[data-export-calendar-active]"),
      calendarJump: section.querySelector("[data-export-calendar-jump]"),
      calendarMonth: section.querySelector("[data-export-calendar-month]"),
      calendarYear: section.querySelector("[data-export-calendar-year]"),
      calendarApply: section.querySelector("[data-export-calendar-apply]"),
      calendarWeekdays: section.querySelector("[data-export-calendar-weekdays]"),
      calendarDays: section.querySelector("[data-export-calendar-days]"),
      calendarPrev: section.querySelector("[data-export-calendar-prev]"),
      calendarNext: section.querySelector("[data-export-calendar-next]"),
      calendarClose: section.querySelector("[data-export-calendar-close]"),
      monthFormatter: monthFormatter,
      weekdayFormatter: weekdayFormatter,
      monthNames: buildMonthNames(monthNameFormatter)
    };

    if (!context.links.length || !context.fromInput || !context.toInput) {
      return null;
    }
    return context;
  }

  function createDateRangeController(context, bounds) {
    function setExportLinksDisabled(disabled) {
      for (var index = 0; index < context.links.length; index++) {
        var link = context.links[index];
        link.classList.toggle("export-link-disabled", disabled);
        link.setAttribute("aria-disabled", disabled ? "true" : "false");
      }
    }

    function parseAndNormalizeInput(input) {
      sanitizeDateInputValue(input);
      var raw = String(input.value || "").trim();
      input.value = raw;

      if (!raw) {
        input.setCustomValidity("");
        return { ok: true, date: null };
      }

      var parsed = parseISODate(raw);
      if (!parsed) {
        input.setCustomValidity(context.invalidDateMessage);
        return { ok: false, date: null };
      }

      input.value = formatISODate(parsed);
      input.setCustomValidity("");
      return { ok: true, date: parsed };
    }

    function validate(changedSide) {
      var fromResult = parseAndNormalizeInput(context.fromInput);
      var toResult = parseAndNormalizeInput(context.toInput);
      if (!fromResult.ok || !toResult.ok) {
        setExportLinksDisabled(true);
        return false;
      }

      var fromDate = fromResult.date;
      var toDate = toResult.date;

      if (bounds.hasBounds) {
        if (!fromDate) {
          fromDate = cloneDate(bounds.minBound);
          context.fromInput.value = formatISODate(fromDate);
        }
        if (!toDate) {
          toDate = cloneDate(bounds.maxBound);
          context.toInput.value = formatISODate(toDate);
        }
      }

      if (fromDate && toDate && dateKey(toDate) < dateKey(fromDate)) {
        if (changedSide === "to") {
          context.toInput.value = formatISODate(fromDate);
          toDate = fromDate;
        } else {
          context.fromInput.value = formatISODate(toDate);
          fromDate = toDate;
        }
      }

      context.fromInput.setCustomValidity("");
      context.toInput.setCustomValidity("");
      if (fromDate && toDate && dateKey(toDate) < dateKey(fromDate)) {
        context.toInput.setCustomValidity(context.invalidRangeMessage);
        setExportLinksDisabled(true);
        return false;
      }

      setExportLinksDisabled(false);
      return true;
    }

    function computePresetRange(rawPreset) {
      if (!bounds.hasBounds) {
        return null;
      }

      var preset = String(rawPreset || "").trim().toLowerCase();
      if (preset === "all") {
        return { from: cloneDate(bounds.minBound), to: cloneDate(bounds.maxBound) };
      }

      var days = Number(preset);
      if (!Number.isFinite(days) || days < 1) {
        return null;
      }

      var toDate = cloneDate(bounds.maxBound);
      var fromDate = new Date(toDate.getFullYear(), toDate.getMonth(), toDate.getDate() - days + 1);
      return { from: fromDate, to: toDate };
    }

    function updatePresetState() {
      if (!context.presetButtons.length) {
        return;
      }

      var fromDate = parseISODate(context.fromInput.value);
      var toDate = parseISODate(context.toInput.value);

      for (var index = 0; index < context.presetButtons.length; index++) {
        var button = context.presetButtons[index];
        var presetValue = button.getAttribute("data-export-preset") || "";
        var range = computePresetRange(presetValue);
        var active = !!(range && fromDate && toDate && isSameDay(fromDate, range.from) && isSameDay(toDate, range.to));

        setButtonDisabled(button, !bounds.hasBounds);
        button.classList.toggle("btn-primary", active);
        button.classList.toggle("btn-soft", !active);
      }
    }

    function applyPreset(rawPreset) {
      var range = computePresetRange(rawPreset);
      if (!range) {
        return false;
      }
      context.fromInput.value = formatISODate(range.from);
      context.toInput.value = formatISODate(range.to);
      validate("to");
      updatePresetState();
      return true;
    }

    function syncInitialRange() {
      if (!bounds.hasBounds) {
        return;
      }

      var fromValue = parseISODate(context.fromInput.value);
      var toValue = parseISODate(context.toInput.value);
      fromValue = fromValue || cloneDate(bounds.minBound);
      toValue = toValue || cloneDate(bounds.maxBound);

      context.fromInput.value = formatISODate(fromValue);
      context.toInput.value = formatISODate(toValue);
      validate("init");
    }

    return {
      setExportLinksDisabled: setExportLinksDisabled,
      validate: validate,
      updatePresetState: updatePresetState,
      applyPreset: applyPreset,
      syncInitialRange: syncInitialRange,
      buildExportEndpoint: function (baseEndpoint) {
        return buildEndpoint(baseEndpoint, context.fromInput.value, context.toInput.value);
      }
    };
  }
  function createSummaryController(context, bounds, rangeController) {
    var summaryTimer = 0;
    var summaryRequestID = 0;
    var lastSummaryEndpoint = "";
    var summaryAbortController = null;

    function updateSummaryText(totalEntries, hasData, dateFrom, dateTo, selectedFrom, selectedTo) {
      if (context.summaryTotalNode) {
        context.summaryTotalNode.textContent = formatTemplate(context.summaryTotalTemplate, [Number(totalEntries) || 0]);
      }
      if (!context.summaryRangeNode) {
        return;
      }

      var selectedRangeFrom = String(selectedFrom || "").trim();
      var selectedRangeTo = String(selectedTo || "").trim();
      if (selectedRangeFrom && selectedRangeTo) {
        context.summaryRangeNode.textContent = formatTemplate(context.summaryRangeTemplate, [selectedRangeFrom, selectedRangeTo]);
        return;
      }

      if (hasData && dateFrom && dateTo) {
        context.summaryRangeNode.textContent = formatTemplate(context.summaryRangeTemplate, [dateFrom, dateTo]);
      } else {
        context.summaryRangeNode.textContent = context.summaryRangeEmpty;
      }
    }

    function buildSummaryEndpoint() {
      return buildEndpoint(SUMMARY_ENDPOINT, context.fromInput.value, context.toInput.value);
    }

    async function refresh() {
      if (!bounds.hasBounds) {
        return;
      }
      if (!rangeController.validate("summary")) {
        lastSummaryEndpoint = "";
        return;
      }

      var endpoint = buildSummaryEndpoint();
      if (endpoint === lastSummaryEndpoint) {
        return;
      }
      lastSummaryEndpoint = endpoint;

      if (summaryAbortController) {
        summaryAbortController.abort();
      }
      summaryAbortController = typeof AbortController === "function" ? new AbortController() : null;

      var requestID = ++summaryRequestID;
      try {
        var response = await fetch(endpoint, {
          credentials: "same-origin",
          headers: buildAcceptLanguageHeaders(),
          signal: summaryAbortController ? summaryAbortController.signal : undefined
        });
        if (!response.ok) {
          throw new Error("summary_failed");
        }

        var payload = await response.json();
        if (requestID !== summaryRequestID) {
          return;
        }
        updateSummaryText(
          payload.total_entries,
          payload.has_data,
          payload.date_from,
          payload.date_to,
          context.fromInput ? context.fromInput.value : "",
          context.toInput ? context.toInput.value : ""
        );
      } catch (error) {
        if (error && error.name === "AbortError") {
          return;
        }
        // Keep previous summary values if refresh fails.
      } finally {
        if (requestID === summaryRequestID) {
          summaryAbortController = null;
        }
      }
    }

    function scheduleRefresh() {
      if (!bounds.hasBounds) {
        return;
      }
      if (summaryTimer) {
        window.clearTimeout(summaryTimer);
      }
      summaryTimer = window.setTimeout(function () {
        refresh();
      }, SUMMARY_REFRESH_DELAY_MS);
    }

    return {
      scheduleRefresh: scheduleRefresh
    };
  }
