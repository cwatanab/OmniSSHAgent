<script>
  import { createEventDispatcher } from "svelte";
  import { Button, TextBox, ToggleSwitch, Expander, Checkbox, InfoBar } from "fluent-svelte";
  import { onMount } from "svelte";

  export let lang = "en";

  const t = {
    en: {
      settingsTitle: "Settings",
      settingsStartAtLogin: "Start at Windows login (fast)",
      settingsStartHidden: "Hide the window on launch",
      settingsDebugLog: "Enable debug logging",
      settingsOpenLogDir: "Open log directory",
      settingsBalloon: "Show a balloon notification when an SSH key is used",
      settingsPageant: "Enable Pageant",
      settingsNamedPipe: "Enable Named pipe agent",
      settingsUnix: "Enable Unix domain socket agent",
      settingsUnixPath: "Unix domain socket file path (WSL1):",
      settingsUnixHelper: "Set the path of Unix domain socket file",
      settingsCygwin: "Enable Cygwin unix domain socket agent",
      settingsCygwinPath: "Cygwin Unix domain socket file path (MSYS2):",
      settingsCygwinHelper:
        "Set the path of Cygwin(MSYS2) Unix domain socket file",
      settingsProxy:
        "Enable proxy mode for OpenSSH agent (also compatible with 1Password)",
      settingsLanguage: "Language",
      settingsLanguageAuto: "Auto",
      settingsLanguageJa: "Japanese",
      settingsLanguageEn: "English",
      settingsTheme: "Theme",
      settingsThemeAuto: "Auto",
      settingsThemeDark: "Dark",
      settingsThemeLight: "Light",
      settingsAccent: "Accent color",
      settingsAccentDefault: "Default (Windows)",
      settingsAccentBlue: "Blue",
      settingsAccentIndigo: "Indigo",
      settingsAccentTeal: "Teal",
      settingsAccentEmerald: "Emerald",
      settingsAccentSunset: "Sunset",
      settingsAccentRose: "Rose",
      settingsAccentCustom: "Custom...",
      settingsSave: "Save",
      settingsCancel: "Cancel",
      settingsSaved: "Saved the settings",
      sectionGeneral: "General Settings",
      sectionAgent: "SSH Agent Settings",
    },
    ja: {
      settingsTitle: "設定",
      settingsStartAtLogin: "Windows ログオン時に起動する (高速)",
      settingsStartHidden: "起動時にウィンドウを非表示にする",
      settingsDebugLog: "デバッグログを有効にする",
      settingsOpenLogDir: "ログディレクトリを開く",
      settingsBalloon: "SSH鍵が使用されたときにバルーン通知を表示する",
      settingsPageant: "Pageant を有効にする",
      settingsNamedPipe: "Named pipe エージェントを有効にする",
      settingsUnix: "Unix ドメインソケットエージェントを有効にする",
      settingsUnixPath: "Unix ドメインソケットファイルパス (WSL1):",
      settingsUnixHelper: "Unix ドメインソケットファイルのパスを指定します",
      settingsCygwin: "Cygwin Unix ドメインソケットエージェントを有効にする",
      settingsCygwinPath: "Cygwin Unix ドメインソケットファイルパス (MSYS2):",
      settingsCygwinHelper:
        "Cygwin(MSYS2) Unix ドメインソケットファイルのパスを指定します",
      settingsProxy:
        "OpenSSH エージェントのプロキシモードを有効にする (1Password 互換)",
      settingsLanguage: "言語",
      settingsLanguageAuto: "自動",
      settingsLanguageJa: "日本語",
      settingsLanguageEn: "英語",
      settingsTheme: "テーマ",
      settingsThemeAuto: "自動",
      settingsThemeDark: "ダーク",
      settingsThemeLight: "ライト",
      settingsAccent: "アクセントカラー",
      settingsAccentDefault: "デフォルト (Windows)",
      settingsAccentBlue: "ブルー",
      settingsAccentIndigo: "インディゴ",
      settingsAccentTeal: "ティール",
      settingsAccentEmerald: "エメラルド",
      settingsAccentSunset: "サンセット",
      settingsAccentRose: "ローズ",
      settingsAccentCustom: "カスタム...",
      settingsSave: "保存",
      settingsCancel: "キャンセル",
      settingsSaved: "設定を保存しました",
      sectionGeneral: "一般設定",
      sectionAgent: "SSHエージェント設定",
    },
  };

  const dispatch = createEventDispatcher();

  let infoBarOpen = false;
  let infoBarTitle = "";
  let infoBarMessage = "";
  let infoBarSeverity = "information";

  function showNotification(title, message = "", severity = "information") {
    infoBarTitle = title;
    infoBarMessage = message;
    infoBarSeverity = severity;
    infoBarOpen = true;

    if (severity === "success") {
      setTimeout(() => {
        infoBarOpen = false;
      }, 5000);
    }
  }

  let data = {
    StartHidden: false,
    StartAtLogin: false,
    PageantAgent: false,
    NamedPipeAgent: false,
    UnixSocketAgent: false,
    CygWinAgent: false,
    ShowBalloon: false,
    UnixSocketPath: "",
    CygWinSocketPath: "",
    ProxyModeOfNamedPipe: false,
    DebugLog: false,
  };

  const STORAGE_KEY_LANG = "omni-settings-lang";
  const STORAGE_KEY_THEME = "omni-settings-theme";
  const STORAGE_KEY_ACCENT = "omni-settings-accent";
  let uiLang = "auto";
  let uiTheme = "auto";

  const accentColorsMap = {
    blue: { light: "#0078d4", dark: "#60cdff" },
    indigo: { light: "#4f46e5", dark: "#a78bfa" },
    teal: { light: "#0f766e", dark: "#2dd4bf" },
    emerald: { light: "#047857", dark: "#34d399" },
    sunset: { light: "#c2410c", dark: "#fb923c" },
    rose: { light: "#be123c", dark: "#fb7185" },
  };

  let windowsAccentColor = "#0078d4";
  let customAccentColor = "#0078d4";
  let uiAccent = "default";
  let systemThemeIsLight = true;

  $: accentPalette =
    uiTheme === "auto" ? (systemThemeIsLight ? "light" : "dark") : uiTheme;

  $: effectiveAccentColor = (() => {
    if (uiAccent === "default") return windowsAccentColor;
    if (uiAccent === "custom") return customAccentColor;
    return accentColorsMap[uiAccent]?.[accentPalette] || windowsAccentColor;
  })();

  onMount(async () => {
    try {
      const savedata = await window.go.main.App.GetSettings();
      data = { ...savedata };
    } catch (err) {
      console.error(err);
      showNotification("Error loading settings", err.message || err, "critical");
    }
    const storedLang = localStorage.getItem(STORAGE_KEY_LANG);
    if (storedLang) uiLang = storedLang;
    const storedTheme = localStorage.getItem(STORAGE_KEY_THEME);
    if (storedTheme) uiTheme = storedTheme;

    try {
      const winColor = await window.go.main.App.GetAccentColor();
      if (winColor) windowsAccentColor = winColor;
    } catch (e) {
      console.error("Failed to load Windows accent color:", e);
    }

    try {
      systemThemeIsLight = await window.go.main.App.GetAppsUseLightTheme();
    } catch (e) {
      console.error("Failed to load Windows theme preference:", e);
      systemThemeIsLight = !window.matchMedia("(prefers-color-scheme: dark)")
        .matches;
    }

    const storedAccent = localStorage.getItem(STORAGE_KEY_ACCENT);
    if (storedAccent) {
      if (accentColorsMap[storedAccent]) {
        uiAccent = storedAccent;
      } else if (storedAccent === "default") {
        uiAccent = "default";
      } else {
        uiAccent = "custom";
        customAccentColor = storedAccent;
      }
    }
  });

  const save = async () => {
    try {
      await window.go.main.App.Save(data);
      localStorage.setItem(STORAGE_KEY_LANG, uiLang);
      localStorage.setItem(STORAGE_KEY_THEME, uiTheme);

      let accentToStore = uiAccent;
      if (uiAccent === "custom") {
        accentToStore = customAccentColor;
      }
      localStorage.setItem(STORAGE_KEY_ACCENT, accentToStore);

      showNotification(t[lang].settingsSaved, "", "success");
      dispatch("save");
    } catch (err) {
      console.error(err);
      showNotification("Error saving settings", err.message || err, "critical");
    }
  };

  const cancel = () => {
    dispatch("cancel");
  };

  const openLogDir = async () => {
    try {
      await window.go.main.App.OpenLogDir();
    } catch (err) {
      console.error(err);
      showNotification("Error opening log directory", err.message || err, "critical");
    }
  };

  const namePipeToggle = () => {
    if (data.ProxyModeOfNamedPipe) {
      data.ProxyModeOfNamedPipe = false;
    }
  };
  const proxyToggle = () => {
    if (data.ProxyModeOfNamedPipe) {
      data.NamedPipeAgent = false;
    }
  };
</script>

<div class="settings-view">
  <div class="settings-header">
    <h2>{t[lang].settingsTitle}</h2>
  </div>

  {#if infoBarOpen}
    <div style="margin-bottom: 16px;">
      <InfoBar
        bind:open={infoBarOpen}
        severity={infoBarSeverity}
        title={infoBarTitle}
        message={infoBarMessage}
        closable={true}
      />
    </div>
  {/if}

  <div class="settings-content">
    <!-- General Settings Section -->
    <Expander expanded={true}>
      <div class="section-header-content">
        <span class="material-icons section-icon">tune</span>
        <span>{t[lang].sectionGeneral}</span>
      </div>

      <svelte:fragment slot="content">
        <div class="settings-list">
          <!-- Language -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">language</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsLanguage}</span>
              </div>
            </div>
            <span class="select-wrapper">
              <select
                id="lang-select"
                class="settings-select"
                bind:value={uiLang}
              >
                <option value="auto">{t[lang].settingsLanguageAuto}</option>
                <option value="ja">{t[lang].settingsLanguageJa}</option>
                <option value="en">{t[lang].settingsLanguageEn}</option>
              </select>
            </span>
          </div>

          <!-- Theme -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">dark_mode</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsTheme}</span>
              </div>
            </div>
            <span class="select-wrapper">
              <select
                id="theme-select"
                class="settings-select"
                bind:value={uiTheme}
              >
                <option value="auto">{t[lang].settingsThemeAuto}</option>
                <option value="dark">{t[lang].settingsThemeDark}</option>
                <option value="light">{t[lang].settingsThemeLight}</option>
              </select>
            </span>
          </div>

          <!-- Accent Color -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">palette</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsAccent}</span>
              </div>
            </div>
            <div class="accent-picker-container">
              <span
                class="accent-preview-dot"
                style="background-color: {effectiveAccentColor};"
              ></span>
              <span class="select-wrapper">
                <select
                  id="accent-select"
                  class="settings-select"
                  bind:value={uiAccent}
                >
                  <option value="default">{t[lang].settingsAccentDefault}</option>
                  <option value="blue">{t[lang].settingsAccentBlue}</option>
                  <option value="indigo">{t[lang].settingsAccentIndigo}</option>
                  <option value="teal">{t[lang].settingsAccentTeal}</option>
                  <option value="emerald">{t[lang].settingsAccentEmerald}</option>
                  <option value="sunset">{t[lang].settingsAccentSunset}</option>
                  <option value="rose">{t[lang].settingsAccentRose}</option>
                  <option value="custom">{t[lang].settingsAccentCustom}</option>
                </select>
              </span>
              {#if uiAccent === "custom"}
                <input
                  type="color"
                  class="accent-color-picker"
                  bind:value={customAccentColor}
                />
              {/if}
            </div>
          </div>

          <!-- Start at Login -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">login</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsStartAtLogin}</span>
              </div>
            </div>
            <ToggleSwitch bind:checked={data.StartAtLogin} />
          </div>

          <!-- Start Hidden -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">visibility_off</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsStartHidden}</span>
              </div>
            </div>
            <ToggleSwitch bind:checked={data.StartHidden} />
          </div>

          <!-- Balloon Notification -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">notifications</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsBalloon}</span>
              </div>
            </div>
            <ToggleSwitch bind:checked={data.ShowBalloon} />
          </div>

          <!-- Debug Log -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">bug_report</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsDebugLog}</span>
              </div>
            </div>
            <div class="settings-item-actions">
              <ToggleSwitch bind:checked={data.DebugLog} />
              <Button on:click={openLogDir}>
                <span class="material-icons" style="margin-right: 6px; font-size: 16px;">folder_open</span>
                {t[lang].settingsOpenLogDir}
              </Button>
            </div>
          </div>
        </div>
      </svelte:fragment>
    </Expander>

    <!-- SSH Agent Settings Section -->
    <Expander expanded={true}>
      <div class="section-header-content">
        <span class="material-icons section-icon">terminal</span>
        <span>{t[lang].sectionAgent}</span>
      </div>

      <svelte:fragment slot="content">
        <div class="settings-list">
          <!-- Pageant -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">vpn_key</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsPageant}</span>
              </div>
            </div>
            <ToggleSwitch bind:checked={data.PageantAgent} />
          </div>

          <!-- Named Pipe Agent -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">settings_ethernet</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsNamedPipe}</span>
              </div>
            </div>
            <ToggleSwitch
              bind:checked={data.NamedPipeAgent}
              on:change={namePipeToggle}
            />
          </div>

          <!-- Proxy Mode -->
          <div class="settings-item-row">
            <div class="settings-item-row-left">
              <span class="material-icons">swap_horiz</span>
              <div class="settings-item-text">
                <span class="settings-item-title">{t[lang].settingsProxy}</span>
              </div>
            </div>
            <ToggleSwitch
              bind:checked={data.ProxyModeOfNamedPipe}
              on:change={proxyToggle}
            />
          </div>

          <!-- Unix Domain Socket -->
          <div class="settings-item-row-group">
            <div class="settings-item-row">
              <div class="settings-item-row-left">
                <span class="material-icons">folder</span>
                <div class="settings-item-text">
                  <span class="settings-item-title">{t[lang].settingsUnix}</span>
                </div>
              </div>
              <ToggleSwitch bind:checked={data.UnixSocketAgent} />
            </div>
            {#if data.UnixSocketAgent}
              <div class="settings-field-row">
                <div class="settings-field-label">
                  <span class="field-title">{t[lang].settingsUnixPath}</span>
                  <span class="field-helper">{t[lang].settingsUnixHelper}</span>
                </div>
                <TextBox
                  bind:value={data.UnixSocketPath}
                  style="width: 100%;"
                />
              </div>
            {/if}
          </div>

          <!-- Cygwin Unix Domain Socket -->
          <div class="settings-item-row-group">
            <div class="settings-item-row">
              <div class="settings-item-row-left">
                <span class="material-icons">dns</span>
                <div class="settings-item-text">
                  <span class="settings-item-title">{t[lang].settingsCygwin}</span>
                </div>
              </div>
              <ToggleSwitch bind:checked={data.CygWinAgent} />
            </div>
            {#if data.CygWinAgent}
              <div class="settings-field-row">
                <div class="settings-field-label">
                  <span class="field-title">{t[lang].settingsCygwinPath}</span>
                  <span class="field-helper">{t[lang].settingsCygwinHelper}</span>
                </div>
                <TextBox
                  bind:value={data.CygWinSocketPath}
                  style="width: 100%;"
                />
              </div>
            {/if}
          </div>
        </div>
      </svelte:fragment>
    </Expander>
  </div>

  <div class="settings-actions">
    <Button variant="accent" on:click={save}>
      {t[lang].settingsSave}
    </Button>
    <Button on:click={cancel}>
      {t[lang].settingsCancel}
    </Button>
  </div>
</div>

<style>
  .settings-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
    padding: 16px;
    background-color: var(--fds-solid-background-base, #f3f3f3);
    color: var(--fds-text-primary, #1f1f1f);
  }
  .settings-header {
    margin-bottom: 16px;
  }
  .settings-header h2 {
    margin: 0;
    font-size: 20px;
    font-weight: 600;
  }
  .settings-content {
    flex: 1;
    overflow-y: auto;
    margin-bottom: 16px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding-right: 4px;
  }
  
  /* section header inside Expander */
  .section-header-content {
    display: flex;
    align-items: center;
    gap: 12px;
    font-size: 15px;
    font-weight: 600;
  }
  .section-icon {
    font-size: 18px;
    color: var(--fds-text-secondary, #5f5f5f);
  }

  .settings-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  /* Fluent-like Row Styling */
  .settings-item-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 16px;
    background-color: var(--fds-card-background-default, #ffffff);
    border: 1px solid var(--fds-card-stroke-default, #e5e5e5);
    border-radius: 4px;
    transition: background-color 0.15s ease;
  }
  .settings-item-row:hover {
    background-color: var(--fds-subtle-fill-secondary, #f0f0f0);
  }
  .settings-item-row-left {
    display: flex;
    align-items: center;
    gap: 16px;
  }
  .settings-item-row-left .material-icons {
    font-size: 20px;
    color: var(--fds-text-secondary, #5f5f5f);
  }
  .settings-item-text {
    display: flex;
    flex-direction: column;
  }
  .settings-item-title {
    font-size: 14px;
    font-weight: 500;
  }
  .settings-item-actions {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .settings-item-row-group {
    background-color: var(--fds-card-background-default, #ffffff);
    border: 1px solid var(--fds-card-stroke-default, #e5e5e5);
    border-radius: 4px;
    display: flex;
    flex-direction: column;
  }
  .settings-item-row-group .settings-item-row {
    border: none;
    border-radius: 4px 4px 0 0;
    background-color: transparent;
  }
  .settings-item-row-group .settings-item-row:hover {
    background-color: var(--fds-subtle-fill-secondary, #f0f0f0);
  }

  /* TextBox and Field customization */
  .settings-field-row {
    padding: 12px 16px 16px 52px;
    border-top: 1px solid var(--fds-divider-stroke-default, #e5e5e5);
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .settings-field-label {
    display: flex;
    flex-direction: column;
  }
  .field-title {
    font-size: 13px;
    font-weight: 500;
  }
  .field-helper {
    font-size: 11px;
    color: var(--fds-text-secondary, #5f5f5f);
  }

  /* Accent picker */
  .accent-picker-container {
    display: flex;
    align-items: center;
    gap: 12px;
    flex: 1;
    min-width: 170px;
    max-width: 320px;
  }
  .accent-color-picker {
    width: 30px;
    height: 30px;
    padding: 0;
    border: 1px solid var(--fds-control-stroke-default, #e5e5e5);
    border-radius: 4px;
    background: none;
    cursor: pointer;
    box-sizing: border-box;
  }
  .accent-preview-dot {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    border: 1px solid var(--fds-control-stroke-default, #e5e5e5);
    display: inline-block;
    box-shadow: var(--fds-card-shadow);
  }

  /* Select wrapper for theme-aware caret */
  .select-wrapper {
    position: relative;
    display: flex;
    flex: 1;
    min-width: 170px;
    max-width: 320px;
  }
  .select-wrapper::after {
    content: "";
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    width: 0;
    height: 0;
    border-left: 5px solid transparent;
    border-right: 5px solid transparent;
    border-top: 6px solid var(--text-color);
    pointer-events: none;
  }

  /* Standard Select styling */
  .settings-select {
    width: 100%;
    padding: 6px 32px 6px 12px;
    border: 1px solid var(--border-color);
    border-radius: 4px;
    background-color: var(--surface-color);
    color: var(--text-color);
    font-size: 13px;
    font-family: inherit;
    outline: none;
    cursor: pointer;
    appearance: none;
    transition: background-color 0.15s, border-color 0.15s;
  }
  .settings-select:hover {
    background-color: var(--hover-bg);
    border-color: var(--text-secondary);
  }
  .settings-select:focus {
    border-color: var(--primary-color);
  }
  .settings-select option {
    background-color: var(--surface-color);
    color: var(--text-color);
  }

  /* Expander custom variables override to match card design */
  :global(.fds-expander) {
    border: 1px solid var(--fds-card-stroke-default, #e5e5e5) !important;
    background-color: var(--fds-card-background-default, #ffffff) !important;
    border-radius: 4px !important;
    overflow: hidden;
  }
  :global(.fds-expander-content) {
    background-color: var(--fds-card-background-secondary, #fafafa) !important;
    padding: 8px !important;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .settings-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding-top: 16px;
    border-top: 1px solid var(--fds-divider-stroke-default, #e5e5e5);
  }
  .settings-actions :global(button) {
    min-width: 120px;
  }
</style>
