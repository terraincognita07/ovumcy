(function () {
  "use strict";

  var CHART_SELECTOR = "[data-chart]";
  var RESIZE_DEBOUNCE_MS = 140;
  var MAX_VISIBLE_LABELS = 10;

  function isFiniteNumber(value) {
    return typeof value === "number" && isFinite(value);
  }

  function toFiniteNumber(value) {
    var numeric = Number(value);
    return isFiniteNumber(numeric) ? numeric : null;
  }

  function toText(value) {
    if (value === null || value === undefined) {
      return "";
    }
    return String(value);
  }

  function cssVar(name, fallback) {
    var raw = getComputedStyle(document.documentElement).getPropertyValue(name);
    var value = raw ? raw.trim() : "";
    return value || fallback;
  }

  function parseChartData(container) {
    var raw = container.getAttribute("data-chart");
    if (!raw) {
      return null;
    }

    try {
      var parsed = JSON.parse(raw);
      if (!parsed || typeof parsed !== "object") {
        return null;
      }

      var labelsSource = Array.isArray(parsed.labels) ? parsed.labels : [];
      var valuesSource = Array.isArray(parsed.values) ? parsed.values : [];
      var values = [];
      var labels = [];

      for (var index = 0; index < valuesSource.length; index++) {
        var numeric = toFiniteNumber(valuesSource[index]);
        if (!isFiniteNumber(numeric)) {
          continue;
        }
        values.push(numeric);

        var label = index < labelsSource.length ? toText(labelsSource[index]).trim() : "";
        labels.push(label || String(index + 1));
      }

      var baseline = toFiniteNumber(parsed.baseline);
      if (!isFiniteNumber(baseline) || baseline <= 0) {
        baseline = null;
      }

      return {
        labels: labels,
        values: values,
        baseline: baseline
      };
    } catch (_) {
      return null;
    }
  }

  function renderMessage(container, text) {
    container.textContent = "";
    var content = document.createElement("div");
    content.className = "flex h-full items-center justify-center text-sm journal-muted";
    content.textContent = text;
    container.appendChild(content);
  }

  function getContainerSize(container) {
    var width = Math.max(240, Math.floor(container.clientWidth || 640));
    var height = Math.max(190, Math.floor(container.clientHeight || 280));
    return { width: width, height: height };
  }

  function createCanvas(container, size) {
    var canvas = document.createElement("canvas");
    var context = canvas.getContext("2d");
    if (!context) {
      return null;
    }
    var dpr = Math.max(1, window.devicePixelRatio || 1);

    canvas.width = Math.floor(size.width * dpr);
    canvas.height = Math.floor(size.height * dpr);
    canvas.style.width = "100%";
    canvas.style.height = "100%";
    canvas.style.display = "block";
    container.appendChild(canvas);

    context.scale(dpr, dpr);

    return {
      canvas: canvas,
      context: context
    };
  }

  function createDomain(values, baseline) {
    var rangeValues = values.slice();
    if (isFiniteNumber(baseline)) {
      rangeValues.push(baseline);
    }

    if (!rangeValues.length) {
      return null;
    }

    var minValue = Math.min.apply(null, rangeValues);
    var maxValue = Math.max.apply(null, rangeValues);

    if (minValue === maxValue) {
      minValue -= 1;
      maxValue += 1;
    }

    return {
      min: minValue,
      max: maxValue
    };
  }

  function formatDays(value, daySuffix) {
    return String(Math.round(value)) + daySuffix;
  }

  function drawGrid(context, padding, width, height, color) {
    context.strokeStyle = color;
    context.lineWidth = 1;
    context.beginPath();

    for (var row = 0; row < 4; row++) {
      var y = padding.top + (height / 3) * row;
      context.moveTo(padding.left, y);
      context.lineTo(padding.left + width, y);
    }

    context.stroke();
  }

  function drawBaseline(context, padding, width, yForValue, baseline, baselineLabel, daySuffix, color) {
    var baselineY = yForValue(baseline);

    context.save();
    context.setLineDash([6, 4]);
    context.strokeStyle = color;
    context.lineWidth = 2;
    context.beginPath();
    context.moveTo(padding.left, baselineY);
    context.lineTo(padding.left + width, baselineY);
    context.stroke();
    context.restore();

    context.fillStyle = color;
    context.font = "10px Quicksand, Nunito, sans-serif";
    context.textAlign = "right";
    context.textBaseline = "bottom";
    context.fillText(
      baselineLabel + " " + formatDays(baseline, daySuffix),
      padding.left + width - 8,
      Math.max(padding.top + 12, baselineY - 6)
    );
  }

  function drawValueLine(context, values, xForIndex, yForValue, color) {
    if (!values.length) {
      return;
    }

    context.strokeStyle = color;
    context.lineWidth = 3;
    context.beginPath();

    for (var index = 0; index < values.length; index++) {
      var x = xForIndex(index);
      var y = yForValue(values[index]);
      if (index === 0) {
        context.moveTo(x, y);
      } else {
        context.lineTo(x, y);
      }
    }

    context.stroke();
  }

  function drawValuePoints(context, values, xForIndex, yForValue, color) {
    context.fillStyle = color;

    for (var index = 0; index < values.length; index++) {
      context.beginPath();
      context.arc(xForIndex(index), yForValue(values[index]), 4.2, 0, Math.PI * 2);
      context.fill();
    }
  }

  function drawXLabels(context, labels, xForIndex, canvasHeight, padding, color) {
    context.fillStyle = color;
    context.font = "12px Quicksand, Nunito, sans-serif";
    context.textAlign = "center";
    context.textBaseline = "top";

    if (!labels.length) {
      return;
    }

    var step = Math.max(1, Math.ceil(labels.length / MAX_VISIBLE_LABELS));
    var lastDrawnIndex = -1;
    for (var index = 0; index < labels.length; index += step) {
      context.fillText(labels[index], xForIndex(index), canvasHeight - padding.bottom + 10);
      lastDrawnIndex = index;
    }

    var lastIndex = labels.length - 1;
    if (lastDrawnIndex !== lastIndex) {
      context.fillText(labels[lastIndex], xForIndex(lastIndex), canvasHeight - padding.bottom + 10);
    }
  }

  function drawYLabels(context, domain, padding, height, daySuffix, color) {
    context.fillStyle = color;
    context.font = "12px Quicksand, Nunito, sans-serif";
    context.textAlign = "right";
    context.textBaseline = "middle";
    context.fillText(formatDays(domain.max, daySuffix), padding.left - 8, padding.top + 2);
    context.fillText(formatDays(domain.min, daySuffix), padding.left - 8, padding.top + height);
  }

  function drawChart(container) {
    if (!container) {
      return;
    }

    var emptyText = container.getAttribute("data-empty-text") || "Not enough cycle data yet.";
    var daySuffix = container.getAttribute("data-days-suffix") || "d";
    var baselineLabel = container.getAttribute("data-baseline-label") || "Baseline";
    var chartData = parseChartData(container);

    container.textContent = "";

    if (!chartData) {
      renderMessage(container, "Unable to render chart.");
      return;
    }

    var hasBaseline = isFiniteNumber(chartData.baseline);
    if (!chartData.values.length && !hasBaseline) {
      renderMessage(container, emptyText);
      return;
    }

    var size = getContainerSize(container);
    var canvasBundle = createCanvas(container, size);
    if (!canvasBundle) {
      renderMessage(container, "Unable to render chart.");
      return;
    }
    var context = canvasBundle.context;
    var padding = { top: 26, right: 22, bottom: 40, left: 46 };
    var innerWidth = size.width - padding.left - padding.right;
    var innerHeight = size.height - padding.top - padding.bottom;
    var domain = createDomain(chartData.values, chartData.baseline);

    if (!domain) {
      renderMessage(container, emptyText);
      return;
    }

    var xForIndex = function (index) {
      if (chartData.values.length <= 1) {
        return padding.left + innerWidth / 2;
      }
      return padding.left + (index * innerWidth) / (chartData.values.length - 1);
    };

    var yForValue = function (value) {
      var ratio = (value - domain.min) / (domain.max - domain.min);
      return padding.top + innerHeight - ratio * innerHeight;
    };

    var colors = {
      grid: cssVar("--chart-grid", "rgba(172, 136, 96, 0.26)"),
      line: cssVar("--chart-line", "#c4895a"),
      dot: cssVar("--chart-dot", "#b9753e"),
      baseline: cssVar("--chart-baseline", "#9f8a75"),
      label: cssVar("--text-muted", "#9b8b7a")
    };

    context.clearRect(0, 0, size.width, size.height);
    drawGrid(context, padding, innerWidth, innerHeight, colors.grid);

    if (hasBaseline) {
      drawBaseline(context, padding, innerWidth, yForValue, chartData.baseline, baselineLabel, daySuffix, colors.baseline);
    }

    drawValueLine(context, chartData.values, xForIndex, yForValue, colors.line);
    drawValuePoints(context, chartData.values, xForIndex, yForValue, colors.dot);
    drawXLabels(context, chartData.labels, xForIndex, size.height, padding, colors.label);
    drawYLabels(context, domain, padding, innerHeight, daySuffix, colors.label);
  }

  function renderCharts(root) {
    var scope = root && root.querySelectorAll ? root : document;
    if (scope !== document && scope.matches && scope.matches(CHART_SELECTOR)) {
      drawChart(scope);
    }

    var charts = scope.querySelectorAll(CHART_SELECTOR);
    for (var index = 0; index < charts.length; index++) {
      drawChart(charts[index]);
    }
  }

  var resizeTimer = null;
  function scheduleRender() {
    if (resizeTimer !== null) {
      clearTimeout(resizeTimer);
    }
    resizeTimer = setTimeout(function () {
      renderCharts(document);
    }, RESIZE_DEBOUNCE_MS);
  }

  window.addEventListener("DOMContentLoaded", function () {
    renderCharts(document);
  });

  window.addEventListener("resize", scheduleRender);

  document.body.addEventListener("htmx:afterSwap", function (event) {
    var target = event && event.detail ? event.detail.target : null;
    renderCharts(target || document);
  });
})();
