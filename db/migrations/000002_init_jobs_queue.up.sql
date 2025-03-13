create table if not exists jobs_queue (
    id integer primary key,
    is_removed boolean not null,
    group_name text not null,
    item text not null,
    created_at datetime not null
);

create index by_group on jobs_queue(is_removed, group_name, created_at);

