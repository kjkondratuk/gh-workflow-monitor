package github_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListAllFailedWorkflows(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mocks.MockClient)
		days         int
		wantErr      bool
		wantFailures int
	}{
		{
			name: "successful retrieval",
			setupMock: func(m *mocks.MockClient) {
				failures := map[string][]github.WorkflowFailure{
					"https://github.com/owner/repo1/pull/1": {
						{
							Repo:      "repo1",
							PRNumber:  1,
							Workflow:  "workflow1",
							StartedAt: time.Now(),
							URL:       "https://github.com/owner/repo1/actions/runs/123",
							PRURL:     "https://github.com/owner/repo1/pull/1",
						},
					},
					"https://github.com/owner/repo2/pull/2": {
						{
							Repo:      "repo2",
							PRNumber:  2,
							Workflow:  "workflow2",
							StartedAt: time.Now(),
							URL:       "https://github.com/owner/repo2/actions/runs/456",
							PRURL:     "https://github.com/owner/repo2/pull/2",
						},
					},
				}
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(failures, nil)
			},
			days:         7,
			wantErr:      false,
			wantFailures: 2,
		},
		{
			name: "error from client",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(nil, fmt.Errorf("mock error"))
			},
			days:         7,
			wantErr:      true,
			wantFailures: 0,
		},
		{
			name: "no failures",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(map[string][]github.WorkflowFailure{}, nil)
			},
			days:         7,
			wantErr:      false,
			wantFailures: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			tt.setupMock(mockClient)

			failures, err := mockClient.ListAllFailedWorkflows(context.Background(), tt.days)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantFailures, len(failures))
		})
	}
}

func TestGetFailedWorkflows(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockClient)
		prNumber  string
		repo      string
		wantErr   bool
	}{
		{
			name: "successful retrieval",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().GetFailedWorkflows(mock.Anything, "123", "test-repo").Return(nil)
			},
			prNumber: "123",
			repo:     "test-repo",
			wantErr:  false,
		},
		{
			name: "error from client",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().GetFailedWorkflows(mock.Anything, "123", "test-repo").Return(fmt.Errorf("mock error"))
			},
			prNumber: "123",
			repo:     "test-repo",
			wantErr:  true,
		},
		{
			name: "empty repository",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().GetFailedWorkflows(mock.Anything, "123", "").Return(fmt.Errorf("repository name is required"))
			},
			prNumber: "123",
			repo:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			tt.setupMock(mockClient)

			err := mockClient.GetFailedWorkflows(context.Background(), tt.prNumber, tt.repo)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
