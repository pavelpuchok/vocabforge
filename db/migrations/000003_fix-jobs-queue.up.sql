drop index by_group;
alter table jobs_queue drop column is_removed;

create index by_group on jobs_queue(group_name, created_at);
