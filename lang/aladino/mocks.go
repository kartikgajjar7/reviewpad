// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package aladino

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/reviewpad/host-event-handler/handler"
	gh "github.com/reviewpad/reviewpad/v3/codehost/github"
	"github.com/reviewpad/reviewpad/v3/collector"
	"github.com/shurcooL/githubv4"
)

const DefaultMockPrID = 1234
const DefaultMockPrNum = 6
const DefaultMockPrOwner = "foobar"
const DefaultMockPrRepoName = "default-mock-repo"

var DefaultMockPrDate = time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC)
var DefaultMockContext = context.Background()
var DefaultMockCollector = collector.NewCollector("", "", "pull_request", "")
var DefaultMockTargetEntity = &handler.TargetEntity{
	Owner:  DefaultMockPrOwner,
	Repo:   DefaultMockPrRepoName,
	Number: DefaultMockPrNum,
	Kind:   handler.PullRequest,
}

func GetDefaultMockPullRequestDetails() *github.PullRequest {
	prNum := DefaultMockPrNum
	prId := int64(DefaultMockPrID)
	prOwner := DefaultMockPrOwner
	prRepoName := DefaultMockPrRepoName
	prUrl := fmt.Sprintf("https://api.github.com/repos/%v/%v/pulls/%v", prOwner, prRepoName, prNum)
	prDate := DefaultMockPrDate

	return &github.PullRequest{
		ID:   &prId,
		User: &github.User{Login: github.String("john")},
		Assignees: []*github.User{
			{Login: github.String("jane")},
		},
		Title:     github.String("Amazing new feature"),
		Body:      github.String("Please pull these awesome changes in!"),
		URL:       github.String("https://foo.bar"),
		CreatedAt: &prDate,
		Comments:  github.Int(6),
		Commits:   github.Int(5),
		Number:    github.Int(prNum),
		Milestone: &github.Milestone{
			Title: github.String("v1.0"),
		},
		Labels: []*github.Label{
			{
				ID:   github.Int64(1),
				Name: github.String("enhancement"),
			},
		},
		Head: &github.PullRequestBranch{
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("john"),
				},
				URL:  github.String(prUrl),
				Name: github.String(prRepoName),
			},
			Ref: github.String("new-topic"),
		},
		Base: &github.PullRequestBranch{
			Repo: &github.Repository{
				Owner: &github.User{
					Login: github.String("john"),
				},
				URL:  github.String(prUrl),
				Name: github.String(prRepoName),
			},
			Ref: github.String("master"),
		},
		RequestedReviewers: []*github.User{
			{Login: github.String("jane")},
		},
	}
}

func GetDefaultMockPullRequestDetailsWith(pr *github.PullRequest) *github.PullRequest {
	defaultPullRequest := GetDefaultMockPullRequestDetails()

	if pr.User != nil {
		defaultPullRequest.User = pr.User
	}

	if pr.Number != nil {
		defaultPullRequest.Number = pr.Number
	}

	if pr.Head != nil {
		defaultPullRequest.Head = pr.Head
	}

	if pr.Base != nil {
		defaultPullRequest.Base = pr.Base
	}

	if pr.Assignees != nil {
		defaultPullRequest.Assignees = pr.Assignees
	}

	if pr.Commits != nil {
		defaultPullRequest.Commits = pr.Commits
	}

	if pr.Labels != nil {
		defaultPullRequest.Labels = pr.Labels
	}

	if pr.Milestone != nil {
		defaultPullRequest.Milestone = pr.Milestone
	}

	if pr.RequestedReviewers != nil {
		defaultPullRequest.RequestedReviewers = pr.RequestedReviewers
	}

	if pr.RequestedTeams != nil {
		defaultPullRequest.RequestedTeams = pr.RequestedTeams
	}

	if pr.Additions != nil {
		defaultPullRequest.Additions = pr.Additions
	}

	if pr.Deletions != nil {
		defaultPullRequest.Deletions = pr.Deletions
	}

	if pr.Title != nil {
		defaultPullRequest.Title = pr.Title
	}

	if pr.Body != nil {
		defaultPullRequest.Body = pr.Body
	}

	if pr.Draft != nil {
		defaultPullRequest.Draft = pr.Draft
	}

	if pr.NodeID != nil {
		defaultPullRequest.NodeID = pr.NodeID
	}

	return defaultPullRequest
}

func getDefaultMockPullRequestFileList() []*github.CommitFile {
	prRepoName := DefaultMockPrRepoName
	return []*github.CommitFile{
		{
			Filename: github.String(fmt.Sprintf("%v/file1.ts", prRepoName)),
			Patch:    github.String("@@ -2,9 +2,11 @@ package main\n- func previous1() {\n+ func new1() {\n+\nreturn"),
		},
		{
			Filename: github.String(fmt.Sprintf("%v/file2.ts", prRepoName)),
			Patch:    github.String("@@ -2,9 +2,11 @@ package main\n- func previous2() {\n+ func new2() {\n+\nreturn"),
		},
		{
			Filename: github.String(fmt.Sprintf("%v/file3.ts", prRepoName)),
			Patch:    github.String("@@ -2,9 +2,11 @@ package main\n- func previous3() {\n+ func new3() {\n+\nreturn"),
		},
	}
}

func MockBuiltIns() *BuiltIns {
	return &BuiltIns{
		Functions: map[string]*BuiltInFunction{
			"emptyFunction": {
				Type: BuildFunctionType([]Type{}, nil),
				Code: func(e Env, args []Value) (Value, error) {
					return nil, nil
				},
				SupportedKinds: []handler.TargetEntityKind{handler.PullRequest, handler.Issue},
			},
			"zeroConst": {
				Type: BuildFunctionType([]Type{}, BuildIntType()),
				Code: func(e Env, args []Value) (Value, error) {
					return BuildIntValue(0), nil
				},
				SupportedKinds: []handler.TargetEntityKind{handler.PullRequest, handler.Issue},
			},
			"returnStr": {
				Type: BuildFunctionType([]Type{BuildStringType()}, BuildStringType()),
				Code: func(e Env, args []Value) (Value, error) {
					return args[0].(*StringValue), nil
				},
				SupportedKinds: []handler.TargetEntityKind{handler.PullRequest, handler.Issue},
			},
		},
		Actions: map[string]*BuiltInAction{
			"emptyAction": {
				Type: BuildFunctionType([]Type{}, nil),
				Code: func(e Env, args []Value) error {
					return nil
				},
				SupportedKinds: []handler.TargetEntityKind{handler.PullRequest, handler.Issue},
			},
		},
	}
}

func mockTypes() map[string]Type {
	mockedBuiltIns := MockBuiltIns()
	builtInsType := make(map[string]Type)

	for builtInName, builtInFunction := range mockedBuiltIns.Functions {
		builtInsType[builtInName] = builtInFunction.Type
	}

	for builtInName, builtInAction := range mockedBuiltIns.Actions {
		builtInsType[builtInName] = builtInAction.Type
	}

	return builtInsType
}

func MockTypeEnv() TypeEnv {
	return TypeEnv(mockTypes())
}

func mockHttpClientWith(clientOptions ...mock.MockBackendOption) *http.Client {
	return mock.NewMockedHTTPClient(clientOptions...)
}

func mockEnvWith(prOwner string, prRepoName string, prNum int, githubClient *gh.GithubClient, eventPayload interface{}, builtIns *BuiltIns) (Env, error) {
	ctx := context.Background()

	env, err := NewEvalEnv(
		ctx,
		false,
		githubClient,
		DefaultMockCollector,
		DefaultMockTargetEntity,
		eventPayload,
		builtIns,
	)
	if err != nil {
		return nil, err
	}

	return env, nil
}

// mockDefaultHttpClient mocks an HTTP client with default values and ready for Aladino.
// Being ready for Aladino means that at least the request to build Aladino Env need to be mocked.
// As for now, at least two github request need to be mocked in order to build Aladino Env, mainly:
// - Get pull request details (i.e. /repos/{owner}/{repo}/pulls/{pull_number})
// - Get pull request files (i.e. /repos/{owner}/{repo}/pulls/{pull_number}/files)
func mockDefaultHttpClient(clientOptions []mock.MockBackendOption) *http.Client {
	defaultMocks := []mock.MockBackendOption{
		mock.WithRequestMatchHandler(
			// Mock request to get pull request details
			mock.GetReposPullsByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write(mock.MustMarshal(GetDefaultMockPullRequestDetails()))
			}),
		),
		mock.WithRequestMatchHandler(
			// Mock request to get pull request changed files
			mock.GetReposPullsFilesByOwnerByRepoByPullNumber,
			http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Write(mock.MustMarshal(getDefaultMockPullRequestFileList()))
			}),
		),
	}

	mocks := append(clientOptions, defaultMocks...)

	return mockHttpClientWith(mocks...)
}

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	return w.Result(), nil
}

func MockDefaultGithubClient(ghApiClientOptions []mock.MockBackendOption, ghGraphQLHandler func(http.ResponseWriter, *http.Request)) *gh.GithubClient {
	client := github.NewClient(mockDefaultHttpClient(ghApiClientOptions))

	// Handle GraphQL
	var clientGQL *githubv4.Client
	if ghGraphQLHandler != nil {
		mux := http.NewServeMux()
		mux.HandleFunc("/graphql", ghGraphQLHandler)
		clientGQL = githubv4.NewClient(&http.Client{Transport: localRoundTripper{handler: mux}})
	}

	return gh.NewGithubClient(client, clientGQL)
}

// MockDefaultEnv mocks an Aladino Env with default values.
func MockDefaultEnv(
	t *testing.T,
	ghApiClientOptions []mock.MockBackendOption,
	ghGraphQLHandler func(http.ResponseWriter, *http.Request),
	builtIns *BuiltIns,
	eventPayload interface{},
) Env {
	prOwner := DefaultMockPrOwner
	prRepoName := DefaultMockPrRepoName
	prNum := DefaultMockPrNum
	githubClient := MockDefaultGithubClient(ghApiClientOptions, ghGraphQLHandler)

	mockedEnv, err := mockEnvWith(prOwner, prRepoName, prNum, githubClient, eventPayload, builtIns)
	if err != nil {
		t.Fatalf("[MockDefaultEnv] failed to create mock env: %v", err)
	}

	return mockedEnv
}

func MustRead(r io.Reader) string {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func MustWrite(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}
