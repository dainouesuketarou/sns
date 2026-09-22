# 不変条件の置き場所

「守りたいルール」を、DB制約・ドメインクラス・ユースケース層のどこで守るか整理したメモ。

## DB制約で守るルール（実装済み: 000001_create_initial_tables）

- users.username: UNIQUE（ユーザー名は一意）
- likes (post_id, user_id): 複合UNIQUE（1人1投稿1いいね）
- follows (follower_id, followee_id): 複合UNIQUE（重複フォロー禁止）
- posts: CHECK `posts_body_or_image_check`（`btrim(coalesce(body,'')) <> '' or btrim(coalesce(image_url,'')) <> ''`。本文か画像どちらか必須。NULL も空文字も空白のみも「無し」として扱う。body は nullable にして、この二択を CHECK で表現）
- follows: CHECK `follows_no_self_follow_check`（`follower_id <> followee_id`。自己フォロー禁止。ドメインクラスと二重で守る）
- 各テーブルの必須カラム（posts.body を除く）: NOT NULL
- created_at: NOT NULL + `default now()`。アプリは常に明示的に渡す（デフォルトは直接 INSERT する運用のための保険）
- 各外部キー: FOREIGN KEY（`on delete` は指定なし = NO ACTION。退会は論理削除にするので CASCADE しない）
- 外部キー列のインデックス: `posts(user_id)`、`comments(post_id)`、`follows(followee_id)`、`notifications(recipient_id)`。`likes(user_id)` は「自分がいいねした投稿」の予定が無いので未作成

CHECK 制約は、ドメインクラス側の検証と意図的に重複させている。判断軸: 「将来も変わりにくく、壊れたら致命的（不整合データがそのまま業務ロジックの前提を崩す）」なコア不変条件は DB にも二重で入れる。逆に変わりやすいルール（文字数上限など）はドメインのみ。エラーメッセージの質やテストのしやすさを優先してアプリ側の検証も必ず用意する。

## ドメインクラスで守るルール

判断基準: オブジェクト1個だけ渡されて、そのルールが真か答えられるか。

- Post（PostContent 値オブジェクト、実装済み）:
  - body と image_url のどちらか一方は必須。空白のみは「無し」とみなす。※ DB 側にも同じ意図の CHECK あり
  - body は保存時にトリムしない（ユーザーの入力を原文のまま保持する）
  - body は 1000 文字以下（バイト数ではなく文字数）。ドメインのみ、DB には入れない
  - image_url は http(s) の URL 形式。ドメインのみ
- Follow: follower_id と followee_id が同一でないこと（自己フォロー禁止）。※ DB 側にも同じ意図の CHECK あり
- Comment: body が空文字・空白のみでないこと
- User: username が許可された文字種・長さの範囲内であること

## ユースケース層で守るルール（後のフェーズ）

判断基準: 集約をまたぐが、DB制約でも表せないルール。

- コメント投稿: 対象の post が is_public = true のときだけ許可
- いいね: 対象の post が is_public = true のときだけ許可
- フォロー: 相手ユーザーが存在し、かつブロックされていないときだけ許可（ブロック機能は未実装）

## 設計方針（決定済み）

- 保存済みデータは信頼する: `Rebuild*` は検証しない。ルールを変えるときは `New*` を直し、既存データはマイグレーションで揃える（遡及するか既存を容認するかはその時に決める）
- リポジトリの `Create` は INSERT 専用。更新は別メソッドにする。`Create` は DB が返した行から組み立てた「保存後の Post」を返す
- リポジトリは pgx のエラーをドメインエラーに翻訳する（型の翻訳）。エラー文言を外部に出さないのは API 層の責務
- 退会は論理削除: `users.deleted_at` を足す（退会ユースケースのカードで対応）。FK は NO ACTION のまま。退会者の投稿は非表示にし、他人が付けたコメントは残す。猶予期間・匿名化は後のフェーズ
- トランザクション: 開く = インフラ層の TxManager、範囲を決める = ユースケース、届ける = ctx 経由。複数集約を1トランザクションで扱うユースケースが来たら実装する
- 現在時刻は Clock（`func() time.Time`）を注入する: ドメイン層もユースケース層も `time.Now()` を直接呼ばない。`main.go` で `time.Now` を配線し、テストは固定時刻を返す関数を渡す。実行時刻に依存するテストを作らないため
- 一覧取得の絞り込みは SQL の WHERE でやる: 取得後に Go 側で弾くと、LIMIT n 件を返すのに何件読めばいいか決まらずページングが壊れる
- 一覧のページングはカーソル方式: `order by created_at desc, id desc` と `(created_at, id)` の組で位置を表す。`created_at` だけだと同時刻の行を取りこぼす
- 値オブジェクトにしたものは、エンティティのゲッターでも値オブジェクトのまま返す: `AvatarURL() string` ではなく `AvatarURL() AvatarURL`。裸のプリミティブを返すと、外でそれを検証したり組み立て直したりする誘惑が生まれ、ルールが分散する
- テスト名: 関数名は英語（`TestUserRepository_FindByID_UnknownIDReturnsErrUserNotFound` のように `Test対象_条件_期待` の順）。ケースの説明（`t.Run` の第1引数）とエラーメッセージは日本語でよい。関数名を日本語にすると `go test -run` で打ちにくく、grep・補完・スタックトレースで扱いづらい
- 貧血症の検出基準: エンティティから値を取り出して外で `if` を書き始めたら、その判断はエンティティのメソッドにする。例) resolver で `if time.Since(u.CreatedAt()) < 24*time.Hour && !u.Description().IsSet()` と書きたくなったら `u.IsNewcomer(now)` として User 側に置く。ゲッターの数が多いこと自体は問題ではなく、判断がどこにあるかで見る
- 同じ意味に見えるエラーは名前で区別する: 「検索して見つからない」（`user.ErrUserNotFound`）、「フォロー相手が実在しない」（`follow.ErrFolloweeNotFound`）、「認証済みのはずの自分が実在しない＝異常系」（`follow.ErrFollowerNotFound`）は API 層で別のメッセージにするため、同名にせず分ける
