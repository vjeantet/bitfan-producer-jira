package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/structs"
	jira "gopkg.in/andygrunwald/go-jira.v1"
)

const defaultFields string = "created,updated,duedate,resolutiondate,summary,description,issuetype,reporter,status"

type jiraClient struct {
	c    *jira.Client
	Logf func(string, ...interface{})
}

func newJiraClient(url, username, password string, log func(string, ...interface{})) (*jiraClient, error) {
	var err error
	j := &jiraClient{}
	j.Logf = log

	j.c, err = jira.NewClient(nil, url)
	if err != nil {
		return nil, fmt.Errorf("JiraJob : error jira connect : %s", err)
	}

	if username != "" && j.c.Authentication.Authenticated() == false {
		res, err := j.c.Authentication.AcquireSessionCookie(username, password)
		if err != nil || res == false {
			return nil, fmt.Errorf("Authentification error : %s", err)
		}
	}

	return j, nil
}

func (j *jiraClient) CountIssuesByKey(issueID string) (int, error) {
	fmt.Println(issueID)
	_, _, err := j.c.Issue.Get(issueID, nil)

	if err != nil {
		j.Logf("CountIssuesByKey : %s, %s", err, issueID)
		return 0, nil
	}

	return 1, nil
}

func (j *jiraClient) CountIssuesByJQL(jql string) (int, error) {
	options := jira.SearchOptions{
		StartAt:    0,
		MaxResults: 1,
		Fields:     []string{"summary"},
	}
	_, body, err := j.c.Issue.Search(jql, &options)
	if err != nil {
		return 0, fmt.Errorf("CountIssuesByJQL : %s, %s", err, jql)
	}

	return body.Total, nil
}

func (j *jiraClient) FindOneIssueByKey(issueID string, max int, cfields []string) (chan map[string]interface{}, error) {
	issue, _, err := j.c.Issue.Get(issueID, &jira.GetQueryOptions{
		Fields: defaultFields,
	})

	if err != nil {
		return nil, err
	}

	iChan := make(chan map[string]interface{}, 1)
	iChan <- issueToMSI(issue)
	close(iChan)

	return iChan, nil
}

func (j *jiraClient) FindIssuesByJQL(jql string, max int, cfields []string) (chan map[string]interface{}, error) {
	dfields := append(strings.Split(defaultFields, ","), cfields...)
	options := jira.SearchOptions{
		StartAt:    0,
		MaxResults: max,
		Fields:     dfields,
	}

	iChan := make(chan map[string]interface{}, max)

	go func() {
		defer close(iChan)
		j.Logf("JQL=%s", jql)
		err := j.c.Issue.SearchPages(jql, &options, func(issue jira.Issue) error {
			iChan <- issueToMSI(&issue)
			return nil
		})
		if err != nil {
			j.Logf("%s", err.Error())
		}
	}()

	return iChan, nil
}

func issueToMSI(i *jira.Issue) map[string]interface{} {
	msi := structs.Map(i.Fields)
	if !time.Time(i.Fields.Created).IsZero() {
		msi["created"] = time.Time(i.Fields.Created)
	}
	if !time.Time(i.Fields.Updated).IsZero() {
		msi["updated"] = time.Time(i.Fields.Updated)
	}
	if !time.Time(i.Fields.Duedate).IsZero() {
		msi["duedate"] = time.Time(i.Fields.Duedate)
	}
	if !time.Time(i.Fields.Resolutiondate).IsZero() {
		msi["resolutiondate"] = time.Time(i.Fields.Resolutiondate)
	}
	return msi
}

// func (j *jiraClient) CountIssuesByFilterID(filterID string) (int, error) {
// 	return 34, nil
// }
// func (j *jiraClient) FindIssuesByFilterID(filterID string, max int, cfields []string) (chan map[string]interface{}, error) {
// 	iChan := make(chan map[string]interface{})
// 	go func() {
// 		defer close(iChan)
// 		for i := 1; i <= 10; i++ {
// 			iChan <- map[string]interface{}{
// 				"f":       strconv.Itoa(i),
// 				"message": filterID,
// 			}
// 		}
// 	}()
// 	return iChan, nil
// }
