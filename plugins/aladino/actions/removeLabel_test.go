// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions_test

import (
	"net/http"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/gorilla/mux"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
	plugins_aladino "github.com/reviewpad/reviewpad/v3/plugins/aladino"
	"github.com/stretchr/testify/assert"
)

var removeLabel = plugins_aladino.PluginBuiltIns().Actions["removeLabel"].Code

func TestRemoveLabel_WhenLabelIsNotAppliedToPullRequest(t *testing.T) {
	wantLabel := "bug"
	var isLabelRemoved bool
	mockedEnv := aladino.MockDefaultEnv(
		t,
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposLabelsByOwnerByRepoByName,
				github.Label{
					Name: github.String(wantLabel),
					ID:   github.Int64(1),
				},
			),
			mock.WithRequestMatchHandler(
				mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// If the remove label request was performed then the label was removed
					isLabelRemoved = true
				}),
			),
		},
		nil,
		aladino.MockBuiltIns(),
		nil,
	)

	args := []aladino.Value{aladino.BuildStringValue(wantLabel)}
	err := removeLabel(mockedEnv, args)

	assert.Nil(t, err)
	assert.False(t, isLabelRemoved, "The label should not be removed")
}

func TestRemoveLabel_WhenLabelIsAppliedToPullRequestAndLabelIsInEnvironment(t *testing.T) {
	wantLabel := "enhancement"
	var gotLabel string
	mockedEnv := aladino.MockDefaultEnv(
		t,
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposLabelsByOwnerByRepoByName,
				github.Label{
					Name: github.String(wantLabel),
					ID:   github.Int64(1),
				},
			),
			mock.WithRequestMatchHandler(
				mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					vars := mux.Vars(r)
					gotLabel = vars["name"]
				}),
			),
		},
		nil,
		aladino.MockBuiltIns(),
		nil,
	)

	internalLabelID := aladino.BuildInternalLabelID(wantLabel)
	mockedEnv.GetRegisterMap()[internalLabelID] = aladino.BuildStringValue(wantLabel)

	args := []aladino.Value{aladino.BuildStringValue(wantLabel)}
	err := removeLabel(mockedEnv, args)

	assert.Nil(t, err)
	assert.Equal(t, wantLabel, gotLabel)
}

func TestRemoveLabel_WhenLabelIsAppliedToPullRequestAndLabelIsNotInEnvironment(t *testing.T) {
	wantLabel := "enhancement"
	var gotLabel string
	mockedEnv := aladino.MockDefaultEnv(
		t,
		[]mock.MockBackendOption{
			mock.WithRequestMatch(
				mock.GetReposLabelsByOwnerByRepoByName,
				github.Label{
					Name: github.String(wantLabel),
					ID:   github.Int64(1),
				},
			),
			mock.WithRequestMatchHandler(
				mock.DeleteReposIssuesLabelsByOwnerByRepoByIssueNumberByName,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					vars := mux.Vars(r)
					gotLabel = vars["name"]
				}),
			),
		},
		nil,
		aladino.MockBuiltIns(),
		nil,
	)

	args := []aladino.Value{aladino.BuildStringValue(wantLabel)}
	err := removeLabel(mockedEnv, args)

	assert.Nil(t, err)
	assert.Equal(t, wantLabel, gotLabel)
}
