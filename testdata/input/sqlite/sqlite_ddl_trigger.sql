-- sqlfmt d:sqlite

/*
References:
https://www.sqlite.org/lang_createtrigger.html
https://www.sqlite.org/lang_droptrigger.html
*/

CREATE TRIGGER update_customer_address AFTER UPDATE OF address ON customers
  BEGIN
    UPDATE orders SET address = new.address WHERE customer_name = old.name;
  END;

CREATE TABLE customer(
  cust_id INTEGER PRIMARY KEY,
  cust_name TEXT,
  cust_addr TEXT
);

CREATE VIEW customer_address AS
   SELECT cust_id, cust_addr FROM customer;

CREATE TRIGGER cust_addr_chng
INSTEAD OF UPDATE OF cust_addr ON customer_address
BEGIN
  UPDATE customer SET cust_addr=NEW.cust_addr
   WHERE cust_id=NEW.cust_id;
END;

DROP TRIGGER IF EXISTS cust_addr_chng ;

/*

CREATE [ TEMP | TEMPORARY ] VIEW [ IF NOT EXISTS ] trigger_name
    { BEFORE | AFTER | INSTEAD OF } { DELETE | INSERT | UPDATE [ OF column_name [, ...] } ON table_name
    [ FOR EACH ROW | WHEN expression ]
BEGIN
    statements;
END ;

*/
