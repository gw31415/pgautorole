# pgautorole - Discord bot

- [INVITE LINK](https://discord.com/oauth2/authorize?client_id=1252439534747254824&permissions=268435456&integration_type=0&scope=bot+applications.commands)
  - 現在開発者 (Ama) のみ招待できるようになっています。

## 権限について

- 必要な権限は「ロールの管理」です。「メンバー」と「新規会員」に割り当てたロールを操作するので、このボットのロールをそれらよりも上位にする必要があります。
- 備考：ボットの内部設定で `SERVER MEMBERS INTENT` が有効になっています。

## 開発

### 環境変数

設定すべき環境変数は [`.env.example`](.env.example) に記載されています。適当な値を設定して `.env` という名前で保存してください。
各変数の説明はコメントに記載されています。

### セットアップ

1. [`mise`](https://mise.jdx.dev/) をインストールします。
2. `.env`を作成します。
3. `mise` の設定ファイルを信頼します。
   ```bash
   mise trust
   ```
4. 依存パッケージをインストールします。
   ```bash
   mise install
   ```
