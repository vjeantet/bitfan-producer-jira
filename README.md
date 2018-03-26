# bitfan-producer-jira
jira producer xprocessor

## Params

### URL
### Username
### Password

### Count
* mapString
key => JQL
key => JQL

### Issues
* StringSlice
["JQL","JQL","JQL","JQL","ISSUE-KEY"]

### MaxResult
* int

### Fields
* stringSlice

### Event_by
* string
"result" or "issues"
default "result"

# Behavior

## retreive multiple jql count
* produce one event with named fields with count key

## jira search
* retreive basic fields by default
* allow to retreive more fields
* produce one event with all results, or one event by issue event_by like SQL
* allow multiple jql

## allow jira issue by KEY
* instead of jql, use key

## allow usage in filter
* enrich received event with issuescount or issues fields
