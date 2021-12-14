[![Build](https://img.shields.io/github/workflow/status/vshn/odootools/Test)][build]
![Go version](https://img.shields.io/github/go-mod/go-version/vshn/odootools)
[![Version](https://img.shields.io/github/v/release/vshn/odootools)][releases]
[![Maintainability](https://img.shields.io/codeclimate/maintainability/vshn/odootools)][codeclimate]
[![GitHub downloads](https://img.shields.io/github/downloads/vshn/odootools/total)][releases]
[![License](https://img.shields.io/github/license/vshn/odootools)][license]

[build]: https://github.com/vshn/odootools/actions?query=workflow%3ATest
[releases]: https://github.com/vshn/odootools/releases
[license]: https://github.com/vshn/odootools/blob/master/LICENSE
[codeclimate]: https://codeclimate.com/github/vshn/odootools

# odootools

odootools is a small tool that allows you to calculate overtime based on your attendances.
It has VSHN-specific business rules integrated that are otherwise calculated manually.

Simply login with your Odoo credentials, configure the report settings and generate your reports.

It's currently aimed at Odoo 8.

### Run the tool

First, you need to export Odoo settings:
```bash
export ODOO_URL=https://...
export ODOO_DB=...
```

You can run the tool in different ways:

1. using `make run` (uses `go run`).
2. using `make run.docker` (uses `docker run`)
3. using a configuration of your favorite IDE

### Deploy to OpenShift

Setup the project and deploy user

```bash
ns=vshn-odoo-prod
nstest=vshn-odoo-test
sa=odootools-deployer

oc new-project $ns
oc new-project $nstest

oc -n $ns create sa $sa

# Allow the deployer user to manage deployments in test namespace
oc -n $ns     policy add-role-to-user admin -z $sa --rolebinding-name admin
oc -n $nstest policy add-role-to-user admin -z $sa --rolebinding-name admin
oc -n $nstest policy add-role-to-user system:image-pusher -z $sa
oc -n $nstest patch rolebinding admin --type='json' -p='[{"op": "replace", "path": "/subjects/1/namespace", "value":"'$ns'"}]'
oc -n $nstest patch rolebinding system:image-pusher --type='json' -p='[{"op": "replace", "path": "/subjects/0/namespace", "value":"'$ns'"}]'

# Get SA token
oc -n $ns sa get-token $sa
```
