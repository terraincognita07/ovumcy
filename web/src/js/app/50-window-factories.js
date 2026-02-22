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
    var periodExceedsCycleMessage = String(
      safeConfig.periodExceedsCycleMessage || "Period length must not exceed cycle length."
    );

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
      periodExceedsCycleMessage: periodExceedsCycleMessage,
      clearStepStatuses: function () {
        var statusIDs = ["onboarding-step1-status", "onboarding-step2-status", "onboarding-step3-status"];
        for (var index = 0; index < statusIDs.length; index++) {
          var node = document.getElementById(statusIDs[index]);
          if (node) {
            node.textContent = "";
          }
        }
      },
      clearStepStatus: function (statusID) {
        var node = document.getElementById(statusID);
        if (!node) {
          return;
        }
        node.textContent = "";
      },
      renderStepStatus: function (statusID, kind, message) {
        var node = document.getElementById(statusID);
        if (!node) {
          return;
        }
        node.textContent = "";
        if (!message) {
          return;
        }

        var status = document.createElement("div");
        status.className = kind;
        status.textContent = String(message);
        node.appendChild(status);
      },
      normalizeStepTwoValues: function () {
        var cycle = Number(this.cycleLength);
        if (!Number.isFinite(cycle)) {
          cycle = 28;
        }
        cycle = Math.max(15, Math.min(90, Math.round(cycle)));
        this.cycleLength = cycle;

        var period = Number(this.periodLength);
        if (!Number.isFinite(period)) {
          period = 5;
        }
        period = Math.max(1, Math.min(10, Math.round(period)));
        this.periodLength = period;
      },
      onCycleLengthChanged: function () {
        this.normalizeStepTwoValues();
        this.clearStepStatus("onboarding-step2-status");
      },
      onPeriodLengthChanged: function () {
        this.normalizeStepTwoValues();
        this.clearStepStatus("onboarding-step2-status");
      },
      validateStepTwoBeforeSubmit: function (event) {
        this.normalizeStepTwoValues();
        if (this.periodLength > this.cycleLength) {
          if (event && typeof event.preventDefault === "function") {
            event.preventDefault();
          }
          this.renderStepStatus("onboarding-step2-status", "status-error", this.periodExceedsCycleMessage);
          return false;
        }
        this.clearStepStatus("onboarding-step2-status");
        return true;
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
        this.clearStepStatus("onboarding-step1-status");
      },
      setPeriodEndDate: function (value) {
        this.periodEndDate = value || "";
        this.clearStepStatus("onboarding-step1-status");
      },
      setPeriodStatus: function (value) {
        this.periodStatus = value || "";
        if (this.periodStatus !== "finished") {
          this.periodEndDate = "";
        }
        this.refreshEndDayOptions();
        this.clearStepStatus("onboarding-step1-status");
      },
      onStartDateChanged: function () {
        this.clearStepStatus("onboarding-step1-status");
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
      copyFailed: false,
      downloaded: false,
      downloadFailed: false,
      recoveryMessage: function (key, fallback) {
        var root = this.$root;
        if (root && root.dataset && root.dataset[key]) {
          return String(root.dataset[key] || "");
        }
        return String(fallback || "");
      },
      notify: function (key, fallback, kind) {
        var message = this.recoveryMessage(key, fallback);
        if (!message || typeof window.showToast !== "function") {
          return;
        }
        window.showToast(message, kind);
      },
      copyCode: function () {
        var code = getRecoveryCodeText(this.$refs);
        if (!code) {
          return;
        }

        var self = this;
        writeTextToClipboard(code).then(function () {
          self.copied = true;
          self.copyFailed = false;
          self.downloaded = false;
          self.downloadFailed = false;
          self.notify("copySuccessMessage", "Recovery code copied.", "ok");
          setTimedFlag(self, "copied", STATUS_CLEAR_MS);
        }).catch(function () {
          self.copied = false;
          self.copyFailed = true;
          self.downloaded = false;
          self.downloadFailed = false;
          self.notify("copyFailedMessage", "Failed to copy recovery code.", "error");
          setTimedFlag(self, "copyFailed", STATUS_CLEAR_MS);
        });
      },
      downloadCode: function () {
        var code = getRecoveryCodeText(this.$refs);
        if (!code) {
          return;
        }

        var self = this;
        this.copied = false;
        this.copyFailed = false;
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

          self.notify("downloadSuccessMessage", "Recovery code downloaded.", "ok");
          setTimedFlag(self, "downloaded", STATUS_CLEAR_MS);
        } catch {
          self.notify("downloadFailedMessage", "Failed to download recovery code.", "error");
          setTimedFlag(self, "downloadFailed", STATUS_CLEAR_MS);
        }
      }
    };
  };
