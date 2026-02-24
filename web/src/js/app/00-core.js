(function () {
  "use strict";

  var PASSWORD_HIDE_ICON = "\u{1F648}";
  var PASSWORD_SHOW_ICON = "\u{1F441}";
  var TOAST_VISIBLE_MS = 5200;
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
