-- sqlfmt d:sqlite

/*
References:
https://www.sqlite.org/lang_createtable.html
https://www.sqlite.org/lang_altertable.html
https://www.sqlite.org/lang_droptable.html
https://www.sqlite.org/lang_createvtab.html
*/

CREATE TABLE t1(x INT CHECK( x>3 ));
/* Insert a row with X less than 3 by directly writing into the
** database file using an external program */
PRAGMA integrity_check;  -- Reports row with x less than 3 as corrupt
INSERT INTO t1(x) VALUES(2);  -- Fails with SQLITE_CORRUPT
SELECT x FROM t1;  -- Returns an integer less than 3 in spite of the CHECK constraint


    CREATE TABLE t(x INTEGER PRIMARY KEY ASC, y, z);
    CREATE TABLE t(x INTEGER, y, z, PRIMARY KEY(x ASC));
    CREATE TABLE t(x INTEGER, y, z, PRIMARY KEY(x DESC));

CREATE TABLE parent(a PRIMARY KEY, b UNIQUE, c, d, e, f);
CREATE UNIQUE INDEX i1 ON parent(c, d);
CREATE INDEX i2 ON parent(e);
CREATE UNIQUE INDEX i3 ON parent(f COLLATE nocase);

CREATE TABLE child1(f, g REFERENCES parent(a));                        -- Ok
CREATE TABLE child2(h, i REFERENCES parent(b));                        -- Ok
CREATE TABLE child3(j, k, FOREIGN KEY(j, k) REFERENCES parent(c, d));  -- Ok

CREATE TABLE artist(
  artistid    INTEGER PRIMARY KEY,
  artistname  TEXT
);
CREATE TABLE track(
  trackid     INTEGER,
  trackname   TEXT,
  trackartist INTEGER REFERENCES artist
);
CREATE INDEX trackindex ON track(trackartist);


CREATE TABLE album(
  albumartist TEXT,
  albumname TEXT,
  albumcover BINARY,
  PRIMARY KEY(albumartist, albumname)
);

CREATE TABLE song(
  songid     INTEGER,
  songartist TEXT,
  songalbum TEXT,
  songname   TEXT,
  FOREIGN KEY(songartist, songalbum) REFERENCES album(albumartist, albumname)
);

CREATE TABLE artist(
  artistid    INTEGER PRIMARY KEY,
  artistname  TEXT
);

CREATE TABLE track(
  trackid     INTEGER,
  trackname   TEXT,
  trackartist INTEGER REFERENCES artist(artistid) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE track(
  trackid     INTEGER,
  trackname   TEXT,
  trackartist INTEGER DEFAULT 0 REFERENCES artist(artistid) ON DELETE SET DEFAULT
);

CREATE TABLE parent(x PRIMARY KEY);
CREATE TABLE child(y REFERENCES parent ON UPDATE SET NULL);



/*

CREATE [ TEMP | TEMPORARY ] VIEW [ IF NOT EXISTS ] table_name (
  ...
) ;

CREATE [ TEMP | TEMPORARY ] VIEW [ IF NOT EXISTS ] table_name
    AS
    select-statement ;

*/
