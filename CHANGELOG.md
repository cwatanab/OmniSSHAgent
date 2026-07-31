# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.8] - 2026-07-31

### Fixed
- ウィンドウのリサイズ時に入力フィールドおよびレイアウトが不均一に伸び縮みしていた問題を修正し、画面幅に合わせて統一してストレッチするよう改善。

### Changed
- サイドバーの鍵一覧において、無効状態の鍵ラベルの非表示化など UI スタイルを調整。

## [0.7.7] - 2026-07-29

### Fixed
- サイドバーの鍵一覧において、無効化した鍵がコンポーネントレベルの非活性化設定により操作不可能・非表示状態になる不具合を修正。
- 無効化された鍵にはグレーアウトアイコンと視覚的な状態表示を追加し、一覧上で選択・有効化の再切り替えができるように改善。

## [0.7.6] - 2026-07-21

### Added
- Windows 10/11 のモダンなトースト通知（Silent Toast）に対応。
- アプリ固有 of AppUserModelID (AUMID: `masahide.OmniSSHAgent`) のレジストリ自動登録およびアプリアイコンとの紐付け機能を追加。
- 鍵利用時の通知用に鍵アイコン（`keyicon.png`）を追加。

### Fixed
- 鍵の無効化（Deactivate）機能の動作不良を修正。
  - SSH接続試行時（SSHクライアント向けの鍵一覧返却）に、無効化された鍵が正しく除外されるように修正。
  - UI 上では無効化された鍵がリストから消えず、無効状態のまま残り続けるよう挙動を整理。
  - 鍵の無効化状態が設定ファイル（`settings.json`）に正しく保存され、再起動後も状態が維持されるように修正。

### Removed
- VSCode 関連の不要な開発設定ファイル（`.vscode/` 配下の `launch.json`, `settings.json`, `tasks.json`）の削除。
- 不要な Windows マニフェストファイルやビルド一時ファイルの整理。

[0.7.8]: https://github.com/cwatanab/OmniSSHAgent/compare/v0.7.7...v0.7.8
[0.7.7]: https://github.com/cwatanab/OmniSSHAgent/compare/v0.7.6...v0.7.7
[0.7.6]: https://github.com/cwatanab/OmniSSHAgent/compare/v0.7.5...v0.7.6
