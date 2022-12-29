# security-falco-cserver
This project serves up falco rule sets over an HTTP(S) web service. It currently has 3 endpoints:

* `/`
* `/ruleset`
* `/sum`

This service concatenates rule sets based on the ones requested. 

# Authentication
User config is defined in a file named: `/secrets/auth.yml`. This file should be mounted by GCP secrets, a docker volume, or similar method. A sample user config file is in `test/auth.yml`. It is a yaml file that has a map of `token`->`username`

Requests should be authenticated by putting the user token in an `X-Auth-Token` header on the request. 

Requests may also use basic authentication.

The health check (`/`) is always allowed without authentication

# Endpoints

## /
This endpoint returns a 200 and is a basic health check

## /ruleset
This API accepts `GET` requests in the form:

```shell
curl -H "X-Auth-Token: some_token_1" localhost:8080/ruleset?rulesets=test,other_test
```

`rulesets` is a comma separated list of rule sets that will be concatenated with the default rule set (`_default.yml`) and be returned to the user **in the order requested**. In this example the following rule sets will be concatenated: `_default.yml`+`test.yml`+`other_test.yml`.

**note: in the request, do NOT include the yml or yaml extension**. In the above sample we use `test` for `test.yml` in the `rules` directory.

## /sum
This API accepts `GET` requests in the form:

```shell
curl -H "X-Auth-Token: some_token_1" localhost:8080/sum?ruletsets=puppet
```

It returns the `sha256` hash of the file generated by the equivilent `/ruleset` call.

Note: this API also supports basic auth as outlined in the `Authentication` section of this README

# CI
The github action defined in `.github\workflows\check-falco.yml` will check all permutations of the rules files defined in the `rules` directory and ensure that they parse correctly by using the following command on each permutation:

```
docker run -v <some path>:/etc/falco/falco_rules.yaml -e SKIP_DRIVER_LOADER=1 falcosecurity/falco /usr/bin/falco -V /etc/falco/falco_rules.yaml
 ```