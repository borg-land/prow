/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package close

import (
	"fmt"
	"regexp"

	"k8s.io/test-infra/prow/github"
	"k8s.io/test-infra/prow/plugins"
)

const pluginName = "close"

var closeRe = regexp.MustCompile(`(?mi)^\/close\r?$`)

func init() {
	plugins.RegisterIssueCommentHandler(pluginName, handleIssueComment)
}

type githubClient interface {
	CreateComment(owner, repo string, number int, comment string) error
	CloseIssue(owner, repo string, number int) error
}

func handleIssueComment(pa *plugins.PluginAgent, ic github.IssueCommentEvent) error {
	return handle(pa.GitHubClient, ic)
}

func handle(gc githubClient, ic github.IssueCommentEvent) error {
	// Only consider open issues and new comments.
	if ic.Issue.State != "open" || ic.Issue.IsPullRequest() || ic.Action != "created" {
		return nil
	}

	if !closeRe.MatchString(ic.Comment.Body) {
		return nil
	}

	org := ic.Repo.Owner.Login
	repo := ic.Repo.Name
	number := ic.Issue.Number

	commentAuthor := ic.Comment.User.Login

	// Allow assignees and authors to close issues.
	if !ic.Issue.IsAuthor(commentAuthor) && !ic.Issue.IsAssignee(commentAuthor) {
		return gc.CreateComment(org, repo, number, fmt.Sprintf("@%s: you can't close an issue unless you authored it or you are assigned to it.", commentAuthor))
	}

	return gc.CloseIssue(org, repo, number)
}
