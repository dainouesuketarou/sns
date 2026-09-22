create table users (
  id uuid primary key default gen_random_uuid(),
  username text not null unique,
  description text,
  avatar_url text,
  created_at timestamptz not null default now()
);

create table posts (
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null,
  body text,
  image_url text,
  is_public boolean not null,
  created_at timestamptz not null default now(),
  foreign key (user_id) references users (id),
  constraint posts_body_or_image_check
    check (btrim(coalesce(body, '')) <> '' or btrim(coalesce(image_url, '')) <> '')
);

create index posts_user_id_idx on posts (user_id);

create table comments (
  id uuid primary key default gen_random_uuid(),
  post_id uuid not null,
  user_id uuid not null,
  body text not null,
  created_at timestamptz not null default now(),
  foreign key (post_id) references posts (id),
  foreign key (user_id) references users (id)
);

create index comments_post_id_idx on comments (post_id);

create table follows (
  id uuid primary key default gen_random_uuid(),
  follower_id uuid not null,
  followee_id uuid not null,
  created_at timestamptz not null default now(),
  foreign key (follower_id) references users (id),
  foreign key (followee_id) references users (id),
  unique (follower_id, followee_id),
  constraint follows_no_self_follow_check check (follower_id <> followee_id)
);

create index follows_followee_id_idx on follows (followee_id);

create table likes (
  id uuid primary key default gen_random_uuid(),
  post_id uuid not null,
  user_id uuid not null,
  created_at timestamptz not null default now(),
  foreign key (post_id) references posts (id),
  foreign key (user_id) references users (id),
  unique (post_id, user_id)
);

create table notifications (
  id uuid primary key default gen_random_uuid(),
  recipient_id uuid not null,
  actor_id uuid not null,
  type text not null,
  post_id uuid,
  comment_id uuid,
  is_read boolean not null,
  created_at timestamptz not null default now(),
  foreign key (recipient_id) references users (id),
  foreign key (actor_id) references users (id),
  foreign key (post_id) references posts (id),
  foreign key (comment_id) references comments (id)
);

create index notifications_recipient_id_idx on notifications (recipient_id);
