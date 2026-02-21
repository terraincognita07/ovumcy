(function () {
    var section = document.querySelector("[data-export-section]");
    if (!section) return;

    var rawMinDate = section.getAttribute("data-export-min") || "";
    var rawMaxDate = section.getAttribute("data-export-max") || "";
    var successMessage = section.getAttribute("data-export-success") || "Data exported successfully";
    var failedMessage = section.getAttribute("data-export-failed") || "Failed to export data";
    var invalidRangeMessage = section.getAttribute("data-export-invalid-range") || "End date must be on or after start date";
    var invalidDateMessage = section.getAttribute("data-export-invalid-date") || "Use YYYY-MM-DD";
    var openCalendarLabel = section.getAttribute("data-export-open-calendar") || "Open calendar";
    var jumpTitle = section.getAttribute("data-export-jump-title") || "Choose month and year";
    var summaryTotalTemplate = section.getAttribute("data-export-summary-total-template") || "Total entries: %d";
    var summaryRangeTemplate = section.getAttribute("data-export-summary-range-template") || "Date range: %s to %s";
    var summaryRangeEmpty = section.getAttribute("data-export-summary-range-empty") || "Date range: -";
    var links = section.querySelectorAll("a[data-export-link]");
    var presetButtons = section.querySelectorAll("button[data-export-preset]");
    var fromInput = section.querySelector("input[data-export-from]");
    var toInput = section.querySelector("input[data-export-to]");
    var summaryTotalNode = section.querySelector("[data-export-summary-total]");
    var summaryRangeNode = section.querySelector("[data-export-summary-range]");
    var calendarPanel = section.querySelector("[data-export-calendar-panel]");
    var calendarTitle = section.querySelector("[data-export-calendar-title]");
    var calendarTitleToggle = section.querySelector("[data-export-calendar-title-toggle]");
    var calendarActive = section.querySelector("[data-export-calendar-active]");
    var calendarJump = section.querySelector("[data-export-calendar-jump]");
    var calendarMonth = section.querySelector("[data-export-calendar-month]");
    var calendarYear = section.querySelector("[data-export-calendar-year]");
    var calendarApply = section.querySelector("[data-export-calendar-apply]");
    var calendarWeekdays = section.querySelector("[data-export-calendar-weekdays]");
    var calendarDays = section.querySelector("[data-export-calendar-days]");
    var calendarPrev = section.querySelector("[data-export-calendar-prev]");
    var calendarNext = section.querySelector("[data-export-calendar-next]");
    var calendarClose = section.querySelector("[data-export-calendar-close]");
    var locale = (document.documentElement.getAttribute("lang") || "").toLowerCase().indexOf("ru") === 0 ? "ru-RU" : "en-US";
    var monthFormatter = new Intl.DateTimeFormat(locale, { month: "long", year: "numeric" });
    var weekdayFormatter = new Intl.DateTimeFormat(locale, { weekday: "short" });
    var monthNameFormatter = new Intl.DateTimeFormat(locale, { month: "long" });
    var monthNames = [];
    var activeInput = null;
    var visibleMonth = null;
    var summaryTimer = 0;
    var summaryRequestID = 0;
    if (!links.length || !fromInput || !toInput) return;

    for (var monthIndex = 0; monthIndex < 12; monthIndex++) {
      monthNames.push(monthNameFormatter.format(new Date(2024, monthIndex, 1)));
    }

    if (calendarMonth) {
      calendarMonth.innerHTML = "";
      for (var optionMonth = 0; optionMonth < 12; optionMonth++) {
        var option = document.createElement("option");
        option.value = String(optionMonth);
        option.textContent = monthNames[optionMonth];
        calendarMonth.appendChild(option);
      }
    }

    function padNumber(value) {
      if (value < 10) {
        return "0" + String(value);
      }
      return String(value);
    }

    function formatISODate(value) {
      if (!value) return "";
      return [
        String(value.getFullYear()),
        padNumber(value.getMonth() + 1),
        padNumber(value.getDate())
      ].join("-");
    }

    function parseISODate(raw) {
      var normalized = String(raw || "").trim();
      if (!normalized) return null;
      var match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(normalized);
      if (!match) return null;

      var year = Number(match[1]);
      var month = Number(match[2]) - 1;
      var day = Number(match[3]);
      var parsed = new Date(year, month, day);
      if (
        parsed.getFullYear() !== year ||
        parsed.getMonth() !== month ||
        parsed.getDate() !== day
      ) {
        return null;
      }
      return parsed;
    }

    function sanitizeDateInputValue(input) {
      if (!input) return;
      var digits = String(input.value || "").replace(/\D/g, "").slice(0, 8);
      var year = digits.slice(0, 4);
      var month = digits.slice(4, 6);
      var day = digits.slice(6, 8);

      if (month.length === 2) {
        var monthNumber = Number(month);
        if (monthNumber < 1) {
          month = "01";
        } else if (monthNumber > 12) {
          month = "12";
        } else {
          month = monthNumber < 10 ? "0" + String(monthNumber) : String(monthNumber);
        }
      }

      if (day.length === 2) {
        var dayNumber = Number(day);
        if (dayNumber < 1) {
          day = "01";
        } else if (dayNumber > 31) {
          day = "31";
        } else {
          day = dayNumber < 10 ? "0" + String(dayNumber) : String(dayNumber);
        }
      }

      var normalized = year;
      if (month.length > 0) {
        normalized += "-" + month;
      }
      if (day.length > 0) {
        normalized += "-" + day;
      }

      if (normalized !== input.value) {
        input.value = normalized;
      }
    }

    function formatTemplate(template, values) {
      var index = 0;
      return String(template || "").replace(/%[sd]/g, function () {
        var value = index < values.length ? values[index] : "";
        index += 1;
        return String(value);
      });
    }

    function cloneDate(value) {
      return new Date(value.getFullYear(), value.getMonth(), value.getDate());
    }

    function dateKey(value) {
      return Number(formatISODate(value).replace(/-/g, ""));
    }

    function toMonthStart(value) {
      return new Date(value.getFullYear(), value.getMonth(), 1);
    }

    function monthEnd(value) {
      return new Date(value.getFullYear(), value.getMonth() + 1, 0);
    }

    var minBound = parseISODate(rawMinDate);
    var maxBound = parseISODate(rawMaxDate);
    var hasBounds = !!(minBound && maxBound && dateKey(minBound) <= dateKey(maxBound));

    function isSameDay(left, right) {
      if (!left || !right) return false;
      return dateKey(left) === dateKey(right);
    }

    function isWithinBounds(value) {
      if (!hasBounds || !value) return true;
      var key = dateKey(value);
      return key >= dateKey(minBound) && key <= dateKey(maxBound);
    }

    function monthIntersectsBounds(monthValue) {
      if (!hasBounds) return true;
      var start = toMonthStart(monthValue);
      var end = monthEnd(monthValue);
      return dateKey(end) >= dateKey(minBound) && dateKey(start) <= dateKey(maxBound);
    }

    function clampMonthToBounds(monthValue) {
      if (!monthValue) {
        return hasBounds ? toMonthStart(maxBound) : toMonthStart(new Date());
      }
      var normalized = toMonthStart(monthValue);
      if (!hasBounds || monthIntersectsBounds(normalized)) {
        return normalized;
      }
      if (dateKey(normalized) < dateKey(toMonthStart(minBound))) {
        return toMonthStart(minBound);
      }
      return toMonthStart(maxBound);
    }

    function inputLabel(input) {
      if (!input || !input.id) return "";
      var label = section.querySelector('label[for="' + input.id + '"]');
      if (!label) return "";
      return String(label.textContent || "").trim();
    }

    function renderWeekdayLabels() {
      if (!calendarWeekdays) return;
      calendarWeekdays.innerHTML = "";

      for (var weekday = 0; weekday < 7; weekday++) {
        var sample = new Date(2023, 0, 1 + weekday);
        var label = weekdayFormatter.format(sample).replace(/\./g, "");
        var cell = document.createElement("span");
        cell.textContent = label;
        calendarWeekdays.appendChild(cell);
      }
    }

    function setButtonDisabled(button, disabled) {
      if (!button) return;
      button.disabled = disabled;
      button.setAttribute("aria-disabled", disabled ? "true" : "false");
    }

    function setExportLinksDisabled(disabled) {
      for (var index = 0; index < links.length; index++) {
        var link = links[index];
        link.classList.toggle("export-link-disabled", disabled);
        link.setAttribute("aria-disabled", disabled ? "true" : "false");
      }
    }

    function updateSummaryText(totalEntries, hasData, dateFrom, dateTo) {
      if (summaryTotalNode) {
        summaryTotalNode.textContent = formatTemplate(summaryTotalTemplate, [Number(totalEntries) || 0]);
      }
      if (!summaryRangeNode) return;

      if (hasData && dateFrom && dateTo) {
        summaryRangeNode.textContent = formatTemplate(summaryRangeTemplate, [dateFrom, dateTo]);
      } else {
        summaryRangeNode.textContent = summaryRangeEmpty;
      }
    }

    function buildSummaryEndpoint() {
      var url = new URL("/api/export/summary", window.location.origin);
      if (fromInput.value) {
        url.searchParams.set("from", fromInput.value);
      }
      if (toInput.value) {
        url.searchParams.set("to", toInput.value);
      }
      return url.toString();
    }

    function scheduleSummaryRefresh() {
      if (!hasBounds) return;
      if (summaryTimer) {
        window.clearTimeout(summaryTimer);
      }
      summaryTimer = window.setTimeout(function () {
        refreshSummary();
      }, 160);
    }

    async function refreshSummary() {
      if (!hasBounds) return;
      if (!validateExportRange("summary")) return;

      var requestID = ++summaryRequestID;
      var headers = {};
      var currentLang = (document.documentElement.getAttribute("lang") || "").trim();
      if (currentLang) {
        headers["Accept-Language"] = currentLang;
      }

      try {
        var response = await fetch(buildSummaryEndpoint(), {
          credentials: "same-origin",
          headers: headers
        });
        if (!response.ok) {
          throw new Error("summary_failed");
        }
        var payload = await response.json();
        if (requestID != summaryRequestID) return;
        updateSummaryText(payload.total_entries, payload.has_data, payload.date_from, payload.date_to);
      } catch {
        // Keep previous summary values if refresh fails.
      }
    }

    function syncJumpControls() {
      if (!visibleMonth) return;

      if (calendarYear) {
        if (hasBounds) {
          calendarYear.min = String(minBound.getFullYear());
          calendarYear.max = String(maxBound.getFullYear());
        } else {
          calendarYear.min = "1900";
          calendarYear.max = "2200";
        }
        calendarYear.value = String(visibleMonth.getFullYear());
      }

      if (calendarMonth) {
        calendarMonth.value = String(visibleMonth.getMonth());
        var jumpYear = visibleMonth.getFullYear();
        for (var index = 0; index < calendarMonth.options.length; index++) {
          var option = calendarMonth.options[index];
          option.disabled = hasBounds && !monthIntersectsBounds(new Date(jumpYear, index, 1));
        }
      }

      if (calendarApply && calendarMonth && calendarYear) {
        var yearValue = Number(calendarYear.value);
        var monthValue = Number(calendarMonth.value);
        var candidate = new Date(yearValue, monthValue, 1);
        var invalidCandidate = Number.isNaN(yearValue) || Number.isNaN(monthValue);
        setButtonDisabled(calendarApply, invalidCandidate || (hasBounds && !monthIntersectsBounds(candidate)));
      }
    }

    function updateNavButtons() {
      if (!visibleMonth) return;
      if (calendarPrev) {
        var previousMonth = new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() - 1, 1);
        setButtonDisabled(calendarPrev, hasBounds && !monthIntersectsBounds(previousMonth));
      }
      if (calendarNext) {
        var nextMonth = new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() + 1, 1);
        setButtonDisabled(calendarNext, hasBounds && !monthIntersectsBounds(nextMonth));
      }
    }

    function computePresetRange(rawPreset) {
      if (!hasBounds) return null;
      var preset = String(rawPreset || "").trim().toLowerCase();
      if (preset == "all") {
        return { from: cloneDate(minBound), to: cloneDate(maxBound) };
      }
      var days = Number(preset);
      if (!Number.isFinite(days) || days < 1) return null;

      var toDate = cloneDate(maxBound);
      var fromDate = new Date(toDate.getFullYear(), toDate.getMonth(), toDate.getDate() - days + 1);
      return { from: fromDate, to: toDate };
    }

    function updatePresetState() {
      if (!presetButtons.length) return;
      var fromDate = parseISODate(fromInput.value);
      var toDate = parseISODate(toInput.value);

      for (var index = 0; index < presetButtons.length; index++) {
        var button = presetButtons[index];
        var presetValue = button.getAttribute("data-export-preset") || "";
        var range = computePresetRange(presetValue);
        var active = !!(range && fromDate && toDate && isSameDay(fromDate, range.from) && isSameDay(toDate, range.to));

        setButtonDisabled(button, !hasBounds);
        button.classList.toggle("btn-primary", active);
        button.classList.toggle("btn-soft", !active);
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
        input.setCustomValidity(invalidDateMessage);
        return { ok: false, date: null };
      }

      input.value = formatISODate(parsed);
      input.setCustomValidity("");
      return { ok: true, date: parsed };
    }

    function validateExportRange(changedSide) {
      var fromResult = parseAndNormalizeInput(fromInput);
      var toResult = parseAndNormalizeInput(toInput);
      if (!fromResult.ok || !toResult.ok) {
        setExportLinksDisabled(true);
        return false;
      }

      var fromDate = fromResult.date;
      var toDate = toResult.date;
      if (hasBounds) {
        if (!fromDate) {
          fromDate = cloneDate(minBound);
          fromInput.value = formatISODate(fromDate);
        }
        if (!toDate) {
          toDate = cloneDate(maxBound);
          toInput.value = formatISODate(toDate);
        }
      }
      if (fromDate && toDate && dateKey(toDate) < dateKey(fromDate)) {
        if (changedSide == "to") {
          toInput.value = formatISODate(fromDate);
          toDate = fromDate;
        } else {
          fromInput.value = formatISODate(toDate);
          fromDate = toDate;
        }
      }

      fromInput.setCustomValidity("");
      toInput.setCustomValidity("");
      if (fromDate && toDate && dateKey(toDate) < dateKey(fromDate)) {
        toInput.setCustomValidity(invalidRangeMessage);
        setExportLinksDisabled(true);
        return false;
      }
      setExportLinksDisabled(false);
      return true;
    }

    function buildExportEndpoint(baseEndpoint) {
      var url = new URL(baseEndpoint, window.location.origin);
      if (fromInput.value) {
        url.searchParams.set("from", fromInput.value);
      }
      if (toInput.value) {
        url.searchParams.set("to", toInput.value);
      }
      return url.toString();
    }

    function closeCalendar() {
      if (!calendarPanel) return;
      calendarPanel.classList.add("hidden");
      if (calendarJump) {
        calendarJump.classList.add("hidden");
      }
      activeInput = null;
    }

    function toggleCalendarJump() {
      if (!calendarJump) return;
      calendarJump.classList.toggle("hidden");
      if (!calendarJump.classList.contains("hidden") && calendarYear) {
        calendarYear.focus();
      }
    }

    function syncInitialRange() {
      if (!hasBounds) return;

      var fromValue = parseISODate(fromInput.value);
      var toValue = parseISODate(toInput.value);
      fromValue = fromValue || cloneDate(minBound);
      toValue = toValue || cloneDate(maxBound);

      fromInput.value = formatISODate(fromValue);
      toInput.value = formatISODate(toValue);
      validateExportRange("init");
    }

    function renderCalendar() {
      if (!calendarPanel || !calendarTitle || !calendarDays || !activeInput || !visibleMonth) return;
      if (!hasBounds) {
        closeCalendar();
        return;
      }

      visibleMonth = clampMonthToBounds(visibleMonth);
      calendarPanel.classList.remove("hidden");
      calendarTitle.textContent = monthFormatter.format(visibleMonth);
      if (calendarActive) {
        calendarActive.textContent = inputLabel(activeInput);
      }

      renderWeekdayLabels();
      syncJumpControls();
      updateNavButtons();
      calendarDays.innerHTML = "";

      var year = visibleMonth.getFullYear();
      var month = visibleMonth.getMonth();
      var firstWeekday = new Date(year, month, 1).getDay();
      var daysInMonth = new Date(year, month + 1, 0).getDate();
      var selectedDate = parseISODate(activeInput.value);

      for (var blank = 0; blank < firstWeekday; blank++) {
        var placeholder = document.createElement("span");
        placeholder.className = "h-2";
        calendarDays.appendChild(placeholder);
      }

      for (var day = 1; day <= daysInMonth; day++) {
        var dayDate = new Date(year, month, day);
        var button = document.createElement("button");
        button.type = "button";
        button.textContent = String(day);
        button.className = "btn-secondary text-sm export-calendar-day-button";

        var isAllowedDay = isWithinBounds(dayDate);
        if (!isAllowedDay) {
          button.disabled = true;
          button.className = "btn-soft text-sm export-calendar-day-button export-calendar-day-disabled";
        } else {
          (function (selectedDay) {
            button.addEventListener("click", function () {
              if (!activeInput) return;
              activeInput.value = formatISODate(selectedDay);
              validateExportRange(activeInput === toInput ? "to" : "from");
              updatePresetState();
              scheduleSummaryRefresh();
              closeCalendar();
            });
          })(dayDate);
        }

        if (selectedDate && dateKey(selectedDate) === dateKey(dayDate)) {
          button.className = "btn-primary text-sm export-calendar-day-button";
        }

        calendarDays.appendChild(button);
      }
    }

    function openCalendarForInput(input) {
      if (!calendarPanel || !input) return;
      if (!hasBounds) return;
      activeInput = input;
      var selectedValue = parseISODate(input.value);
      var reference = selectedValue || cloneDate(maxBound);
      visibleMonth = clampMonthToBounds(reference);
      renderCalendar();
    }

    function parseFilenameFromDisposition(disposition, fallbackName) {
      if (!disposition) return fallbackName;
      var match = disposition.match(/filename\*?=(?:UTF-8'')?"?([^";]+)"?/i);
      if (!match || !match[1]) return fallbackName;
      try {
        return decodeURIComponent(match[1]);
      } catch {
        return match[1];
      }
    }

    fromInput.title = openCalendarLabel;
    toInput.title = openCalendarLabel;
    if (calendarTitleToggle) {
      calendarTitleToggle.title = jumpTitle;
    }

    if (!hasBounds) {
      fromInput.disabled = true;
      toInput.disabled = true;
      fromInput.value = "";
      toInput.value = "";
      setButtonDisabled(calendarPrev, true);
      setButtonDisabled(calendarNext, true);
      setButtonDisabled(calendarApply, true);
      setButtonDisabled(calendarTitleToggle, true);
      updatePresetState();
      setExportLinksDisabled(false);
    } else {
      syncInitialRange();
      updatePresetState();
      setExportLinksDisabled(false);
      scheduleSummaryRefresh();
    }

    fromInput.addEventListener("input", function () {
      sanitizeDateInputValue(fromInput);
      validateExportRange("from");
      updatePresetState();
      scheduleSummaryRefresh();
    });
    fromInput.addEventListener("blur", function () {
      validateExportRange("from");
      updatePresetState();
      scheduleSummaryRefresh();
    });

    toInput.addEventListener("input", function () {
      sanitizeDateInputValue(toInput);
      validateExportRange("to");
      updatePresetState();
      scheduleSummaryRefresh();
    });
    toInput.addEventListener("blur", function () {
      validateExportRange("to");
      updatePresetState();
      scheduleSummaryRefresh();
    });

    fromInput.addEventListener("focus", function () {
      openCalendarForInput(fromInput);
    });
    fromInput.addEventListener("click", function () {
      openCalendarForInput(fromInput);
    });

    toInput.addEventListener("focus", function () {
      openCalendarForInput(toInput);
    });
    toInput.addEventListener("click", function () {
      openCalendarForInput(toInput);
    });

    for (var presetIndex = 0; presetIndex < presetButtons.length; presetIndex++) {
      (function (button) {
        button.addEventListener("click", function () {
          var presetValue = button.getAttribute("data-export-preset") || "";
          var range = computePresetRange(presetValue);
          if (!range) return;
          fromInput.value = formatISODate(range.from);
          toInput.value = formatISODate(range.to);
          fromInput.dispatchEvent(new Event("input", { bubbles: true }));
          toInput.dispatchEvent(new Event("input", { bubbles: true }));
        });
      })(presetButtons[presetIndex]);
    }

    if (calendarTitleToggle) {
      calendarTitleToggle.addEventListener("click", toggleCalendarJump);
    }

    if (calendarMonth) {
      calendarMonth.addEventListener("change", syncJumpControls);
    }
    if (calendarYear) {
      calendarYear.addEventListener("input", syncJumpControls);
    }

    if (calendarPrev) {
      calendarPrev.addEventListener("click", function () {
        if (!visibleMonth || !activeInput) return;
        var previousMonth = new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() - 1, 1);
        if (hasBounds && !monthIntersectsBounds(previousMonth)) return;
        visibleMonth = previousMonth;
        renderCalendar();
      });
    }

    if (calendarNext) {
      calendarNext.addEventListener("click", function () {
        if (!visibleMonth || !activeInput) return;
        var nextMonth = new Date(visibleMonth.getFullYear(), visibleMonth.getMonth() + 1, 1);
        if (hasBounds && !monthIntersectsBounds(nextMonth)) return;
        visibleMonth = nextMonth;
        renderCalendar();
      });
    }

    if (calendarApply) {
      calendarApply.addEventListener("click", function () {
        if (!calendarMonth || !calendarYear || !activeInput) return;
        var monthValue = Number(calendarMonth.value);
        var yearValue = Number(String(calendarYear.value || "").trim());
        if (Number.isNaN(monthValue) || monthValue < 0 || monthValue > 11) return;
        if (Number.isNaN(yearValue) || yearValue < 1900 || yearValue > 2200) return;
        visibleMonth = clampMonthToBounds(new Date(yearValue, monthValue, 1));
        renderCalendar();
        if (calendarJump) {
          calendarJump.classList.add("hidden");
        }
      });
    }

    if (calendarYear) {
      calendarYear.addEventListener("keydown", function (event) {
        if (event.key === "Enter" && calendarApply) {
          event.preventDefault();
          calendarApply.click();
        }
      });
    }

    if (calendarClose) {
      calendarClose.addEventListener("click", closeCalendar);
    }

    document.addEventListener("click", function (event) {
      if (!calendarPanel || calendarPanel.classList.contains("hidden")) return;
      var target = event.target;
      if (calendarPanel.contains(target)) return;
      if (target === fromInput || target === toInput) return;
      closeCalendar();
    });

    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        closeCalendar();
      }
    });

    async function handleExport(event) {
      event.preventDefault();
      var link = event.currentTarget;
      var baseEndpoint = link.getAttribute("href");
      if (!baseEndpoint) return;

      if (!validateExportRange("export")) {
        if (fromInput && fromInput.validationMessage) {
          fromInput.reportValidity();
        } else if (toInput) {
          toInput.reportValidity();
        }
        if (typeof window.showToast === "function") {
          var message = invalidRangeMessage;
          if (fromInput && fromInput.validationMessage) {
            message = fromInput.validationMessage;
          } else if (toInput && toInput.validationMessage) {
            message = toInput.validationMessage;
          }
          window.showToast(message, "error");
        }
        return;
      }

      var endpoint = buildExportEndpoint(baseEndpoint);
      var type = (link.getAttribute("data-export-type") || "csv").toLowerCase();

      link.classList.add("btn-loading");
      link.setAttribute("aria-disabled", "true");

      try {
        var headers = {};
        var currentLang = (document.documentElement.getAttribute("lang") || "").trim();
        if (currentLang) {
          headers["Accept-Language"] = currentLang;
        }

        var response = await fetch(endpoint, {
          credentials: "same-origin",
          headers: headers
        });
        if (!response.ok) {
          throw new Error("request_failed");
        }

        var blob = await response.blob();
        var extension = type === "json" ? "json" : "csv";
        var fallbackName = "lume-export." + extension;
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
        }, 500);

        if (typeof window.showToast === "function") {
          window.showToast(successMessage, "success");
        }
      } catch {
        if (typeof window.showToast === "function") {
          window.showToast(failedMessage, "error");
        }
      } finally {
        link.classList.remove("btn-loading");
        link.removeAttribute("aria-disabled");
      }
    }

    links.forEach(function (link) {
      link.addEventListener("click", handleExport);
    });
  })();
