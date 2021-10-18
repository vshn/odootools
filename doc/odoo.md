# Odoo API

References:

* https://github.com/odoo/odoo/blob/8.0/openerp/http.py


## Login

POST /web/session/authenticate

```json
{
    "id": "1337",
    "jsonrpc": "2.0",
    "method": "call",
    "params": {
        "db": "<DB>",
        "login": "<username>",
        "password": "<password>"
    }
}
```

### Success

```json
{
    "id": "1337",
    "jsonrpc": "2.0",
    "result": {
        "company_id": 1,
        "db": "<DB>",
        "session_id": "xxx",
        "uid": 42,
        "user_context": {
            "lang": "en_US",
            "tz": "Europe/Zurich",
            "uid": 42
        },
        "username": "xxx"
    }
}
```

### Failure

```json
{
    "id": "1337",
    "jsonrpc": "2.0",
    "result": {
        "company_id": null,
        "db": "xxx",
        "session_id": "xxx",
        "uid": false,
        "user_context": {},
        "username": "xxx"
    }
}
```


## Read Attendances

POST /web/dataset/search_read

Cookie: session_id=xxx

```json
{
    "id": "1337",
    "jsonrpc": "2.0",
    "method": "call",
    "params": {
        "domain": [
            [
                "employee_id.user_id.id",
                "=",
                42
            ]
        ],
        "fields": [
            "employee_id",
            "name",
            "action",
            "action_desc"
        ],
        "limit": 1,
        "model": "hr.attendance",
        "offset": 0
    }
}
```

```json
{
    "id": "1337",
    "jsonrpc": "2.0",
    "result": {
        "length": 4901,
        "records": [
            {
                "action": "sign_out",
                "action_desc": false,
                "employee_id": [
                    1337,
                    "John Doe"
                ],
                "id": 151253,
                "name": "2021-10-15 16:39:06"
            }
        ]
    }
}
```
