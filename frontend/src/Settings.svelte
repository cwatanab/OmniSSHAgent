<script>
  import { createEventDispatcher } from "svelte";
  import Button, { Label } from "@smui/button";
  import Textfield from "@smui/textfield";
  import HelperText from "@smui/textfield/helper-text";
  import Card from "@smui/card";
  import FormField from "@smui/form-field";
  import Checkbox from "@smui/checkbox";
  import { toast } from "@zerodevx/svelte-toast";
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

  const red = {
    duration: 7000,
    theme: {
      "--toastBackground": "#F56565",
      "--toastBarBackground": "#C53030",
    },
  };
  const green = {
    theme: {
      "--toastBackground": "#48BB78",
      "--toastBarBackground": "#2F855A",
    },
  };

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
      toast.push(err, red);
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

      toast.push(t[lang].settingsSaved, green);
      dispatch("save");
    } catch (err) {
      console.error(err);
      toast.push(err, red);
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
      toast.push(err, red);
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

  <div class="settings-content">
    <div class="settings-section">
      <h3 class="section-title">
        <span class="material-icons section-icon">tune</span>
        {t[lang].sectionGeneral}
      </h3>
      <Card padded>
        <div class="settings-list">
          <div class="settings-inline-row">
            <div class="settings-item settings-select-item">
              <span class="material-icons setting-icon">language</span>
              <label class="settings-select-label" for="lang-select"
                >{t[lang].settingsLanguage}</label
              >
              <select
                id="lang-select"
                class="settings-select"
                bind:value={uiLang}
              >
                <option value="auto">{t[lang].settingsLanguageAuto}</option>
                <option value="ja">{t[lang].settingsLanguageJa}</option>
                <option value="en">{t[lang].settingsLanguageEn}</option>
              </select>
            </div>

            <div class="settings-item settings-select-item">
              <span class="material-icons setting-icon">dark_mode</span>
              <label class="settings-select-label" for="theme-select"
                >{t[lang].settingsTheme}</label
              >
              <select
                id="theme-select"
                class="settings-select"
                bind:value={uiTheme}
              >
                <option value="auto">{t[lang].settingsThemeAuto}</option>
                <option value="dark">{t[lang].settingsThemeDark}</option>
                <option value="light">{t[lang].settingsThemeLight}</option>
              </select>
            </div>
          </div>

          <div class="settings-item settings-select-item">
            <span class="material-icons setting-icon">palette</span>
            <label class="settings-select-label" for="accent-select"
              >{t[lang].settingsAccent}</label
            >
            <div class="accent-picker-container">
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
              {#if uiAccent === "custom"}
                <input
                  type="color"
                  class="accent-color-picker"
                  bind:value={customAccentColor}
                />
              {/if}
              <span
                class="accent-preview-dot"
                style="background-color: {effectiveAccentColor};"
              ></span>
            </div>
          </div>

          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.StartAtLogin} />
              <span>{t[lang].settingsStartAtLogin}</span>
            </FormField>
          </div>

          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.StartHidden} />
              <span>{t[lang].settingsStartHidden}</span>
            </FormField>
          </div>

          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.ShowBalloon} />
              <span>{t[lang].settingsBalloon}</span>
            </FormField>
          </div>

          <div class="settings-log-row">
            <div class="settings-item">
              <FormField>
                <Checkbox bind:checked={data.DebugLog} />
                <span>{t[lang].settingsDebugLog}</span>
              </FormField>
            </div>

            <div class="settings-item settings-log-item">
              <Button variant="outlined" on:click={openLogDir}>
                <span class="material-icons" style="margin-right: 6px;"
                  >folder_open</span
                >
                <Label>{t[lang].settingsOpenLogDir}</Label>
              </Button>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <div class="settings-section">
      <h3 class="section-title">
        <span class="material-icons section-icon">terminal</span>
        {t[lang].sectionAgent}
      </h3>
      <Card padded>
        <div class="settings-list">
          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.PageantAgent} />
              <span>{t[lang].settingsPageant}</span>
            </FormField>
          </div>
          <div class="settings-item">
            <FormField>
              <Checkbox
                bind:checked={data.NamedPipeAgent}
                on:change={namePipeToggle}
              />
              <span>{t[lang].settingsNamedPipe}</span>
            </FormField>
          </div>
          <div class="settings-item">
            <FormField>
              <Checkbox
                bind:checked={data.ProxyModeOfNamedPipe}
                on:change={proxyToggle}
              />
              <span>{t[lang].settingsProxy}</span>
            </FormField>
          </div>
          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.UnixSocketAgent} />
              <span>{t[lang].settingsUnix}</span>
            </FormField>
          </div>
          {#if data.UnixSocketAgent}
            <div class="settings-field">
              <Textfield
                bind:value={data.UnixSocketPath}
                label={t[lang].settingsUnixPath}
                style="width: 100%;"
                helperLine$style="width: 100%;"
              >
                <HelperText slot="helper"
                  >{t[lang].settingsUnixHelper}</HelperText
                >
              </Textfield>
            </div>
          {/if}
          <div class="settings-item">
            <FormField>
              <Checkbox bind:checked={data.CygWinAgent} />
              <span>{t[lang].settingsCygwin}</span>
            </FormField>
          </div>
          {#if data.CygWinAgent}
            <div class="settings-field">
              <Textfield
                bind:value={data.CygWinSocketPath}
                label={t[lang].settingsCygwinPath}
                style="width: 100%;"
                helperLine$style="width: 100%;"
              >
                <HelperText slot="helper"
                  >{t[lang].settingsCygwinHelper}</HelperText
                >
              </Textfield>
            </div>
          {/if}
        </div>
      </Card>
    </div>
  </div>

  <div class="settings-actions">
    <Button variant="raised" on:click={save}>
      <Label>{t[lang].settingsSave}</Label>
    </Button>
    <Button on:click={cancel}>
      <Label>{t[lang].settingsCancel}</Label>
    </Button>
  </div>
</div>

<style>
  .settings-view {
    display: flex;
    flex-direction: column;
    height: 100%;
    box-sizing: border-box;
  }
  .settings-header {
    margin-bottom: 8px;
  }
  .settings-header h2 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }
  .settings-content {
    flex: 1;
    overflow-y: auto;
    margin-bottom: 8px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .settings-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .settings-section {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }
  .section-title {
    margin: 4px 0 2px 4px;
    font-size: 16px;
    font-weight: 700;
    color: var(--text-secondary, #666);
    text-transform: uppercase;
    letter-spacing: 0.8px;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .section-icon {
    font-size: 20px;
    color: var(--text-secondary, #666);
    text-transform: none !important;
  }
  .setting-icon {
    font-size: 18px;
    color: var(--text-secondary, #666);
    display: inline-flex;
    align-items: center;
    justify-content: center;
  }
  .checkbox-icon {
    margin-right: 6px;
  }
  .settings-item {
    display: flex;
    align-items: center;
  }
  .settings-log-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    gap: 12px;
    align-items: center;
    margin-bottom: 8px;
  }
  .settings-log-item {
    justify-content: flex-end;
    min-width: max-content;
  }
  .settings-item :global(.mdc-form-field) {
    height: 32px;
  }
  .settings-content :global(.mdc-card) {
    padding: 12px 10px !important;
  }
  .accent-picker-container {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  .accent-color-picker {
    width: 28px;
    height: 28px;
    padding: 0;
    border: 1px solid var(--border-color, #e0e0e0);
    border-radius: 4px;
    background: none;
    cursor: pointer;
    box-sizing: border-box;
  }
  .accent-preview-dot {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    border: 1px solid var(--border-color, #e0e0e0);
    display: inline-block;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }
  .settings-field {
    padding-left: 52px;
    margin-top: -4px;
  }
  .settings-inline-row {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
    align-items: center;
  }
  .settings-select-item {
    gap: 12px;
    min-width: 0;
  }
  .settings-select-label {
    font-size: 14px;
    min-width: 60px;
  }
  .settings-select {
    padding: 8px 12px;
    border: 1px solid var(--mdc-text-field-outlined-idle-border-color, #ccc);
    border-radius: 4px;
    background-color: var(--bg-color, #fff);
    color: var(--text-color, #000);
    font-size: 14px;
    font-family: inherit;
    outline: none;
    cursor: pointer;
  }
  .settings-inline-row .settings-select {
    width: 100%;
    min-width: 96px;
  }
  .settings-select:hover {
    border-color: var(--primary-color, #0078d4);
  }
  .settings-select:focus {
    border-color: var(--primary-color, #0078d4);
    border-width: 2px;
  }
  .settings-actions {
    display: flex;
    justify-content: flex-end;
    gap: 8px;
    padding-top: 10px;
    border-top: 1px solid var(--border-color, #e0e0e0);
  }
</style>
