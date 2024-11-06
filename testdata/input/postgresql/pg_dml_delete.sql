-- sqlfmt dialect: PostgreSQL

/*
References:
https://www.postgresql.org/docs/17/sql-delete.html
https://www.postgresql.org/docs/17/sql-truncate.html
*/


DELETE FROM films USING producers
  WHERE producer_id = producers.id AND producers.name = 'foo';


DELETE FROM films
  WHERE producer_id IN (SELECT id FROM producers WHERE name = 'foo');


DELETE FROM films WHERE kind <> 'Musical';


DELETE FROM films;


DELETE FROM tasks WHERE status = 'DONE' RETURNING *;


DELETE FROM tasks WHERE CURRENT OF c_tasks;


WITH delete_batch AS (
  SELECT l.ctid FROM user_logs AS l
    WHERE l.status = 'archived'
    ORDER BY l.creation_date
    FOR UPDATE
    LIMIT 10000
)
DELETE FROM user_logs AS dl
  USING delete_batch AS del
  WHERE dl.ctid = del.ctid;


truncate table foo;

TRUNCATE bigtable, fattable;


TRUNCATE bigtable, fattable RESTART IDENTITY;


TRUNCATE othertable CASCADE;
