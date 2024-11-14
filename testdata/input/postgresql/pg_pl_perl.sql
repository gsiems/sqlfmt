-- sqlfmt d:postgres

/*
References:
https://www.postgresql.org/docs/current/plperl-funcs.html
*/

CREATE OR REPLACE FUNCTION funcname (a_id integer)
RETURNS text
-- function attributes can go here
AS $$
    # PL/Perl function body goes here
$$ LANGUAGE plperl;

DO $$
    # PL/Perl code
$$ LANGUAGE plperl;

CREATE FUNCTION perl_max (integer, integer) RETURNS integer AS $$
    if ($_[0] > $_[1]) { return $_[0]; }
    return $_[1];
$$ LANGUAGE plperl;

CREATE FUNCTION perl_max (integer, integer) RETURNS integer AS $$
    my ($x, $y) = @_;
    if (not defined $x) {
        return undef if not defined $y;
        return $y;
    }
    return $x if not defined $y;
    return $x if $x > $y;
    return $y;
$$ LANGUAGE plperl;

CREATE FUNCTION perl_and(bool, bool) RETURNS bool
TRANSFORM FOR TYPE bool
LANGUAGE plperl
AS $$
  my ($a, $b) = @_;
  return $a && $b;
$$ ;
