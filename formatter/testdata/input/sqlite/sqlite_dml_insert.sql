-- sqlfmt d:sqlite

/*
References:
https://www.sqlite.org/lang_insert.html
https://www.sqlite.org/lang_upsert.html
*/


INSERT INTO phonebook2(name,phonenumber,validDate)
  VALUES('Alice','704-555-1212','2018-05-08')
  ON CONFLICT(name) DO UPDATE SET
    phonenumber=excluded.phonenumber,
    validDate=excluded.validDate
  WHERE excluded.validDate>phonebook2.validDate;

INSERT INTO t1 SELECT * FROM t2 WHERE true
ON CONFLICT(x) DO UPDATE SET y=excluded.y;
