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
