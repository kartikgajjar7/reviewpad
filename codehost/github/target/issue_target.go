// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package target

import (
	"context"

	"github.com/google/go-github/v45/github"
	"github.com/reviewpad/reviewpad/v3/codehost"
	gh "github.com/reviewpad/reviewpad/v3/codehost/github"
	"github.com/reviewpad/reviewpad/v3/handler"
)

type IssueTarget struct {
	*CommonTarget

	ctx          context.Context
	targetEntity *handler.TargetEntity
	githubClient *gh.GithubClient
	issue        *github.Issue
}

// ensure IssueTarget conforms to Target interface
var _ codehost.Target = (*IssueTarget)(nil)

func NewIssueTarget(ctx context.Context, targetEntity *handler.TargetEntity, githubClient *gh.GithubClient, issue *github.Issue) *IssueTarget {
	return &IssueTarget{
		NewCommonTarget(ctx, targetEntity, githubClient),
		ctx,
		targetEntity,
		githubClient,
		issue,
	}
}

func (t *IssueTarget) GetNodeID() string {
	return t.issue.GetNodeID()
}

func (t *IssueTarget) Close() error {
	ctx := t.ctx
	targetEntity := t.targetEntity
	owner := targetEntity.Owner
	repo := targetEntity.Repo
	number := targetEntity.Number
	issue := t.issue
	issue.State = github.String("closed")
	issueRequest := &github.IssueRequest{
		State: issue.State,
	}

	_, _, err := t.githubClient.EditIssue(ctx, owner, repo, number, issueRequest)

	return err
}

func (t *IssueTarget) GetLabels() ([]*codehost.Label, error) {
	issue := t.issue
	labels := make([]*codehost.Label, len(issue.Labels))

	for i, label := range issue.Labels {
		labels[i] = &codehost.Label{
			ID:   *label.ID,
			Name: *label.Name,
		}
	}

	return labels, nil
}

func (t *IssueTarget) GetAuthor() (*codehost.User, error) {
	issue := t.issue

	return &codehost.User{
		Login: *issue.User.Login,
	}, nil
}

func (t *IssueTarget) GetProjectByName(name string) (*codehost.Project, error) {
	ctx := t.ctx
	targetEntity := t.targetEntity
	owner := targetEntity.Owner
	repo := targetEntity.Repo

	project, err := t.githubClient.GetProjectV2ByName(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}

	return &codehost.Project{
		ID:     project.ID,
		Number: project.Number,
	}, nil
}

func (t *IssueTarget) GetAssignees() ([]*codehost.User, error) {
	issue := t.issue
	assignees := make([]*codehost.User, len(issue.Assignees))

	for i, assignee := range issue.Assignees {
		assignees[i] = &codehost.User{
			Login: *assignee.Login,
		}
	}

	return assignees, nil
}

func (t *IssueTarget) GetCommentCount() (int, error) {
	return t.issue.GetComments(), nil
}

func (t *IssueTarget) GetCreatedAt() (string, error) {
	return t.issue.GetCreatedAt().String(), nil
}

func (t *IssueTarget) GetDescription() (string, error) {
	return t.issue.GetBody(), nil
}

func (t *IssueTarget) GetTitle() string {
	return t.issue.GetTitle()
}
