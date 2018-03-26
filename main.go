package main

import (
	"fmt"
	"regexp"

	"github.com/clbanning/mxj"
	xp "github.com/vjeantet/bitfan/commons/xprocessor"
)

var envs map[string]string
var r *xp.Runner

func main() {
	r = xp.New(Configure, Start, Receive, Stop)
	r.Description = "Produce or enrich events from issues found in JIRA"
	r.ShortDescription = ""

	r.OptionString("url", true, "http(s) jira rest endpoint", "")
	r.OptionString("username", true, "username", "")
	r.OptionString("password", true, "password", "")

	r.OptionMapString("count", false, "return count for each provided jql or filterID or Key", nil)
	r.OptionStringSlice("issues", false, "return all issue for each jql or filterID or Key", nil)

	r.OptionInt("max_result", false, "maximum issue to return (usable only with `issues` and `keys` fields)", 0)
	// r.OptionInt("min_result", false, "do not generate event when jira returns less result than this number", 0)

	r.OptionStringSlice("fields", false, "search results returns basic fields. List here additional fields to retreive", nil)
	r.OptionString("event_by", false, "`issue` => produce one event for each found issue, or `result` for one event with all resulting issues", "result")

	r.Run(1)
}

func Configure() error {
	if len(r.Opt.MapString("count")) == 0 && len(r.Opt.StringSlice("issues")) == 0 {
		return fmt.Errorf("missing `count` or `issues` param")
	}

	return nil
}

func Start() error {
	return nil
}

const (
	FILTER_ID = iota + 1
	JQL
	KEY
)

var reKey *regexp.Regexp
var reFilterId *regexp.Regexp

func jiraRequestKind(request string) int {
	if reKey == nil {
		reKey = regexp.MustCompile("^[a-zA-Z]+-[0-9]+$")
		reFilterId = regexp.MustCompile("^[0-9]+$")
	}
	if reKey.MatchString(request) {
		return KEY
	}
	if reFilterId.MatchString(request) {
		return FILTER_ID
	}
	return JQL
}

func Receive(data interface{}) error {
	message := mxj.Map(data.(map[string]interface{}))

	// Connect to JIRA and acquire Auth.
	r.Debugf("connecting")
	j, err := newJiraClient(
		r.Opt.String("url"),
		r.Opt.String("username"),
		r.Opt.String("password"),
		r.Logf,
	)
	if err != nil {
		return err
	}

	// Process count
	countMap := map[string]int{}
	for fname, requestString := range r.Opt.MapString("count") {
		var count int
		var err error

		switch jiraRequestKind(requestString) {
		case KEY:
			// CountIssuesByKey
			r.Debugf("%s => CountIssuesByKey(`%s`)", fname, requestString)
			count, err = j.CountIssuesByKey(requestString)
		// case FILTER_ID:
		// 	// CountIssuesByFilterID
		// 	r.Debugf("%s => CountIssuesByFilterID(`%s`)", fname, requestString)
		// 	count, err = j.CountIssuesByFilterID(requestString)
		case JQL:
			// CountIssuesByJQL
			r.Debugf("%s => CountIssuesByJQL(`%s`)", fname, requestString)
			count, err = j.CountIssuesByJQL(requestString)
		}

		if err != nil {
			r.Logf("%s", err)
			continue
		}

		countMap[fname] = count
	}
	if len(countMap) > 0 {
		message.SetValueForPath(countMap, "counts")
	}

	// Process issues
	issues := []map[string]interface{}{}
	for _, requestString := range r.Opt.StringSlice("issues") {
		var err error
		var issuesChan chan map[string]interface{}

		switch jiraRequestKind(requestString) {
		case KEY:
			// FindOneIssueByKey
			r.Debugf("FindOneIssueByKey(`%s`,`%d`,`%s`)", requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
			issuesChan, err = j.FindOneIssueByKey(requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
		// case FILTER_ID:
		// 	r.Debugf("FindIssuesByFilterID(`%s`,`%d`,`%s`)", requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
		// 	issuesChan, err = j.FindIssuesByFilterID(requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
		case JQL:
			// FindIssuesByJQL
			r.Debugf("FindIssuesByJQL(`%s`,`%d`,`%s`)", requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
			issuesChan, err = j.FindIssuesByJQL(requestString, r.Opt.Int("max_result"), r.Opt.StringSlice("fields"))
		}

		if err != nil {
			r.Logf(err.Error())
			continue
		}

		for issue := range issuesChan {
			if r.Opt.String("event_by") == "issue" {
				r.Send(issue)
			} else {
				issues = append(issues, issue)
			}
		}
	}

	if r.Opt.String("event_by") != "issue" && len(issues) > 0 {
		message.SetValueForPath(issues, "issues")
	}

	if message.Exists("issues") || message.Exists("counts") {
		return r.Send(message)
	}

	return nil

}

func Stop() error {
	return nil
}
