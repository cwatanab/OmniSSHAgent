<script>
  import { onMount } from "svelte";
  import { SvelteToast, toast } from "@zerodevx/svelte-toast";
  import Button, { Label } from "@smui/button";
  import Paper, { Content } from "@smui/paper";
  import Textfield from "@smui/textfield";
  import FormField from "@smui/form-field";
  import Switch from "@smui/switch";
  import Settings from "./Settings.svelte";
  import AddFileDialog from "./AddFileDialog.svelte";

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

  const STORAGE_KEY_SIDEBAR_WIDTH = "omni-sidebar-width";
  let sidebarWidth = 270;
  let isResizing = false;

  const startResize = (event) => {
    isResizing = true;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';
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
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    window.removeEventListener("mousemove", resize);
    window.removeEventListener("mouseup", stopResize);
    localStorage.setItem(STORAGE_KEY_SIDEBAR_WIDTH, sidebarWidth.toString());
  };

  const STORAGE_KEY_LANG = "omni-settings-lang";
  const STORAGE_KEY_THEME = "omni-settings-theme";

  const storedLang = localStorage.getItem(STORAGE_KEY_LANG);
  let lang = storedLang && storedLang !== "auto"
    ? storedLang
    : (navigator.language.startsWith("ja") ? "ja" : "en");

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
      welcomeDesc: "Select a key from the sidebar to view details, or add a new key to get started.",
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
      keyDisabledLabel: "Key is disabled"
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
      welcomeDesc: "サイドバーから鍵を選択して詳細を表示するか、新しい鍵を追加してください。",
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
      keyDisabledLabel: "鍵は無効です"
    }
  };

  let keys = [];
  let selectedKey = null;
  let activeView = "welcome"; // 'welcome', 'detail', 'settings'
  let addFileDialog;

  let togglingKey = null;

  let settingsData = { ProxyModeOfNamedPipe: false };

  onMount(async () => {
    const storedWidth = localStorage.getItem(STORAGE_KEY_SIDEBAR_WIDTH);
    if (storedWidth) {
      const parsed = parseInt(storedWidth, 10);
      if (!isNaN(parsed) && parsed >= 180 && parsed <= 500) {
        sidebarWidth = parsed;
      }
    }
    await loadSettings();
    await loadKeys();
    await applyWindowsTheme();
  });

  const applyWindowsTheme = async () => {
    const accentColorsMap = {
      blue: "#0078d4",
      indigo: "#6366f1",
      teal: "#0d9488",
      emerald: "#10b981",
      sunset: "#f97316",
      rose: "#f43f5e",
    };

    let color = null;
    const storedAccent = localStorage.getItem("omni-settings-accent");
    if (storedAccent && storedAccent !== "default") {
      color = accentColorsMap[storedAccent] || storedAccent;
    } else {
      try {
        color = await window.go.main.App.GetAccentColor();
      } catch (e) {
        console.error("Failed to load Windows accent color:", e);
      }
    }

    if (color) {
      const root = document.documentElement;
      root.style.setProperty("--primary-color", color);
      root.style.setProperty("--mdc-theme-primary", color);
      root.style.setProperty("--mdc-theme-on-primary", "#ffffff");
    }

    const storedTheme = localStorage.getItem(STORAGE_KEY_THEME);
    let isLight;
    if (storedTheme === "light") {
      isLight = true;
    } else if (storedTheme === "dark") {
      isLight = false;
    } else {
      try {
        isLight = await window.go.main.App.GetAppsUseLightTheme();
      } catch (e) {
        console.error("Failed to load Windows theme preference:", e);
        isLight = false;
      }
    }
    if (isLight) {
      document.documentElement.classList.add("light-theme");
      document.documentElement.classList.remove("dark-theme");
    } else {
      document.documentElement.classList.add("dark-theme");
      document.documentElement.classList.remove("light-theme");
    }
  };

  const syncAppState = async () => {
    await loadKeys();
    await applyWindowsTheme();
  };

  const loadSettings = async () => {
    await window.go.main.App.GetSettings()
      .then((savedata) => {
        settingsData = { ...savedata };
      })
      .catch((err) => {
        console.error(err);
        toast.push(err, red);
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
          toast.push(err, red);
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
    const storedLang = localStorage.getItem(STORAGE_KEY_LANG);
    lang = storedLang && storedLang !== "auto"
      ? storedLang
      : (navigator.language.startsWith("ja") ? "ja" : "en");
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
        toast.push(t[lang].addedSuccess, green);
        loadKeys();
      })
      .catch((err) => {
        console.error("addkey err:" + err);
        toast.push(err, red);
      });
  };

  const delKey = async (sha256) => {
    if (!confirm(t[lang].confirmDelete)) {
      return;
    }
    await window.go.main.App.DeleteKey(sha256)
      .then(() => {
        toast.push(t[lang].deletedSuccess, green);
        loadKeys();
        if (selectedKey && selectedKey.publickey.sha256 === sha256) {
          selectedKey = null;
          activeView = "welcome";
        }
      })
      .catch((err) => {
        if (err == "cancel") return;
        toast.push(err, red);
      });
  };

  const toggleKey = async (key) => {
    togglingKey = key.publickey.sha256;
    try {
      await window.go.main.App.ToggleKey(key.publickey.sha256);
      const idx = keys.findIndex(k => k.publickey.sha256 === key.publickey.sha256);
      if (idx !== -1) {
        keys[idx].disabled = !keys[idx].disabled;
        toast.push(keys[idx].disabled ? t[lang].disableKey : t[lang].enableKey, green);
        keys = keys;
      }
    } catch (err) {
      toast.push(t[lang].toggleFail + ": " + err, red);
    } finally {
      togglingKey = null;
    }
  };

  const copyText = (text) => {
    navigator.clipboard.writeText(text)
      .then(() => {
        toast.push(t[lang].copied, green);
      })
      .catch((err) => {
        toast.push(t[lang].failCopy + ": " + err, red);
      });
  };

  const onLoadKeysEvent = (message) => {
    if (togglingKey) return;
    loadKeys();
    console.log(message);
  };

  window.runtime.EventsOn("LoadKeysEvent", onLoadKeysEvent);
</script>

<main class="app-container">
  <!-- Left Panel: Sidebar -->
  <aside class="sidebar" style="width: {sidebarWidth}px; min-width: {sidebarWidth}px; max-width: {sidebarWidth}px;" data-wails-no-drag>
    <!-- Keys list -->
    <div class="sidebar-content">
      <div class="keys-section-title">{t[lang].keysTitle}</div>
      {#if keys.length === 0}
        <div class="empty-keys-message">{t[lang].emptyKeys}</div>
      {:else}
        <ul class="keys-list">
          {#each keys as key}
            <li class="key-item {selectedKey && selectedKey.publickey.sha256 === key.publickey.sha256 ? 'active' : ''} {key.disabled ? 'disabled' : ''}" on:click={() => selectKey(key)}>
              <span class="material-icons key-icon">key</span>
              <div class="key-info">
                <span class="key-name">{key.name || t[lang].unnamedKey}</span>
                <span class="key-type">{key.publickey.type || t[lang].unknownType}</span>
              </div>
            </li>
          {/each}
        </ul>
      {/if}
    </div>

    <!-- Sidebar footer action buttons -->
    <div class="sidebar-footer">
      {#if !settingsData.ProxyModeOfNamedPipe}
        <Button class="action-btn add-btn" variant="raised" on:click={() => addFileDialog.show()}>
          <span class="material-icons">add</span>
          <Label>{t[lang].addPrivateKey}</Label>
        </Button>
      {/if}
      <Button class="action-btn settings-btn" variant="outlined" on:click={openSettings}>
        <span class="material-icons">settings</span>
        <Label>{t[lang].settings}</Label>
      </Button>
    </div>
  </aside>

  <!-- Resizer Handle -->
  <div class="sidebar-resizer {isResizing ? 'resizing' : ''}" on:mousedown={startResize}></div>

  <!-- Right Panel: Main Content Area -->
  <section class="main-content" data-wails-no-drag>
    {#if activeView === 'welcome'}
      <div class="welcome-view">
        <span class="material-icons welcome-icon">security</span>
        <h2>{t[lang].welcomeTitle}</h2>
        <p>{t[lang].welcomeDesc}</p>
      </div>
    {:else if activeView === 'detail' && selectedKey}
      <div class="detail-view">
        <div class="detail-header">
          {#if !settingsData.ProxyModeOfNamedPipe}
            <Switch
              checked={!selectedKey.disabled}
              value="Toggle key enabled/disabled"
              on:SMUISwitch:change={() => toggleKey(selectedKey)}
            />
          {/if}
          <h2>{selectedKey.name || t[lang].unnamedKey}</h2>
        </div>
        
        <div class="detail-fields">

          <div class="field-group">
            <span class="field-label">{t[lang].filePath}</span>
            <div class="field-with-action">
              <Textfield disabled style="flex: 1;" value={selectedKey.filePath} />
              <Button variant="outlined" on:click={() => copyText(selectedKey.filePath)}>{t[lang].copy}</Button>
            </div>
          </div>

          <div class="field-group">
            <span class="field-label">{t[lang].keyType}</span>
            <Textfield disabled style="width: 100%;" value={selectedKey.publickey.type} />
          </div>

          <div class="field-group">
            <span class="field-label">{t[lang].fingerprintSha256}</span>
            <Textfield disabled style="width: 100%;" value={selectedKey.publickey.sha256} />
          </div>

          <div class="field-group">
            <span class="field-label">{t[lang].fingerprintMd5}</span>
            <Textfield disabled style="width: 100%;" value={selectedKey.publickey.md5} />
          </div>

          <div class="field-group">
            <span class="field-label">{t[lang].publicKey}</span>
            <div class="textarea-with-action">
              <Paper variant="outlined" class="publickey-box">
                <Content>{selectedKey.publickey.string}</Content>
              </Paper>
              <Button variant="outlined" on:click={() => copyText(selectedKey.publickey.string)}>{t[lang].copyKey}</Button>
            </div>
          </div>
        </div>

        {#if !settingsData.ProxyModeOfNamedPipe}
          <div class="detail-actions">
            <Button class="delete-btn" variant="outlined" on:click={() => delKey(selectedKey.publickey.sha256)}>
              <span class="material-icons">delete</span>
              <Label>{t[lang].deleteKey}</Label>
            </Button>
          </div>
        {/if}
      </div>
    {:else if activeView === 'settings'}
      <Settings lang={lang} on:save={handleSettingsSave} on:cancel={handleSettingsCancel} />
    {/if}
  </section>

  <!-- Dialogs & Toasts -->
  <AddFileDialog lang={lang} bind:this={addFileDialog} on:eventAddPkfile={handleData} />
  <SvelteToast />
</main>

<!-- Mouse triggers for syncing keyrings when window becomes active -->
<svelte:body on:mouseenter={syncAppState} on:mouseleave={syncAppState} />
