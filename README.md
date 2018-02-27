ip2country is a service that returns country name detected from IP address of
a visitor.

ip2country obtains country data from one of the following lookup services:

* http://freegeoip.net/
* http://geoip.nekudo.com/

List of those services could be expanded. Data received from those lookup
services should be cached in local database. Duration of caching should be
configurable.

Lookup services have their own rate limits which should be honored by this
service. If one service is above rate limit other service must be selected
for lookups.

Settings of lookup services and their list should be stored in configuration
file.

There were made several assumptions about the task:

* services use GET requests;
* IP address is placed directly within URL;
* services return JSON.

This service also returns json, with two keys, IP and Country.
Two backends for caching currently available, in-memory cache and PostgreSQL.
