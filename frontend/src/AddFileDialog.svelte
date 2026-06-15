<script>
  import { createEventDispatcher } from "svelte";
  import { Button, TextBox, ToggleSwitch, ContentDialog, InfoBar } from "fluent-svelte";

  export let lang = "en";

  const t = {
    en: {
      addKeyTitle: "Add a Private key",
      addKeyFile: "Private key file",
      addKeyOpenFile: "Open file",
      addKeyType: "key type",
      addKeyEncryption: "Encryption?",
      addKeyEncrypted: "Encrypted with passphrase",
      addKeyNotEncrypted: "Not encrypted",
      addKeyPassphrase: "passphrase",
      addKeyCheck: "check",
      addKeyAdd: "Add",
      addKeyCancel: "Cancel",
      decryptedSuccess: "Decrypted secret key"
    },
    ja: {
      addKeyTitle: "秘密鍵の追加",
      addKeyFile: "秘密鍵ファイル",
      addKeyOpenFile: "ファイルを開く",
      addKeyType: "鍵の種類",
      addKeyEncryption: "暗号化",
      addKeyEncrypted: "パスフレーズで暗号化されています",
      addKeyNotEncrypted: "暗号化されていません",
      addKeyPassphrase: "パスフレーズ",
      addKeyCheck: "チェック",
      addKeyAdd: "追加",
      addKeyCancel: "キャンセル",
      decryptedSuccess: "秘密鍵の復号に成功しました"
    }
  };

  let open = false;
  let addButton = false;
  let adding = false;

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

  const newPkfile = () => {
    return {
      filePath: "",
      type: "",
      encryption: false,
      passphrase: "",
      publickey: {
        type: "",
      },
    };
  };
  let pkFile = newPkfile();

  $: keytype = [pkFile.fileType, pkFile.publickey.type].filter(Boolean).join(":");
  $: canAdd = !adding && pkFile.filePath && (addButton || pkFile.encryption) && (!pkFile.encryption || pkFile.passphrase.length > 0);

  const dispatch = createEventDispatcher();
  async function add() {
    if (!canAdd) return;
    adding = true;
    try {
      if (pkFile.encryption) {
        const pass = pkFile.passphrase;
        const file = await window.go.main.App.CheckKeyType(pkFile.filePath, pass);
        pkFile = { ...file };
        pkFile.passphrase = pass;
      }
      dispatch("eventAddPkfile", pkFile);
      open = false;
      addButton = false;
      pkFile = newPkfile();
    } catch (err) {
      console.error("add key error:" + err);
      addButton = false;
      showNotification("Error adding key", err.message || err, "critical");
    } finally {
      adding = false;
    }
  }
  const openFile = async () => {
    await window.go.main.App.OpenFile()
      .then((file) => {
        pkFile.filePath = file;
        pkFile.passphrase = "";
        checkKeyType();
      })
      .catch((err) => {
        console.error("OpenFile error:" + err);
        addButton = false;
        showNotification("Error opening file", err.message || err, "critical");
      });
  };
  const checkKeyType = async () => {
    await window.go.main.App.CheckKeyType(pkFile.filePath, pkFile.passphrase)
      .then((file) => {
        let pass = pkFile.passphrase;
        pkFile = { ...file };
        pkFile.passphrase = pass;
        console.debug(pkFile);
        addButton = true;
        if (pkFile.encryption && pkFile.passphrase.length > 0) {
          showNotification(t[lang].decryptedSuccess, "", "success");
        }
        if (pkFile.encryption && pkFile.passphrase.length == 0) {
          addButton = false;
        }
      })
      .catch((err) => {
        console.error("checkKeyTpye error:" + err);
        addButton = false;
        showNotification("Error checking key", err.message || err, "critical");
      });
  };

  export function show(filePath = "") {
    addButton = false;
    pkFile = newPkfile();
    infoBarOpen = false;
    open = true;
    if (filePath) {
      pkFile.filePath = filePath;
      checkKeyType();
    }
  }
</script>

<ContentDialog
  bind:open
  title={t[lang].addKeyTitle}
  style="width: 650px; max-width: calc(100vw - 32px);"
>
  <div style="display: flex; flex-direction: column; gap: 16px;">
    {#if infoBarOpen}
      <InfoBar
        bind:open={infoBarOpen}
        severity={infoBarSeverity}
        title={infoBarTitle}
        message={infoBarMessage}
        closable={true}
      />
    {/if}

    <div style="display: flex; flex-direction: column; gap: 12px;">
      <div>
        <span style="font-size: 13px; font-weight: 500; display: block; margin-bottom: 4px;">{t[lang].addKeyFile}</span>
        <div style="display: flex; gap: 8px;">
          <TextBox
            readonly
            value={pkFile.filePath}
            placeholder=".ppk, id_rsa..."
            style="flex: 1;"
          />
          <Button on:click={openFile} style="min-width: 36px; padding: 0;" title={t[lang].addKeyOpenFile}>
            <span class="material-icons" style="font-size: 18px;">folder_open</span>
          </Button>
        </div>
      </div>

      <div>
        <span style="font-size: 13px; font-weight: 500; display: block; margin-bottom: 4px;">{t[lang].addKeyType}</span>
        <TextBox
          readonly
          value={keytype}
          placeholder="private key type"
        />
      </div>

      <div style="display: flex; align-items: center; gap: 12px;">
        <ToggleSwitch
          bind:checked={pkFile.encryption}
          disabled
        />
        <span style="font-size: 13px;">
          {pkFile.encryption
            ? t[lang].addKeyEncrypted
            : t[lang].addKeyNotEncrypted}
        </span>
      </div>

      {#if pkFile.encryption}
        <div>
          <span style="font-size: 13px; font-weight: 500; display: block; margin-bottom: 4px;">{t[lang].addKeyPassphrase}</span>
          <TextBox
            bind:value={pkFile.passphrase}
            type="password"
            placeholder="passphrase of private key"
          />
        </div>
      {/if}
    </div>
  </div>

  <svelte:fragment slot="footer">
    {#if addButton || pkFile.encryption}
      <Button variant="accent" on:click={add} disabled={!canAdd}>
        {t[lang].addKeyAdd}
      </Button>
    {/if}
    <Button on:click={() => (open = false)}>
      {t[lang].addKeyCancel}
    </Button>
  </svelte:fragment>
</ContentDialog>
