-- Yasumi MVP synced data schema.
-- Rollback stance: this initial schema is reversible in local/test databases by
-- dropping the tables and schema_migrations row. Production rollback should
-- restore from backup because synced user data may already exist.

create table if not exists areas (
	id uuid primary key,
	user_id uuid not null,
	name text not null,
	sort_order integer not null default 0,
	created_at timestamptz not null,
	updated_at timestamptz not null,
	deleted_at timestamptz null,
	archived_at timestamptz null,
	hidden_reason text null,
	client_updated_at timestamptz not null,
	server_updated_at timestamptz not null,
	created_by_device_id text not null,
	updated_by_device_id text not null,
	revision bigint not null,
	constraint areas_user_id_id_unique unique (user_id, id),
	constraint areas_name_non_empty check (btrim(name) <> ''),
	constraint areas_hidden_reason_allowed check (
		hidden_reason is null or hidden_reason in ('converted_to_recurring_template', 'recurring_skipped')
	)
);

create table if not exists recurring_task_templates (
	id uuid primary key,
	user_id uuid not null,
	title text not null,
	note text null,
	area_id uuid null,
	frequency text not null,
	interval integer not null default 1,
	weekdays jsonb not null default '[]'::jsonb,
	recurrence_basis text not null,
	start_date date not null,
	end_type text not null default 'never',
	end_date date null,
	end_after_count integer null,
	completed_count integer not null default 0,
	next_sequence integer not null default 1,
	scheduled_time time null,
	reminder_rule jsonb not null default '{}'::jsonb,
	generated_task_defaults jsonb not null default '{}'::jsonb,
	status text not null default 'active',
	created_at timestamptz not null,
	updated_at timestamptz not null,
	deleted_at timestamptz null,
	archived_at timestamptz null,
	hidden_reason text null,
	client_updated_at timestamptz not null,
	server_updated_at timestamptz not null,
	created_by_device_id text not null,
	updated_by_device_id text not null,
	revision bigint not null,
	constraint recurring_templates_user_id_id_unique unique (user_id, id),
	constraint recurring_templates_area_user_fk foreign key (user_id, area_id) references areas(user_id, id),
	constraint recurring_templates_title_non_empty check (btrim(title) <> ''),
	constraint recurring_templates_frequency_allowed check (frequency in ('daily', 'weekly', 'monthly', 'yearly')),
	constraint recurring_templates_interval_positive check (interval > 0),
	constraint recurring_templates_basis_allowed check (recurrence_basis in ('completion_date', 'scheduled_date', 'deadline_date')),
	constraint recurring_templates_end_type_allowed check (end_type in ('never', 'after_count', 'on_date')),
	constraint recurring_templates_status_allowed check (status in ('active', 'on_hold', 'abandoned')),
	constraint recurring_templates_hidden_reason_allowed check (
		hidden_reason is null or hidden_reason in ('converted_to_recurring_template', 'recurring_skipped')
	),
	constraint recurring_templates_completed_count_non_negative check (completed_count >= 0),
	constraint recurring_templates_next_sequence_positive check (next_sequence >= 1),
	constraint recurring_templates_end_shape check (
		(end_type = 'never' and end_date is null and end_after_count is null)
		or (end_type = 'on_date' and end_date is not null and end_after_count is null)
		or (end_type = 'after_count' and end_after_count is not null and end_after_count > 0 and end_date is null)
	)
);

create table if not exists items (
	id uuid primary key,
	user_id uuid not null,
	item_type text not null,
	title text not null,
	note text null,
	status text not null,
	area_id uuid null,
	scheduled_date date null,
	scheduled_time time null,
	planned_work_date date null,
	deadline_date date null,
	deadline_time time null,
	deadline_at timestamptz null,
	deadline_timezone text null,
	review_date date null,
	reminder_date date null,
	reminder_time time null,
	reminder_at timestamptz null,
	reminder_intent text null,
	scheduled_time_zone_mode text null,
	deadline_time_zone_mode text null,
	reminder_time_zone_mode text null,
	recurring_template_id uuid null,
	recurring_sequence integer null,
	recurring_anchor_date date null,
	generated_from_item_id uuid null,
	importance integer null,
	estimated_effort integer null,
	pressure_metadata jsonb not null default '{}'::jsonb,
	quick_add_source_text text null,
	quick_add_parse_result jsonb null,
	created_at timestamptz not null,
	updated_at timestamptz not null,
	deleted_at timestamptz null,
	archived_at timestamptz null,
	hidden_reason text null,
	client_updated_at timestamptz not null,
	server_updated_at timestamptz not null,
	created_by_device_id text not null,
	updated_by_device_id text not null,
	revision bigint not null,
	constraint items_user_id_id_unique unique (user_id, id),
	constraint items_area_user_fk foreign key (user_id, area_id) references areas(user_id, id),
	constraint items_recurring_template_user_fk foreign key (user_id, recurring_template_id) references recurring_task_templates(user_id, id),
	constraint items_generated_from_user_fk foreign key (user_id, generated_from_item_id) references items(user_id, id),
	constraint items_type_allowed check (item_type in ('inbox', 'date_task', 'deadline_task', 'idea')),
	constraint items_status_allowed check (status in ('active', 'completed', 'postponed', 'on_hold', 'abandoned')),
	constraint items_title_non_empty check (btrim(title) <> ''),
	constraint items_hidden_reason_allowed check (
		hidden_reason is null or hidden_reason in ('converted_to_recurring_template', 'recurring_skipped')
	),
	constraint items_scheduled_tz_mode_allowed check (
		scheduled_time_zone_mode is null or scheduled_time_zone_mode in ('floating')
	),
	constraint items_deadline_tz_mode_allowed check (
		deadline_time_zone_mode is null or deadline_time_zone_mode in ('date_only', 'floating', 'fixed')
	),
	constraint items_reminder_tz_mode_allowed check (
		reminder_time_zone_mode is null or reminder_time_zone_mode in ('floating', 'fixed')
	),
	constraint items_importance_range check (importance is null or importance between 1 and 5),
	constraint items_estimated_effort_range check (estimated_effort is null or estimated_effort between 1 and 5),
	constraint items_recurring_sequence_shape check (
		(recurring_template_id is null and recurring_sequence is null)
		or (recurring_template_id is not null and recurring_sequence is not null and recurring_sequence > 0)
	),
	constraint items_deadline_shape check (
		(deadline_date is null and deadline_time is null and deadline_at is null and deadline_time_zone_mode is null)
		or (deadline_time_zone_mode = 'date_only' and deadline_date is not null and deadline_time is null and deadline_at is null)
		or (deadline_time_zone_mode = 'floating' and deadline_date is not null and deadline_time is not null)
		or (deadline_time_zone_mode = 'fixed' and deadline_at is not null and deadline_date is null and deadline_time is null)
	)
);

create table if not exists operation_history (
	id uuid primary key,
	user_id uuid not null,
	item_id uuid null,
	recurring_template_id uuid null,
	event_type text not null,
	previous_value jsonb not null default '{}'::jsonb,
	new_value jsonb not null default '{}'::jsonb,
	reason text null,
	idempotency_key text null,
	created_at timestamptz not null,
	created_by_device_id text not null,
	constraint operation_history_item_user_fk foreign key (user_id, item_id) references items(user_id, id),
	constraint operation_history_template_user_fk foreign key (user_id, recurring_template_id) references recurring_task_templates(user_id, id),
	constraint operation_history_event_type_allowed check (
		event_type in (
			'completed',
			'postponed',
			'activated_from_postponed',
			'on_hold',
			'abandoned',
			'restored',
			'reopened',
			'deleted',
			'archived',
			'skipped',
			'generated_next_instance',
			'converted_to_recurring_template'
		)
	),
	constraint operation_history_idempotency_key_non_empty check (
		idempotency_key is null or btrim(idempotency_key) <> ''
	)
);

create or replace function reject_operation_history_mutation()
returns trigger
language plpgsql
as $$
begin
	raise exception 'operation_history is append-only';
end;
$$;

drop trigger if exists operation_history_reject_update on operation_history;
create trigger operation_history_reject_update
	before update on operation_history
	for each row execute function reject_operation_history_mutation();

drop trigger if exists operation_history_reject_delete on operation_history;
create trigger operation_history_reject_delete
	before delete on operation_history
	for each row execute function reject_operation_history_mutation();

create table if not exists user_settings (
	user_id uuid primary key,
	language text not null,
	locale text not null,
	week_start_day text not null,
	time_zone text not null,
	date_display_format text not null,
	time_display_format text not null,
	default_time_zone_mode text not null,
	today_primary_lookahead_days integer not null default 3,
	deadline_awareness_days integer not null default 14,
	weather_city text not null default 'Tokyo',
	created_at timestamptz not null,
	updated_at timestamptz not null,
	client_updated_at timestamptz not null,
	server_updated_at timestamptz not null,
	created_by_device_id text not null,
	updated_by_device_id text not null,
	revision bigint not null,
	constraint user_settings_language_allowed check (language in ('en', 'zh-Hans', 'ja')),
	constraint user_settings_locale_non_empty check (btrim(locale) <> ''),
	constraint user_settings_week_start_allowed check (week_start_day in ('sunday', 'monday')),
	constraint user_settings_time_zone_non_empty check (btrim(time_zone) <> ''),
	constraint user_settings_date_display_format_non_empty check (btrim(date_display_format) <> ''),
	constraint user_settings_time_display_format_allowed check (time_display_format in ('12h', '24h')),
	constraint user_settings_default_tz_mode_allowed check (default_time_zone_mode in ('floating')),
	constraint user_settings_today_lookahead_positive check (today_primary_lookahead_days > 0),
	constraint user_settings_deadline_awareness_positive check (deadline_awareness_days > 0),
	constraint user_settings_weather_city_non_empty check (btrim(weather_city) <> '')
);

create unique index if not exists areas_unique_active_name
	on areas (user_id, lower(name))
	where deleted_at is null;
create index if not exists areas_user_sort_order_idx on areas (user_id, sort_order);
create index if not exists areas_user_visibility_idx on areas (user_id, archived_at, deleted_at);

create index if not exists recurring_templates_user_status_visibility_idx
	on recurring_task_templates (user_id, status, deleted_at, archived_at);
create index if not exists recurring_templates_user_start_date_idx
	on recurring_task_templates (user_id, start_date);

create unique index if not exists items_unique_recurring_instance
	on items (user_id, recurring_template_id, recurring_sequence)
	where recurring_template_id is not null;
create index if not exists items_user_status_visibility_idx on items (user_id, status, deleted_at, archived_at);
create index if not exists items_user_type_scheduled_idx on items (user_id, item_type, scheduled_date);
create index if not exists items_user_planned_work_date_idx on items (user_id, planned_work_date);
create index if not exists items_user_deadline_date_idx on items (user_id, deadline_date);
create index if not exists items_user_deadline_local_idx on items (user_id, deadline_date, deadline_time);
create index if not exists items_user_deadline_at_idx on items (user_id, deadline_at);
create index if not exists items_user_review_date_idx on items (user_id, review_date);
create index if not exists items_user_area_idx on items (user_id, area_id);
create index if not exists items_user_updated_at_idx on items (user_id, updated_at);

create unique index if not exists operation_history_unique_idempotency_key
	on operation_history (user_id, idempotency_key)
	where idempotency_key is not null;
create index if not exists operation_history_user_item_created_idx on operation_history (user_id, item_id, created_at);
create index if not exists operation_history_user_template_created_idx on operation_history (user_id, recurring_template_id, created_at);
create index if not exists operation_history_user_event_created_idx on operation_history (user_id, event_type, created_at);
