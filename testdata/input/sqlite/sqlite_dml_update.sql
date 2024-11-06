-- sqlfmt d:sqlite

/*
References:
https://www.sqlite.org/lang_update.html
https://www.sqlite.org/lang_upsert.html
*/

UPDATE inventory
   SET quantity = quantity - daily.amt
  FROM (SELECT sum(quantity) AS amt, itemId FROM sales GROUP BY 2) AS daily
 WHERE inventory.itemId = daily.itemId;
