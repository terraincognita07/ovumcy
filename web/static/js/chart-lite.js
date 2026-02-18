(function () {
  function cssVar(name, fallback) {
    var value = getComputedStyle(document.documentElement).getPropertyValue(name);
    return value ? value.trim() : fallback;
  }

  function drawChart(container) {
    if (!container) return;
    var raw = container.getAttribute("data-chart");
    if (!raw) return;

    var chartData;
    try {
      chartData = JSON.parse(raw);
    } catch (error) {
      container.textContent = "Unable to render chart.";
      return;
    }

    var labels = chartData.labels || [];
    var values = chartData.values || [];
    var baseline = Number(chartData.baseline || 0);
    var hasBaseline = Number.isFinite(baseline) && baseline > 0;
    var emptyText = container.getAttribute("data-empty-text") || "Not enough cycle data yet.";
    var daySuffix = container.getAttribute("data-days-suffix") || "d";
    var baselineLabel = container.getAttribute("data-baseline-label") || "Baseline";
    container.innerHTML = "";

    if (!values.length && !hasBaseline) {
      container.innerHTML = '<div class="flex h-full items-center justify-center text-sm journal-muted">' + emptyText + "</div>";
      return;
    }

    var canvas = document.createElement("canvas");
    canvas.width = container.clientWidth || 640;
    canvas.height = container.clientHeight || 280;
    canvas.style.width = "100%";
    canvas.style.height = "100%";
    container.appendChild(canvas);

    var context = canvas.getContext("2d");
    var padding = { top: 26, right: 22, bottom: 40, left: 46 };
    var width = canvas.width - padding.left - padding.right;
    var height = canvas.height - padding.top - padding.bottom;

    var rangeValues = values.slice();
    if (hasBaseline) {
      rangeValues.push(baseline);
    }

    var maxValue = Math.max.apply(null, rangeValues);
    var minValue = Math.min.apply(null, rangeValues);
    if (maxValue === minValue) {
      maxValue += 1;
      minValue -= 1;
    }

    var gridColor = cssVar("--chart-grid", "rgba(172, 136, 96, 0.26)");
    var lineColor = cssVar("--chart-line", "#c4895a");
    var dotColor = cssVar("--chart-dot", "#b9753e");
    var baselineColor = cssVar("--chart-baseline", "#9f8a75");
    var labelColor = cssVar("--text-muted", "#9b8b7a");

    function x(index) {
      if (values.length === 1) return padding.left + width / 2;
      return padding.left + (index * width) / (values.length - 1);
    }

    function y(value) {
      var ratio = (value - minValue) / (maxValue - minValue);
      return padding.top + height - ratio * height;
    }

    context.clearRect(0, 0, canvas.width, canvas.height);
    context.strokeStyle = gridColor;
    context.lineWidth = 1;

    for (var i = 0; i < 4; i++) {
      var gy = padding.top + (height / 3) * i;
      context.beginPath();
      context.moveTo(padding.left, gy);
      context.lineTo(padding.left + width, gy);
      context.stroke();
    }

    if (hasBaseline) {
      var baselineY = y(baseline);
      context.save();
      context.setLineDash([6, 4]);
      context.strokeStyle = baselineColor;
      context.lineWidth = 2;
      context.beginPath();
      context.moveTo(padding.left, baselineY);
      context.lineTo(padding.left + width, baselineY);
      context.stroke();
      context.restore();

      context.fillStyle = baselineColor;
      context.font = "10px Quicksand, Nunito, sans-serif";
      context.textAlign = "right";
      context.textBaseline = "bottom";
      context.fillText(
        baselineLabel + " " + String(baseline.toFixed(0)) + daySuffix,
        padding.left + width - 8,
        Math.max(padding.top + 12, baselineY - 6)
      );
    }

    if (values.length) {
      context.strokeStyle = lineColor;
      context.lineWidth = 3;
      context.beginPath();
      for (var p = 0; p < values.length; p++) {
        var px = x(p);
        var py = y(values[p]);
        if (p === 0) {
          context.moveTo(px, py);
        } else {
          context.lineTo(px, py);
        }
      }
      context.stroke();
    }

    if (values.length) {
      context.fillStyle = dotColor;
      for (var j = 0; j < values.length; j++) {
        var cx = x(j);
        var cy = y(values[j]);
        context.beginPath();
        context.arc(cx, cy, 4.2, 0, Math.PI * 2);
        context.fill();
      }
    }

    context.fillStyle = labelColor;
    context.font = "12px Quicksand, Nunito, sans-serif";
    context.textAlign = "center";
    context.textBaseline = "top";

    for (var l = 0; l < labels.length; l++) {
      if (labels.length > 10 && l % 2 !== 0) continue;
      context.fillText(labels[l], x(l), canvas.height - padding.bottom + 10);
    }

    context.textAlign = "right";
    context.textBaseline = "middle";
    context.fillText(String(maxValue.toFixed(0)) + daySuffix, padding.left - 8, padding.top + 2);
    context.fillText(String(minValue.toFixed(0)) + daySuffix, padding.left - 8, padding.top + height);
  }

  function mountCharts() {
    var charts = document.querySelectorAll("[data-chart]");
    charts.forEach(drawChart);
  }

  window.addEventListener("DOMContentLoaded", mountCharts);
  window.addEventListener("resize", function () {
    clearTimeout(window.__lumeChartResize);
    window.__lumeChartResize = setTimeout(mountCharts, 120);
  });
})();
