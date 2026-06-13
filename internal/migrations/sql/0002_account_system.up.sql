-- Yasumi MVP first-party account system.
-- Rollback stance: forward-only for retained environments because synced user
-- data ownership depends on users.id after this migration. Disposable local
-- databases may be reset instead of rolled back.

create table if not exists users (
	id uuid primary key,
	username text not null,
	email text not null,
	email_verified_at timestamptz null,
	display_name text null,
	status text not null default 'active',
	created_at timestamptz not null,
	updated_at timestamptz not null,
	constraint users_username_non_empty check (btrim(username) <> ''),
	constraint users_email_non_empty check (btrim(email) <> ''),
	constraint users_status_allowed check (status in ('active', 'disabled'))
);

create unique index if not exists users_username_ci_unique on users (lower(username));
create unique index if not exists users_email_ci_unique on users (lower(email));

create table if not exists user_credentials (
	user_id uuid primary key references users(id) on delete cascade,
	password_hash text not null,
	password_hash_algorithm text not null,
	password_hash_params jsonb not null default '{}'::jsonb,
	password_changed_at timestamptz not null,
	created_at timestamptz not null,
	updated_at timestamptz not null,
	constraint user_credentials_hash_non_empty check (btrim(password_hash) <> ''),
	constraint user_credentials_algorithm_allowed check (password_hash_algorithm in ('argon2id'))
);

create table if not exists user_sessions (
	id uuid primary key,
	user_id uuid not null references users(id) on delete cascade,
	refresh_token_hash text not null,
	created_at timestamptz not null,
	last_used_at timestamptz null,
	expires_at timestamptz not null,
	revoked_at timestamptz null,
	replaced_by_session_id uuid null references user_sessions(id),
	user_agent text null,
	ip_hash text null,
	constraint user_sessions_refresh_hash_non_empty check (btrim(refresh_token_hash) <> '')
);

create unique index if not exists user_sessions_refresh_hash_unique on user_sessions (refresh_token_hash);
create index if not exists user_sessions_user_active_idx on user_sessions (user_id, revoked_at, expires_at);

insert into users (id, username, email, display_name, status, created_at, updated_at)
select distinct user_id,
	'legacy_' || replace(user_id::text, '-', ''),
	'legacy+' || replace(user_id::text, '-', '') || '@local.invalid',
	'Legacy User',
	'active',
	now(),
	now()
from (
	select user_id from areas
	union
	select user_id from recurring_task_templates
	union
	select user_id from items
	union
	select user_id from operation_history
	union
	select user_id from user_settings
) existing_users
on conflict (id) do nothing;

alter table areas
	add constraint areas_user_fk foreign key (user_id) references users(id);

alter table recurring_task_templates
	add constraint recurring_templates_user_fk foreign key (user_id) references users(id);

alter table items
	add constraint items_user_fk foreign key (user_id) references users(id);

alter table operation_history
	add constraint operation_history_user_fk foreign key (user_id) references users(id);

alter table user_settings
	add constraint user_settings_user_fk foreign key (user_id) references users(id);
