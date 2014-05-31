Mapblue
=======

Mapblue intends to map likely Democrats, using data from the US census and
results from the [General Social Survey](http://www3.norc.org/gss+website/)
(GSS).  The decennial census holds data down to the block level, which can be
immensely useful for GOTV and communications efforts in campaigns with limited
resources.  However, the tradeoff is that the census asks only the most cursory
of demographic questions, putting a ceiling on map accuracy.

Technology
==========

Mapblue's backend and census data loading program is written in
[Go](http://golang.org).  The frontend is written in
[Dart](http://dartlang.org) and employs the Google Maps API.  We use PostgreSQL
for its excellence and for [PostGIS](http://postgis.net).

**WARNING!!!**

TL;DR: Use Go 1.3, even the betas.

There is a bug in Go 1.2 where `database/sql` ignores calls to
`SetMaxOpenConns`.  `database/sql` uses connection polling such that every
query uses a different connection (if possible), and the limit is set by
`SetMaxOpenConns`.  But, if those calls are ignored, `load_census_data` will
quickly butt up against PostgreSQL's configured connection limit and the
program will fail.

Limitations
===========

Mapblue is currently a proof-of-concept, and we've therefore restricted the
usable map to Indiana (my home state).  Other than a lack of resources (servers
with large amounts of fast storage aren't cheap), nothing prevents the other
states from being added other than a few assumptions made in the census data
loading program -- which can easily be fixed.

Authors
=======

Charlie & Sam Gunyon wrote this.

