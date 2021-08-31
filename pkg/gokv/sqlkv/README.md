```sql
create table kv
(
    k       varchar(100) not null primary key,
    v       text,
    state   tinyint      not null default 1,
    updated datetime,
    created datetime     not null
)
```
