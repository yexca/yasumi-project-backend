-- PowerSync logical replication publication for MVP synced tables.
-- Account and credential tables are intentionally excluded from the sync stream.

do $$
begin
	if not exists (select 1 from pg_publication where pubname = 'powersync') then
		create publication powersync for table
			areas,
			recurring_task_templates,
			items,
			operation_history,
			user_settings;
	end if;
end
$$;
