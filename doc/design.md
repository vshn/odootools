# Design

https://discuss.vshn.net/t/automate-over-undertime-calculation/399/6


## FAQs

### Why not an SPA and offload all work into the user's browser?

CORS. Browsers won't connect to Odoo from somewhere else (which is fine).



# Backlog

## Caching

The Odoo client will very likely need some caching. The read queries are quite hefy for Odoo, and currently all data is requested on each page load.

A cache key could be generated using the following things as an input:

* Request URL
* Request Body
* Cookies (or more specific, the Odoo Session ID)

Just do sha256(url+body+SID) to get a cache key.

A very simple implementation could use Redis and TTL's on keys to have a self-cleaning cache.
