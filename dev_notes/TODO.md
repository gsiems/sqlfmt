
## Issues

* In the psql test input the back-slash characters are being dropped (appears
that they are being treated as escape chars somewhere) so tokens that should be
"\set" and "\unset" end up as "et" and "nset"

## Dialects

* ~~consider how to identify the size/precision/scale of a datatype as being part
of the datatype~~

## Parser

* ~~consider how to flag the size/precision/scale of a datatype as being part of
the datatype~~ If needed, this should be moved to post-parsing.

* The parser currently strips any final trailing whitespace from the input.
While this doesn't break the formatter, it should be fixed (if only to make it
easier to test/verify the parser).

## Tagging

Still need to tag:

* Need to tag PostgreSQL COPY commands before anything else (or at least the
plain-text data portion thereof)

* Oracle package, function, and procedure code
* DDL code

## Formatting

* Since this is intended to format code the way that "I" prefer to see/use it,
then there may be some benefit to including a style guide that documents and
explains my preferences.
