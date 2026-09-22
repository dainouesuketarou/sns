# 第2週 設計メモ: ユーザー・フォロー集約とユースケース層

`docs/design_review_2026-09-23.md` のレビューを反映した確定版。
参照: `docs/rules.md`（不変条件と決定済み方針）、`internal/domain/post/`（先週の型）。

---

## 1. ユーザー集約

### 1-1. UserName を値オブジェクトにするか（3条件で判定）

- 常に成り立っていてほしいルールがあるか: ある。空でない / 20文字以内 / 文字種 `[a-zA-Z0-9_]`。
  いずれも値そのものだけで判定できる（UNIQUE は他の行を見ないと判断できないので DB の担当、ここには含めない）
- そのルールを複数箇所で使うか: 使う。作成時（`NewUser`）と変更時（`ChangeUsername`）の両方
- 破られたら致命的か: 致命的。`@name` でのメンション・URL に使うため、不正な文字が入ると表示やルーティングが壊れる
- 判定: **値オブジェクトにする**（`UserName`）。ルールが3つあり、作成時と変更時の2箇所で使うため

### 1-2. User エンティティ

- フィールド: `users` テーブルと同じ5つ。`id uuid.UUID` / `name UserName` / `description string` / `avatarURL string` / `createdAt time.Time`
  - 5つとも持つ。`created_at` は現時点でルールが読まないが、`Post` と形を揃える方を優先する
  - 将来持たせないもの: `password_hash`（認証は別集約。持つと User を扱う全コードが秘密情報を引き回す）、
    `follower_count` などの非正規化した集計値（真実は follows 側。リードモデルで返す）
- コンストラクタで守るルール（`NewUser`）:
  - `UserName` 値オブジェクトを受け取る（空・長さ・文字種の検証は `NewUserName` 側で済んでいる）
  - `description` が500文字以内（文字数。`utf8.RuneCountInString`）
  - `avatarURL` が空でなければ http(s) の URL 形式、かつ2000文字以内
  - `id` は内部生成なので Nil チェック不要。`createdAt` は引数で受け取る（Clock 方針）
- `RebuildUser` の要否: **要る**。`GetUserByID` で DB の行から組み立てるため。
  フィールドが非公開なので `user` パッケージ外からは関数経由でしか作れない。
  `RebuildUserName`（検証なし）も対で用意する（`RebuildPostContent` と同じ方針）
- 状態を変える操作: **今週は `UpdateProfile` のみ**（description / avatarURL を変更）。
  退会（`DeleteUserAccount`）は今週やらない。`rules.md` の決定どおり論理削除にするので、
  `users.deleted_at` の追加・全クエリへの `deleted_at is null`・「退会者をフォローできない」等の
  ルール決めが必要になり、今週の範囲を超えるため。別カードで対応する
  ※ `UpdateProfile` も今週のユースケース一覧には入れない（3-3 参照）
- ドメインエラーの一覧:
  ```
  # ドメイン層（コンストラクタが返す）
  user.ErrEmptyUsername          // 空・空白のみ
  user.ErrUsernameTooLong        // 20文字超
  user.ErrInvalidUsernameChars   // [a-zA-Z0-9_] 以外
  user.ErrDescriptionTooLong     // 500文字超
  user.ErrInvalidAvatarURL       // http(s) URL でない
  user.ErrAvatarURLTooLong       // 2000文字超

  # リポジトリ層（pgx エラーの翻訳先）
  user.ErrUsernameAlreadyExists  // 23505 の翻訳
  user.ErrUserNotFound           // ErrNoRows の翻訳（「探したが居ない」）
  ```

### 1-3. User のリポジトリ interface

- メソッド一覧: **今週は `Create` と `FindByID` のみ**
  - `FindByUsername` は、username の可用性チェック UI を作るときに足す（今週は作らない）
  - `Update` / `Delete` は今週は作らない（1-2 の決定に合わせる）
- username 重複時のエラーをどこで検出し、どのドメインエラーに翻訳するか:
  `UserRepository.Create` の中で検出する。事前に SELECT で存在確認はしない
  （2人が同時に登録すると両方「居ない」と判断して結局 INSERT で落ちるため、
  どのみち INSERT 時の翻訳が要る。クエリも1回増える）。
  INSERT が SQLSTATE 23505（unique_violation）で失敗したら
  `user.ErrUsernameAlreadyExists` に翻訳して返す。元の pgx エラーは `%w` でチェーンに残す。
  `post_repository.go` の `translateCreateError` と同じ形。
  ※ users の UNIQUE は今 username だけなので 23505 = username 重複と断定できる。
    将来 email などに UNIQUE を足したら `pgErr.ConstraintName` で分岐する。

---

## 2. フォロー集約

### 2-1. 集約の境界

**独立した集約**。理由:
1. 整合性を守る範囲が `follows` の1行に閉じており、`users` の行は変更されない
2. A→B のフォローは A の持ち物とも B の持ち物とも言えず、どちらかの集約に押し込むと不自然
3. `User` が `[]Follow` を抱えると、`FindByID` のたびにフォロー全件を読むことになり、
   フォロー数の多いユーザーで破綻する

集約同士は ID で参照する（`Follow` は `follower_id` / `followee_id` という **ID** で User を指す。
`*user.User` は持たない）。

### 2-2. Follow エンティティ

- フィールド（Go の型で。id は先週の決定どおり Go 側で生成し、DB のデフォルトは使わない）:
  ```
  id         uuid.UUID
  followerID uuid.UUID
  followeeID uuid.UUID
  createdAt  time.Time
  ```
- コンストラクタで守るルール:
  - 自己フォロー禁止（`followerID != followeeID`）
  - `followerID` / `followeeID` が `uuid.Nil` でない
  - ※ この2つしか無いのは正常。`Follow` は「誰が誰を、いつ」だけの関連なので、
    引数の中だけで判定できるルースが元々少ない
- DB だけで守るルール:
  - 重複フォロー禁止（複合 UNIQUE）。他の行を見ないと判断できないため。23505 を翻訳する
  - 両者が実在すること（FOREIGN KEY）。別テーブルを見ないと判断できないため。23503 を翻訳する
  - ※ コンストラクタで事前 SELECT はしない。ドメイン層が DB を知ることになる上、
    SELECT から INSERT までの間に競合すれば結局重複する（原子的に守れるのは DB だけ）
- ドメインエラーの一覧:
  ```
  # ドメイン層
  follow.ErrSelfFollow           // follower == followee
  follow.ErrMissingUserID        // どちらかが uuid.Nil

  # リポジトリ層
  follow.ErrAlreadyFollowing     // 23505（follows_follower_id_followee_id_key）
  follow.ErrFolloweeNotFound     // 23503（follows_followee_id_fkey）フォロー相手が実在しない
  follow.ErrFollowerNotFound     // 23503（follows_follower_id_fkey）自分が実在しない＝異常系
  follow.ErrFollowNotFound       // 解除対象が無い（Delete が0件）
  ```
  ※ `user.ErrUserNotFound` と名前を分けている。API 層で
  「そのユーザーは存在しません」（Followee）と、認証済みのはずの自分が居ない異常系（Follower）と、
  検索で見つからない（user）を区別してメッセージを変えるため。
  `pgErr.ConstraintName` で FK のどちらが違反したか判別する

### 2-3. Follow のリポジトリ interface

- メソッド一覧: `Create` / `Delete`
- 「解除」は今週必要か: **やる**（`UnfollowUser` ユースケースとして）

---

## 3. ユースケース層の責務

### 3-1. 責務の言語化

やること:
- 入力を受け取り、ドメインオブジェクトを組み立てる（`NewPostContent` → `NewPost` を順に呼ぶ）
- リポジトリを呼ぶ順番を決める（手順・段取り）
- トランザクションの範囲を決める（先週の3役の「範囲を決める」）。※ 今週は該当なし（3-2 参照）
- 外界の値（現在時刻）を取得してドメインに渡す。
  ただし `time.Now()` を直接呼ばず、Clock（`func() time.Time`）を注入してもらい、
  それを呼んで得た時刻を `NewPost` などに渡す。
  本番は `main.go` で `time.Now` を配線し、テストは固定時刻を返す関数を渡す。
  理由: ドメイン層から `time.Now()` を追い出した（レビュー指摘8）のと同じ問題が、
  ユースケース層で直接呼ぶと再発し、実行時刻に依存するテストになるため。
  時計は DB 接続やロガーと同じ「外界からの入力」として境界で注入する。
  → 今後 `time.Now()` が必要な箇所はすべてこの方式に揃える
- 複数の集約をまたぐ判断（「投稿が公開中ならコメント可」のような、`rules.md` の第3分類）
- 結果を返す

やらないこと:
- 単一オブジェクトで判定できるルールを書く（空文字チェックなど） → ドメインの仕事
- SQL を書く、pgx / pgtype を触る → リポジトリの仕事
- HTTP / GraphQL の型を知る、ステータスコードを決める、エラー文言を作る → API 層の仕事
- DB のエラーコード（23505）を見る → リポジトリが翻訳済みのはず

ドメインエラー以外のエラー（DB 断など）の扱い:
`fmt.Errorf("create post: %w", err)` で文脈だけ足して返す。
リトライするかはユースケースには判断できず、ログは API 層で一元的に出すため、ここでは判断もログもしない。

### 3-2. 「投稿を作る」ユースケース

- 受け取るもの: **専用の入力型を作る**（`usecase.CreatePostInput`）。
  gqlgen が生成した型をそのまま受けると、ユースケース層が GraphQL 層を import することになり
  依存が内側→外側に逆流する。また引数を並べる方式は `body` と `imageURL` がどちらも string で
  取り違えてもコンパイルが通るため避ける。
  ```
  type CreatePostInput struct {
      UserID   uuid.UUID
      Body     string
      ImageURL string
      IsPublic bool
  }
  ```
  DTO なのでルールは持たせない（検証はドメインの仕事）
- 呼ぶもの（順番に）:
  1. `now := u.clock()` で現在時刻を取得
  2. `content, err := post.NewPostContent(in.Body, in.ImageURL)`（本文・画像のルールはここで検証済み）
  3. `p, err := post.NewPost(in.UserID, content, in.IsPublic, now)`
  4. `created, err := u.postRepo.Create(ctx, p)`
  5. `created` を返す
- 返すもの: `*post.Post`（DB に保存された状態）。
  Resolver 側で表示用の型に変換する。ドメインオブジェクトを返すのは依存の向きとして正しい
- トランザクションは要るか: **不要**。今週のユースケースはどれも単一集約に1回書き込むだけで、
  複数集約にまたがるものが無い。`Transactor` は作らない（レビュー指摘16 の対応も先送り）。
  「投稿したらフォロワー全員に通知を作る」のような機能が来たときに初めて必要になる
- `time.Now()` はどこで呼ぶか: 呼ばない。注入された Clock を `CreatePost` が呼び、その値を
  `NewPost` に渡す。`main.go` で `time.Now` を配線する

### 3-3. 今週書くユースケース一覧

- `CreateUser`: ユーザー登録。`NewUserName` → `NewUser` → `userRepo.Create`
- `CreatePost`: 投稿する（先週のドメイン層を初めてユースケースから呼ぶ）
- `FollowUser`: フォローする。`NewFollow` → `followRepo.Create`
- `UnfollowUser`: フォロー解除。`followRepo.Delete`
- `GetTimeline`: フォロー中タイムライン取得
- `GetGlobalTimeline`: 全体タイムライン取得

タイムライン2本は他と性質が違う。読み取り専用で、ドメインオブジェクトを組み立てず
リードモデル（表示用の構造体）を返す。`post.Post` の配列に組み立て直すと投稿者の username を
別途引く必要が出て非効率になるため。

今週作らないもの: `UpdateProfile`（ドメインのメソッドは用意するがユースケースは次週以降）、
退会、username 可用性チェック。

### 3-4. ディレクトリ

- ユースケースを置くパッケージ名とパス: `internal/usecase/`（パッケージ名 `usecase`）。
  1ユースケース1ファイル（`create_user.go` / `create_post.go` / `follow_user.go` /
  `unfollow_user.go` / `get_timeline.go`）。入力 DTO は各ユースケースと同じファイルに置く
- `Transactor` のような interface を置く場所: `internal/usecase/`（使う側に置く原則）。
  ただし今週は `Transactor` 自体を作らない。`Clock`（`func() time.Time`）も同じ場所に置く

---

## 4. 今週書くクエリ一覧

各行: クエリ名 / sqlc の種別 / 用途 / WHERE・JOIN で使う列 / その列に効くインデックス

- `CreateUser` / `:one` / ユーザー登録 / WHERE なし / -
  `returning *` で保存後の行を返す（`CreatePost` と同じ方針）。23505 → `ErrUsernameAlreadyExists`

- `GetUserByID` / `:one` / ユーザー取得 / `id` / PK（自動）

- `CreateFollow` / `:one` / フォロー / WHERE なし / -
  23505 → `ErrAlreadyFollowing`、23503 → `ErrFolloweeNotFound` / `ErrFollowerNotFound`
  （`pgErr.ConstraintName` で判別）

- `DeleteFollow` / `:exec` / フォロー解除 / `follower_id`, `followee_id` / follows の複合 UNIQUE
  ※ 0件削除を検出して `ErrFollowNotFound` を返したいなら `:execrows` にする

- `GetTimeline` / `:many` / フォロー中タイムライン / `follows.follower_id`（WHERE）,
  `posts.user_id`（JOIN）, `posts.created_at`+`posts.id`（ORDER BY / カーソル）/
  `follows.follower_id` は複合 UNIQUE の先頭列で効く。`posts.user_id` は `posts_user_id_idx` あり

- `GetGlobalTimeline` / `:many` / 全体タイムライン / `posts.created_at`+`posts.id`（ORDER BY / カーソル）/
  **効くインデックスが無い**。`posts(created_at desc, id desc)` を追加する必要がある

### タイムラインの仕様

- 2本作る: `GetTimeline`（フォロー中のみ、follows と JOIN）と `GetGlobalTimeline`（全ユーザー、JOIN なし）。
  SQL も効くインデックスも別物なので、クエリを分けて書く
- 自分の投稿を含むか: **含む**（`GetTimeline` では `follows` に加えて自分の user_id も対象にする）
- `is_public = false` の扱い: **SQL の WHERE で除外する**。
  `where (p.is_public = true or p.user_id = $1)` … 自分の非公開投稿だけは自分に見える。
  Go 側で弾くと、非公開が多いほど無駄な行を読み、LIMIT 20 件を返すのに何件読めばいいか
  決まらなくなる（ページングが壊れる）
- 並び順とページング: **カーソル方式**（GraphQL を使うため。Relay Connection 仕様に寄せる）。
  `order by created_at desc, id desc` / `where (created_at, id) < ($2, $3)` / `limit $4`。
  `created_at` だけだと同時刻の投稿を取りこぼすので、`(created_at, id)` の組で一意にする。
  カーソル文字列はこの2つを base64 エンコードしたもの。1ページ20件
- 使うインデックス:
  - `GetTimeline`: `follows(follower_id, followee_id)` の複合 UNIQUE（先頭列で絞る）+ `posts_user_id_idx`。
    ただし `created_at` の ORDER BY は Sort が入る（先週の EXPLAIN で確認済み:
    Hash Join → Sort → Limit）。改善するなら `posts(user_id, created_at desc)` の複合インデックス
  - `GetGlobalTimeline`: `posts(created_at desc, id desc)` を新規に追加する
  - どちらも実装後に `explain analyze` で確認する

---

## 5. 実装の順番（次カード用）

1. マイグレーション追加（`posts(created_at desc, id desc)` インデックス）
2. `internal/domain/user/`（`UserName` → `User` → `Repository` interface → エラー）
3. `internal/domain/follow/`（`Follow` → `Repository` interface → エラー）
4. `db/queries/user.sql` / `follow.sql` / `timeline.sql` → `sqlc generate`
5. `internal/repository/`（`user_repository.go` / `follow_repository.go` / `timeline_query.go`）
6. `internal/usecase/`（Clock、各ユースケース）
7. テスト（ドメインは単体、リポジトリは実 DB、ユースケースは固定 Clock）
