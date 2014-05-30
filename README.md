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

Limitations
===========

Mapblue is currently a proof-of-concept, and I've therefore restricted the
usable map to Indiana (my home state).  Other than a lack of resources (servers
with large amounts of fast storage aren't cheap), nothing prevents the other
states from being added other than a few assumptions made in the census data
loading program -- which can easily be fixed.

Authors
=======

Charlie & Sam Gunyon wrote this.

