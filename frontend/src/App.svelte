<script>
  import { onMount } from "svelte";
  import { ListItem, Button, TextBox, ToggleSwitch, ContentDialog, InfoBar } from "fluent-svelte";
  import "fluent-svelte/theme.css";
  import Settings from "./Settings.svelte";
  import AddFileDialog from "./AddFileDialog.svelte";

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

  const STORAGE_KEY_SIDEBAR_WIDTH = "omni-sidebar-width";
  let sidebarWidth = 270;
  let isResizing = false;

  const startResize = (event) => {
    isResizing = true;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    window.addEventListener("mousemove", resize);
    window.addEventListener("mouseup", stopResize);
  };

  const resize = (event) => {
    if (!isResizing) return;
    const newWidth = event.clientX;
    if (newWidth >= 180 && newWidth <= 500) {
      sidebarWidth = newWidth;
    }
  };

  const stopResize = () => {
    isResizing = false;
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    window.removeEventListener("mousemove", resize);
    window.removeEventListener("mouseup", stopResize);
    localStorage.setItem(STORAGE_KEY_SIDEBAR_WIDTH, sidebarWidth.toString());
  };

  const STORAGE_KEY_LANG = "omni-settings-lang";
  const STORAGE_KEY_THEME = "omni-settings-theme";
  const STORAGE_KEY_ACCENT = "omni-settings-accent";

  const accentColorsMap = {
    blue: { light: "#0078d4", dark: "#60cdff" },
    indigo: { light: "#4f46e5", dark: "#a78bfa" },
    teal: { light: "#0f766e", dark: "#2dd4bf" },
    emerald: { light: "#047857", dark: "#34d399" },
    sunset: { light: "#c2410c", dark: "#fb923c" },
    rose: { light: "#be123c", dark: "#fb7185" },
  };

  const normalizeHexColor = (color) => {
    const match = /^#?([0-9a-f]{3}|[0-9a-f]{6})$/i.exec((color || "").trim());
    if (!match) return null;
    const hex =
      match[1].length === 3
        ? match[1]
            .split("")
            .map((part) => part + part)
            .join("")
        : match[1];
    return `#${hex.toLowerCase()}`;
  };

  const getRelativeLuminance = (hexColor) => {
    const hex = hexColor.slice(1);
    const channels = [0, 2, 4].map((start) =>
      parseInt(hex.slice(start, start + 2), 16),
    );
    const [r, g, b] = channels.map((channel) => {
      const value = channel / 255;
      return value <= 0.03928
        ? value / 12.92
        : Math.pow((value + 0.055) / 1.055, 2.4);
    });
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  };

  const getReadableTextColor = (backgroundColor) => {
    const luminance = getRelativeLuminance(backgroundColor);
    const whiteContrast = 1.05 / (luminance + 0.05);
    const blackContrast = (luminance + 0.05) / 0.05;
    return whiteContrast >= blackContrast ? "#ffffff" : "#000000";
  };

  const storedLang = localStorage.getItem(STORAGE_KEY_LANG);
  let lang =
    storedLang && storedLang !== "auto"
      ? storedLang
      : navigator.language.startsWith("ja")
        ? "ja"
        : "en";

  const t = {
    en: {
      appTitle: "Omni SSH Agent",
      keysTitle: "KEYS",
      emptyKeys: "No keys loaded",
      unnamedKey: "Unnamed Key",
      unknownType: "Unknown",
      addPrivateKey: "Add Private Key",
      settings: "Settings",
      welcomeTitle: "Welcome to Omni SSH Agent",
      welcomeDesc:
        "Select a key from the sidebar to view details, or add a new key to get started.",
      filePath: "File Path of the Private Key",
      keyType: "SSH Key Type",
      fingerprintSha256: "Fingerprint SHA256",
      fingerprintMd5: "Fingerprint MD5",
      publicKey: "SSH Public Key",
      deleteKey: "Delete Key",
      copy: "Copy",
      copyKey: "Copy Key",
      copied: "Copied to clipboard!",
      failCopy: "Failed to copy",
      confirmDelete: "Do you really want to delete this key?",
      deletedSuccess: "Successfully deleted key",
      addedSuccess: "Successful add key",
      enableKey: "Key enabled",
      disableKey: "Key disabled",
      toggleFail: "Failed to toggle key",
      keyEnabledLabel: "Key is enabled",
      keyDisabledLabel: "Key is disabled",
    },
    ja: {
      appTitle: "Omni SSH Agent",
      keysTitle: "鍵一覧",
      emptyKeys: "読み込まれた鍵はありません",
      unnamedKey: "無名の鍵",
      unknownType: "不明",
      addPrivateKey: "秘密鍵の追加",
      settings: "設定",
      welcomeTitle: "Omni SSH Agent へようこそ",
      welcomeDesc:
        "サイドバーから鍵を選択して詳細を表示するか、新しい鍵を追加してください。",
      filePath: "秘密鍵のファイルパス",
      keyType: "SSH鍵の種類",
      fingerprintSha256: "フィンガープリント SHA256",
      fingerprintMd5: "フィンガープリント MD5",
      publicKey: "SSH公開鍵",
      deleteKey: "鍵を削除",
      copy: "コピー",
      copyKey: "鍵をコピー",
      copied: "クリップボードにコピーしました！",
      failCopy: "コピーに失敗しました",
      confirmDelete: "本当にこの鍵を削除しますか？",
      deletedSuccess: "鍵を正常に削除しました",
      addedSuccess: "鍵を追加しました",
      enableKey: "鍵を有効にしました",
      disableKey: "鍵を無効にしました",
      toggleFail: "鍵の切り替えに失敗しました",
      keyEnabledLabel: "鍵は有効です",
      keyDisabledLabel: "鍵は無効です",
    },
  };

  let keys = [];
  let selectedKey = null;
  let activeView = "welcome"; // 'welcome', 'detail', 'settings'
  let addFileDialog;

  let togglingKey = null;

  let deleteDialogOpen = false;
  let deleteTargetSha256 = "";

  let dragOver = false;
  let dragCounter = 0;

  let settingsData = { ProxyModeOfNamedPipe: false };
  let appVersion = "";

  const resetDragState = () => {
    dragCounter = 0;
    dragOver = false;
  };

  const isFileDrag = (e) => {
    return (
      e.dataTransfer && Array.from(e.dataTransfer.types || []).includes("Files")
    );
  };

  const addLocalFilePath = async (filePath) => {
    if (!filePath) return;
    try {
      const privateKeyFile = await window.go.main.App.CheckKeyType(
        filePath,
        "",
      );
      if (privateKeyFile.encryption) {
        addFileDialog.show(filePath);
        return;
      }
      await addlocalFile(privateKeyFile);
    } catch (err) {
      console.error("drop addkey err:" + err);
      showNotification("Error dropping file", err.message || err, "critical");
    }
  };

  const addDroppedFiles = async (paths) => {
    if (settingsData.ProxyModeOfNamedPipe || !paths) return;
    for (const path of paths) {
      await addLocalFilePath(path);
    }
  };

  onMount(async () => {
    const storedWidth = localStorage.getItem(STORAGE_KEY_SIDEBAR_WIDTH);
    if (storedWidth) {
      const parsed = parseInt(storedWidth, 10);
      if (!isNaN(parsed) && parsed >= 180 && parsed <= 500) {
        sidebarWidth = parsed;
      }
    }
    await loadVersion();
    await loadSettings();
    await loadKeys();
    await applyWindowsTheme();

    window.runtime.OnFileDrop((x, y, paths) => {
      resetDragState();
      addDroppedFiles(paths);
    }, true);
  });

  const updateSmuiThemeStylesheet = (isLight) => {
    const lightTheme =
      document.getElementById("smui-theme-light") ||
      document.querySelector('link[href="/smui.css"]');
    const darkTheme =
      document.getElementById("smui-theme-dark") ||
      document.querySelector('link[href="/smui-dark.css"]');
    if (!lightTheme || !darkTheme) return;

    lightTheme.media = isLight ? "all" : "not all";
    darkTheme.media = isLight ? "not all" : "all";
  };

  const applyThemeClass = (isLight) => {
    const root = document.documentElement;
    root.classList.toggle("light-theme", isLight);
    root.classList.toggle("dark-theme", !isLight);
    root.classList.toggle("dark", !isLight);
    updateSmuiThemeStylesheet(isLight);
  };

  const resolveThemePreference = async () => {
    const storedTheme = localStorage.getItem(STORAGE_KEY_THEME);
    if (storedTheme === "light") {
      return true;
    } else if (storedTheme === "dark") {
      return false;
    }

    try {
      return await window.go.main.App.GetAppsUseLightTheme();
    } catch (e) {
      console.error("Failed to load Windows theme preference:", e);
      return !window.matchMedia("(prefers-color-scheme: dark)").matches;
    }
  };

  const resolveAccentColor = async (isLight) => {
    const storedAccent = localStorage.getItem(STORAGE_KEY_ACCENT);
    const preset = accentColorsMap[storedAccent];
    if (preset) {
      return preset[isLight ? "light" : "dark"];
    }
    if (storedAccent && storedAccent !== "default") {
      return storedAccent;
    }

    try {
      return await window.go.main.App.GetAccentColor();
    } catch (e) {
      console.error("Failed to load Windows accent color:", e);
      return null;
    }
  };

  const applyAccentColor = (color) => {
    const root = document.documentElement;
    root.style.removeProperty("--primary-color");
    root.style.removeProperty("--primary-on-color");
    root.style.removeProperty("--mdc-theme-primary");
    root.style.removeProperty("--mdc-theme-on-primary");
    root.style.removeProperty("--fds-accent-default");

    const normalizedColor = normalizeHexColor(color);
    if (!normalizedColor) return;

    root.style.setProperty("--primary-color", normalizedColor);
    root.style.setProperty(
      "--primary-on-color",
      getReadableTextColor(normalizedColor),
    );
    root.style.setProperty("--fds-accent-default", normalizedColor);
  };

  const applyWindowsTheme = async () => {
    const isLight = await resolveThemePreference();
    applyThemeClass(isLight);

    const color = await resolveAccentColor(isLight);
    applyAccentColor(color);
  };

  const syncAppState = async () => {
    await loadKeys();
    await applyWindowsTheme();
  };

  const loadVersion = async () => {
    try {
      appVersion = await window.go.main.App.GetVersion();
    } catch (err) {
      console.error("GetVersion err:" + err);
    }
  };

  const loadSettings = async () => {
    await window.go.main.App.GetSettings()
      .then((savedata) => {
        settingsData = { ...savedata };
      })
      .catch((err) => {
        console.error(err);
        showNotification("Error loading settings", err.message || err, "critical");
      });
  };

  const loadKeys = async () => {
    await window.go.main.App.KeyList()
      .then((list) => {
        keys = list;
      })
      .catch((err) => {
        console.error("KeyList err:" + err);
        if (!settingsData.ProxyModeOfNamedPipe) {
          showNotification("Error listing keys", err.message || err, "critical");
        }
      });
  };

  const selectKey = (key) => {
    selectedKey = key;
    activeView = "detail";
  };

  const openSettings = () => {
    selectedKey = null;
    activeView = "settings";
  };

  const handleSettingsSave = async () => {
    await loadSettings();
    await loadKeys();
    const storedLang = localStorage.getItem(STORAGE_KEY_LANG);
    lang =
      storedLang && storedLang !== "auto"
        ? storedLang
        : navigator.language.startsWith("ja")
          ? "ja"
          : "en";
    await applyWindowsTheme();
    activeView = "welcome";
  };

  const handleSettingsCancel = () => {
    activeView = "welcome";
  };

  const handleData = (event) => {
    addlocalFile(event.detail);
  };

  const addlocalFile = async (privateKeyFile) => {
    await window.go.main.App.AddLocalFile(privateKeyFile)
      .then(() => {
        showNotification(t[lang].addedSuccess, "", "success");
        loadKeys();
      })
      .catch((err) => {
        console.error("addkey err:" + err);
        showNotification("Error adding key", err.message || err, "critical");
      });
  };

  const delKey = (sha256) => {
    deleteTargetSha256 = sha256;
    deleteDialogOpen = true;
  };

  const confirmDelKey = async () => {
    deleteDialogOpen = false;
    const sha256 = deleteTargetSha256;
    deleteTargetSha256 = "";
    await window.go.main.App.DeleteKey(sha256)
      .then(() => {
        showNotification(t[lang].deletedSuccess, "", "success");
        loadKeys();
        if (selectedKey && selectedKey.publickey.sha256 === sha256) {
          selectedKey = null;
          activeView = "welcome";
        }
      })
      .catch((err) => {
        if (err == "cancel") return;
        showNotification("Error deleting key", err.message || err, "critical");
      });
  };

  const toggleKey = async (key) => {
    togglingKey = key.publickey.sha256;
    try {
      await window.go.main.App.ToggleKey(key.publickey.sha256);
      const idx = keys.findIndex(
        (k) => k.publickey.sha256 === key.publickey.sha256,
      );
      if (idx !== -1) {
        keys[idx].disabled = !keys[idx].disabled;
        showNotification(
          keys[idx].disabled ? t[lang].disableKey : t[lang].enableKey,
          "",
          "success"
        );
        keys = keys;
        if (
          selectedKey &&
          selectedKey.publickey.sha256 === key.publickey.sha256
        ) {
          selectedKey.disabled = keys[idx].disabled;
          selectedKey = selectedKey;
        }
      }
    } catch (err) {
      showNotification("Error toggling key", err.message || err, "critical");
    } finally {
      togglingKey = null;
    }
  };

  const copyText = (text) => {
    navigator.clipboard
      .writeText(text)
      .then(() => {
        showNotification(t[lang].copied, "", "success");
      })
      .catch((err) => {
        showNotification("Error copying text", err.message || err, "critical");
      });
  };

  const onLoadKeysEvent = (message) => {
    if (togglingKey) return;
    loadKeys();
    console.log(message);
  };

  window.runtime.EventsOn("LoadKeysEvent", onLoadKeysEvent);

  const handleKeydown = (e) => {
    if (
      e.key === "Delete" &&
      selectedKey &&
      activeView === "detail" &&
      !settingsData.ProxyModeOfNamedPipe
    ) {
      delKey(selectedKey.publickey.sha256);
    }
  };

  const handleDragOver = (e) => {
    if (!isFileDrag(e)) return;
    e.preventDefault();
    if (!settingsData.ProxyModeOfNamedPipe) {
      dragOver = true;
    }
  };

  const handleDragEnter = (e) => {
    if (!isFileDrag(e)) return;
    e.preventDefault();
    dragCounter++;
    if (!settingsData.ProxyModeOfNamedPipe) {
      dragOver = true;
    }
  };

  const handleDragLeave = (e) => {
    if (!isFileDrag(e)) return;
    e.preventDefault();
    dragCounter--;
    if (dragCounter <= 0) {
      dragCounter = 0;
      dragOver = false;
    }
  };

  const handleDrop = (e) => {
    if (!isFileDrag(e)) return;
    e.preventDefault();
    resetDragState();
  };
</script>

<main
  class="app-container"
  class:drag-over={dragOver}
  on:dragenter={handleDragEnter}
  on:dragover={handleDragOver}
  on:dragleave={handleDragLeave}
  on:drop={handleDrop}
>
  <!-- Left Panel: Sidebar -->
  <aside
    class="sidebar"
    style="width: {sidebarWidth}px; min-width: {sidebarWidth}px; max-width: {sidebarWidth}px;"
    data-wails-no-drag
  >
    <!-- Keys list -->
    <div class="sidebar-content">
      {#if keys.length === 0}
        <div class="empty-keys-message">{t[lang].emptyKeys}</div>
      {:else}
        <ul class="keys-list">
          {#each keys as key}
            <ListItem
              class="key-list-item"
              selected={selectedKey && selectedKey.publickey.sha256 === key.publickey.sha256}
              disabled={key.disabled}
              on:click={() => selectKey(key)}
              on:dblclick={() => {
                if (!settingsData.ProxyModeOfNamedPipe) toggleKey(key);
              }}
            >
              <span slot="icon" class="material-icons key-icon">key</span>
              <span class="key-info">
                <span class="key-name">{key.name || t[lang].unnamedKey}</span>
                <span class="key-type"
                  >{key.publickey.type || t[lang].unknownType}</span
                >
              </span>
            </ListItem>
          {/each}
        </ul>
      {/if}
    </div>

    <!-- Sidebar footer action buttons -->
    <div class="sidebar-footer">
      {#if !settingsData.ProxyModeOfNamedPipe}
        <Button
          class="action-btn add-btn"
          variant="accent"
          on:click={() => addFileDialog.show()}
        >
          <span class="material-icons" style="margin-right: 6px;">add</span>
          {t[lang].addPrivateKey}
        </Button>
      {/if}
      <Button
        class="action-btn settings-btn"
        on:click={openSettings}
      >
        <span class="material-icons" style="margin-right: 6px;">settings</span>
        {t[lang].settings}
      </Button>
    </div>
  </aside>

  <!-- Resizer Handle -->
  <div
    class="sidebar-resizer {isResizing ? 'resizing' : ''}"
    on:mousedown={startResize}
  ></div>

  <!-- Right Panel: Main Content Area -->
  <section class="main-content" data-wails-no-drag>
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

    {#if activeView === "welcome"}
      <div class="welcome-view">
        <span class="material-icons welcome-icon">security</span>
        <h2>{t[lang].welcomeTitle}</h2>
        {#if appVersion}
          <div class="welcome-version">v{appVersion}</div>
        {/if}
      </div>
    {:else if activeView === "detail" && selectedKey}
      <div class="detail-view">
        <div class="detail-header" style="display: flex; align-items: center; gap: 16px; margin-bottom: 24px;">
          {#if !settingsData.ProxyModeOfNamedPipe}
            <ToggleSwitch
              checked={!selectedKey.disabled}
              on:change={() => toggleKey(selectedKey)}
            />
          {/if}
          <h2>{selectedKey.name || t[lang].unnamedKey}</h2>
        </div>

        <div class="detail-fields" style="max-width: 800px; display: flex; flex-direction: column; gap: 16px;">
          <!-- Basic Info Card -->
          <div class="detail-card">
            <!-- File Path Row -->
            <div class="detail-row">
              <div class="detail-row-info" style="flex: 1;">
                <span class="detail-row-title">{t[lang].filePath}</span>
                <span class="detail-row-value">{selectedKey.filePath}</span>
              </div>
              <Button on:click={() => copyText(selectedKey.filePath)}>
                <span class="material-icons" style="font-size: 16px; margin-right: 4px;">content_copy</span>
                {t[lang].copy}
              </Button>
            </div>

            <div class="detail-row-divider"></div>

            <!-- Key Type Row -->
            <div class="detail-row">
              <div class="detail-row-info">
                <span class="detail-row-title">{t[lang].keyType}</span>
                <span class="detail-row-value">{selectedKey.publickey.type}</span>
              </div>
            </div>
          </div>

          <!-- Fingerprints Card -->
          <div class="detail-card-title">Fingerprints</div>
          <div class="detail-card">
            <!-- SHA256 -->
            <div class="detail-row">
              <div class="detail-row-info" style="flex: 1;">
                <span class="detail-row-title">{t[lang].fingerprintSha256}</span>
                <span class="detail-row-value monospace">{selectedKey.publickey.sha256}</span>
              </div>
              <Button on:click={() => copyText(selectedKey.publickey.sha256)}>
                <span class="material-icons" style="font-size: 16px; margin-right: 4px;">content_copy</span>
                {t[lang].copy}
              </Button>
            </div>

            <div class="detail-row-divider"></div>

            <!-- MD5 -->
            <div class="detail-row">
              <div class="detail-row-info" style="flex: 1;">
                <span class="detail-row-title">{t[lang].fingerprintMd5}</span>
                <span class="detail-row-value monospace">{selectedKey.publickey.md5}</span>
              </div>
              <Button on:click={() => copyText(selectedKey.publickey.md5)}>
                <span class="material-icons" style="font-size: 16px; margin-right: 4px;">content_copy</span>
                {t[lang].copy}
              </Button>
            </div>
          </div>

          <!-- Public Key Card -->
          <div class="detail-card-title">{t[lang].publicKey}</div>
          <div class="detail-card" style="padding: 16px; gap: 12px;">
            <div class="publickey-box" style="margin: 0; height: 120px;">
              {selectedKey.publickey.string}
            </div>
            <div style="display: flex; justify-content: flex-end;">
              <Button variant="accent" on:click={() => copyText(selectedKey.publickey.string)}>
                <span class="material-icons" style="font-size: 16px; margin-right: 6px;">content_copy</span>
                {t[lang].copyKey}
              </Button>
            </div>
          </div>
        </div>

        {#if !settingsData.ProxyModeOfNamedPipe}
          <div class="detail-actions">
            <Button
              class="delete-btn"
              on:click={() => delKey(selectedKey.publickey.sha256)}
            >
              <span class="material-icons" style="margin-right: 6px;">delete</span>
              {t[lang].deleteKey}
            </Button>
          </div>
        {/if}
      </div>
    {:else if activeView === "settings"}
      <Settings
        {lang}
        on:save={handleSettingsSave}
        on:cancel={handleSettingsCancel}
      />
    {/if}
  </section>

  <!-- Dialogs & Toasts -->
  <AddFileDialog
    {lang}
    bind:this={addFileDialog}
    on:eventAddPkfile={handleData}
  />

  <ContentDialog
    bind:open={deleteDialogOpen}
    title={t[lang].confirmDelete}
  >
    <svelte:fragment slot="footer">
      <Button
        on:click={() => {
          deleteDialogOpen = false;
          deleteTargetSha256 = "";
        }}
      >
        No
      </Button>
      <Button variant="accent" on:click={confirmDelKey}>
        Yes
      </Button>
    </svelte:fragment>
  </ContentDialog>
</main>

<!-- Mouse triggers for syncing keyrings when window becomes active -->
<svelte:body on:mouseenter={syncAppState} on:mouseleave={syncAppState} />
<svelte:window on:keydown={handleKeydown} />
