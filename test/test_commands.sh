# should return _default + test.yml
curl -v -XPOST -H "X-Auth-Token: some_token_1" -d '{"rulesets":["test"]}' localhost:8080/ruleset
# should return just _default
curl -v -XPOST -H "X-Auth-Token: some_token_1" -d '{"rulesets":["default"]}' localhost:8080/ruleset
# should return just _default
curl -v -XPOST -H "X-Auth-Token: some_token_1" -d '{"rulesets":[]}' localhost:8080/ruleset
# should return just _default but log the missing rulset
curl -v -XPOST -H "X-Auth-Token: some_token_1" -d '{"rulesets":["foo"]}' localhost:8080/ruleset