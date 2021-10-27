# Design

https://discuss.vshn.net/t/automate-over-undertime-calculation/399/6

## Holidays

Holidays generally don't influence overtime.
Odoo just calculates the monthly work hours, minus weekends and leaves, multiplied FTE ratio.

In case of VSHN, we have special leave type "Unpaid" - which Odoo treats as normal leaves, but VSHN treats as undertime.

VSHN only allows "full leave days" in respect to FTE ratio.
So, one cannot consume 0.5 leave days, only 1.
In the end it doesn't matter:
- Consuming 1 full day but still work on this day results in overtime, but reduces holidays
- Not using a leave day, but only work half day results in undertime, but still having an additional holiday.
At the end of the year, excess holidays are transformed into overtime.

## FAQs

### Why not an SPA and offload all work into the user's browser?

CORS. Browsers won't connect to Odoo from somewhere else (which is fine).

### Get overtime from previous month?

Payslips can be queried and updated - just not with the normal VSHNeer access levels.
Easiest solution would be to enable at least read access so that odootools can accumulate the delta.

### Get FTE ratio instead of manually entering it?

FTE ratio can also be queried, but not with current VSHNeer access levels.
The contracts and the ratio are available as fields - one just needs access to it.

# Backlog

## Caching

The Odoo client will very likely need some caching. The read queries are quite hefy for Odoo, and currently all data is requested on each page load.

A cache key could be generated using the following things as an input:

* Request URL
* Request Body
* Cookies (or more specific, the Odoo Session ID)

Just do sha256(url+body+SID) to get a cache key.

A very simple implementation could use Redis and TTL's on keys to have a self-cleaning cache.
