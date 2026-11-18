SELECT json_build_object(
  'timestamp', clock_timestamp(),
  'statements', (SELECT json_agg(s) FROM (SELECT queryid::text, calls, total_exec_time, rows, shared_blks_hit, shared_blks_read FROM pg_stat_statements WHERE query LIKE 'SELECT count(*) FROM orders%') s),
  'tables', (SELECT json_agg(s) FROM (SELECT relname, seq_scan, idx_scan, n_live_tup FROM pg_stat_user_tables WHERE relname='orders') s),
  'activity', (SELECT json_agg(s) FROM (SELECT state, wait_event_type, count(*) FROM pg_stat_activity WHERE datname=current_database() GROUP BY state,wait_event_type) s),
  'stats_reset', (SELECT stats_reset FROM pg_stat_statements_info)
);
