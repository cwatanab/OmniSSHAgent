<script>
  import { createEventDispatcher } from "svelte";
  import { Title, Content } from "@smui/paper";
  import Dialog, { Actions } from "@smui/dialog";
  import Button, { Label } from "@smui/button";
  import Textfield from "@smui/textfield";
  import HelperText from "@smui/textfield/helper-text";
  import Card from "@smui/card";
  import FormField from "@smui/form-field";
  import Switch from "@smui/switch";
  import { toast } from "@zerodevx/svelte-toast";

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

  $: keytype = pkFile.fileType + ":" + pkFile.publickey.type;

  const dispatch = createEventDispatcher();
  function add() {
    dispatch("eventAddPkfile", pkFile);
    open = false;
    addButton = false;
    pkFile = newPkfile();
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
        toast.push(err, red);
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
          toast.push(t[lang].decryptedSuccess, green);
        }
        if (pkFile.encryption && pkFile.passphrase.length == 0) {
          addButton = false;
        }
      })
      .catch((err) => {
        console.error("checkKeyTpye error:" + err);
        addButton = false;
        toast.push(err, red);
      });
  };

  export function show() {
    open = true;
  }
</script>

<Dialog
  bind:open
  scrimClickAction=""
  escapeKeyAction=""
  surface$style="width: 850px; max-width: calc(100vw - 32px);"
  aria-labelledby="mandatory-title"
  aria-describedby="mandatory-content"
>
  <div class="dialog">
    <Title id="mandatory-title">{t[lang].addKeyTitle}</Title>
    <Content id="mandatory-content">
      <Card padded>
        <div>
          <div>
            <FormField style="width: 100%;">
              <Textfield
                disabled
                value={pkFile.filePath}
                label={t[lang].addKeyFile}
                style="width: 100%;"
                helperLine$style="width: 100%;"
              >
                <HelperText slot="helper">.ppk, id_rsa...</HelperText>
              </Textfield>
            </FormField>
          </div>
          <div>
            <Button on:click={openFile} variant="raised">
              <Label>{t[lang].addKeyOpenFile}</Label>
            </Button>
          </div>
          <div>
            <FormField style="width: 100%;">
              <Textfield
                disabled
                value={keytype}
                label={t[lang].addKeyType}
                style="width: 100%;"
                helperLine$style="width: 100%;"
              >
                <HelperText slot="helper">private key type</HelperText>
              </Textfield>
            </FormField>
          </div>
          <div>
            <FormField>
              <Switch
                bind:checked={pkFile.encryption}
                disabled
                value={t[lang].addKeyEncryption}
              />
              <span
                >{pkFile.encryption
                  ? t[lang].addKeyEncrypted
                  : t[lang].addKeyNotEncrypted}</span
              >
            </FormField>
          </div>
          {#if pkFile.encryption}
            <FormField style="width: 100%;">
              <Textfield
                bind:value={pkFile.passphrase}
                type="password"
                label={t[lang].addKeyPassphrase}
                style="width: 100%;"
                helperLine$style="width: 100%;"
              >
                <HelperText slot="helper">passphrase of private key</HelperText>
              </Textfield>
            </FormField>
            <Button on:click={checkKeyType}>
              <Label>{t[lang].addKeyCheck}</Label>
            </Button>
          {/if}
        </div>
      </Card>
    </Content>
    <Actions>
      {#if addButton}
        <Button on:click={add}>
          <Label>{t[lang].addKeyAdd}</Label>
        </Button>
      {/if}
      <Button on:click={() => (open = false)}>
        <Label>{t[lang].addKeyCancel}</Label>
      </Button>
    </Actions>
  </div>
</Dialog>

<style>
  .dialog {
    margin-left: 8px;
    margin-right: 8px;
  }
</style>
