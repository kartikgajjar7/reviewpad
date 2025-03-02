// Copyright 2022 Explore.dev Unipessoal Lda. All Rights Reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package plugins_aladino_actions

import (
	"errors"
	"strings"

	gh "github.com/reviewpad/reviewpad/v3/codehost/github"
	"github.com/reviewpad/reviewpad/v3/codehost/github/target"
	"github.com/reviewpad/reviewpad/v3/handler"
	"github.com/reviewpad/reviewpad/v3/lang/aladino"
)

var (
	ErrProjectNotFound         = errors.New("project not found")
	ErrProjectHasNoStatusField = errors.New("project has no status field")
	ErrProjectStatusNotFound   = errors.New("project status not found")
)

type AddProjectV2ItemByIdInput struct {
	ProjectID string `json:"projectId"`
	ContentID string `json:"contentId"`
	// A unique identifier for the client performing the mutation. (Optional.)
	ClientMutationID *string `json:"clientMutationId,omitempty"`
}

type FieldValue struct {
	SingleSelectOptionId string `json:"singleSelectOptionId"`
}

type UpdateProjectV2ItemFieldValueInput struct {
	ItemID    string     `json:"itemId"`
	Value     FieldValue `json:"value"`
	ProjectID string     `json:"projectId"`
	FieldID   string     `json:"fieldId"`
}

func AddToProject() *aladino.BuiltInAction {
	return &aladino.BuiltInAction{
		Type:           aladino.BuildFunctionType([]aladino.Type{aladino.BuildStringType(), aladino.BuildStringType()}, aladino.BuildStringType()),
		Code:           addToProjectCode,
		SupportedKinds: []handler.TargetEntityKind{handler.PullRequest},
	}
}

func addToProjectCode(e aladino.Env, args []aladino.Value) error {
	pr := e.GetTarget().(*target.PullRequestTarget).PullRequest
	owner := gh.GetPullRequestBaseOwnerName(pr)
	repo := gh.GetPullRequestBaseRepoName(pr)
	projectName := args[0].(*aladino.StringValue).Val
	projectStatus := strings.ToLower(args[1].(*aladino.StringValue).Val)
	totalRequestTries := 2

	project, err := e.GetGithubClient().GetProjectV2ByName(e.GetCtx(), owner, repo, projectName)
	if err != nil {
		return err
	}

	fields, err := e.GetGithubClient().GetProjectFieldsByProjectNumber(e.GetCtx(), owner, repo, project.Number, totalRequestTries)
	if err != nil {
		return err
	}

	statusField := gh.FieldDetails{}

	for _, field := range fields {
		if strings.EqualFold(field.Details.Name, "status") {
			statusField = field.Details
			break
		}
	}

	if statusField.ID == "" {
		return ErrProjectHasNoStatusField
	}

	fieldOptionID := ""

	for _, option := range statusField.Options {
		if strings.Contains(strings.ToLower(option.Name), projectStatus) {
			fieldOptionID = option.ID
			break
		}
	}

	if fieldOptionID == "" {
		return ErrProjectStatusNotFound
	}

	var addProjectV2ItemByIdMutation struct {
		AddProjectV2ItemById struct {
			Item struct {
				Id string
			}
		} `graphql:"addProjectV2ItemById(input: $input)"`
	}

	input := AddProjectV2ItemByIdInput{
		ProjectID: project.ID,
		ContentID: *pr.NodeID,
	}

	// FIXME: move mutate to a separate function in the codehost.github package
	err = e.GetGithubClient().GetClientGraphQL().Mutate(e.GetCtx(), &addProjectV2ItemByIdMutation, input, nil)
	if err != nil {
		return err
	}

	var updateProjectV2ItemFieldValueMutation struct {
		UpdateProjetV2ItemFieldValue struct {
			ClientMutationID string
		} `graphql:"updateProjectV2ItemFieldValue(input: $input)"`
	}

	updateInput := UpdateProjectV2ItemFieldValueInput{
		ProjectID: project.ID,
		ItemID:    addProjectV2ItemByIdMutation.AddProjectV2ItemById.Item.Id,
		Value: FieldValue{
			SingleSelectOptionId: fieldOptionID,
		},
		FieldID: statusField.ID,
	}

	return e.GetGithubClient().GetClientGraphQL().Mutate(e.GetCtx(), &updateProjectV2ItemFieldValueMutation, updateInput, nil)
}
