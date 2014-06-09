Mapblue
=======

Mapblue maps likely political persuasions using data from the US census and
results from the [General Social Survey](http://www3.norc.org/gss+website/)
(GSS).

Technology
==========

Mapblue's backend and census data loading program is written in
[Go](http://golang.org).  The frontend is written in JavaScript and employs the
Google Maps API.  We use PostgreSQL for its excellence and for
[PostGIS](http://postgis.net).

We used [Stata](http://www.stata.com) to perform the regression analysis on
various demographic characteristics.

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

The decennial census holds data down to the block level, which is an extreme
level of granularity.  However, the tradeoff is that the census asks only the
most cursory of demographic questions, putting a ceiling on map accuracy.

Mapblue is currently a proof-of-concept, and we've therefore restricted the
usable map to Indiana (our home state).  Other than a lack of resources
(servers with large amounts of fast storage aren't cheap), nothing prevents the
other states from being added other than a few assumptions made in the census
data loading program -- which can easily be fixed.

Possibilities
=============

Professional campaign software pulls data from myriad sources; Mapblue could do
this as well, but only to a limited extent (we don't have the resources to
canvass or purchase large banks of information).

Authors
=======

Charlie & Sam Gunyon wrote this.

