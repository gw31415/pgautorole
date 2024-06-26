# pgautorole - Discord bot

- [INVITE LINK](https://discord.com/oauth2/authorize?client_id=1252439534747254824&permissions=268435456&integration_type=0&scope=bot+applications.commands)
  - 現在開発者 (Ama) のみ招待できるようになっています。

## 機能

- [x] 「新入生」ロール管理。
  - 初回参加日時から一定期間が経過するまでの、ホワイトリストのロールがついていない「PlayGround-Member」を「新入生」とします。
  - 「PlayGround-Member」ロールに変化があった場合、新入生ロールを付与または剥奪します。
  - 「新入生」ロールの手動変更をブロックします。
  - 定期的に「新入生」ロールの更新を行います。
- [x] コース系ロールの管理。
  - 以下の名前のロールが過不足なく一つずつ存在するものを「コース」として認識します。
    - `${コース名}`
    - `${コース名}-アプレンティス`
    - `${コース名}-アシスタント`
    - `${コース名}-ノーマル`
    - `${コース名}-リード`
  - コースレベルのロールを1つのみ選べるようにします。
    - 他のレベルのロールを選んだ際、他のレベルが外れます(ラジオボタンみたいになる)。
  - コースのロールを付与された際、`${コース名}-アプレンティス`のロールを付与します。
  - コースのロールを剥奪された際、コースに関連するロールを全て剥奪します。

## 権限について

- 必要な権限は「ロールの管理」です。このボットのロールを操作するロールよりも上位にする必要があります。
  - 「PlayGround-Member」
  - 「新入生」
  - コース系ロール
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
