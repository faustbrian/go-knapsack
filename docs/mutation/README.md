# Mutation evidence

`golib mutation --module .` and
`golib mutation --module objective/gomoney` run the canonical content-addressed
campaigns. Each production package receives its own
content-addressed campaign and must achieve exact 100% efficacy and mutant
coverage. Every viable mutant must be killed; uncovered, timed-out, live,
malformed, missing, or unclassified results fail closed.

The tracked `raw/` reports and `specification/mutation-classifications.tsv`
describe the superseded standalone campaign. They remain historical audit
material and are not exclusions or release evidence under the current policy.
