package main

import (
	"fmt"

	xp "github.com/vjeantet/bitfan/commons/xprocessor"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

var envs map[string]string
var r *xp.Runner

var url, username, password, jql string

func main() {
	r = xp.New(
		Configure, Start, Receive, Stop,
	)
	r.OptionString("url", true, "", "")
	r.OptionString("username", true, "", "")
	r.OptionString("password", true, "", "")
	r.OptionString("jql", true, "", "")

	r.Run(1)
}

func Configure(options xp.Options) error {
	url = options.String("url")
	username = options.String("username")
	password = options.String("password")
	jql = options.String("jql")
	return nil
}

func Start() error {
	return nil
}

func Receive(data interface{}) error {
	num, err := getNumberOfIssues(jql, url, username, password)
	if err != nil {
		r.Logf("%s\n", err.Error())
		return err
	}

	return r.Send(map[string]interface{}{
		"count": num,
	})
}

func getNumberOfIssues(jql, jiraUrl, username, password string) (int, error) {
	jiraClient, err := jira.NewClient(nil, jiraUrl)
	if err != nil {
		return 0, fmt.Errorf("JiraJob : error jira connect : %s", err)
	}

	if username != "" && jiraClient.Authentication.Authenticated() == false {
		res, err := jiraClient.Authentication.AcquireSessionCookie(username, password)
		if err != nil || res == false {
			return 0, fmt.Errorf("Authentification error : %s", err)
		}
	}

	options := jira.SearchOptions{
		StartAt:    0,
		MaxResults: 1,
	}

	_, body, err := jiraClient.Issue.Search(jql, &options)
	if err != nil {
		return 0, fmt.Errorf("JiraJob : search error : %s, %s", err, jql)
	}

	return body.Total, nil

}

func Stop() error {
	return nil
}
