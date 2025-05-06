package cli_test

import (
	"fmt"
	"testing"

	"github.com/kjkondratuk/gh-workflow-monitor/pkg/cli"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github"
	"github.com/kjkondratuk/gh-workflow-monitor/pkg/github/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleList(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockClient)
		days      int
		wantErr   bool
	}{
		{
			name: "successful list",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(nil, nil)
			},
			days:    7,
			wantErr: false,
		},
		{
			name: "error from client",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(nil, fmt.Errorf("mock error"))
			},
			days:    7,
			wantErr: true,
		},
		{
			name: "no failures",
			setupMock: func(m *mocks.MockClient) {
				m.EXPECT().ListAllFailedWorkflows(mock.Anything, 7).Return(map[string][]github.WorkflowFailure{}, nil)
			},
			days:    7,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			tt.setupMock(mockClient)

			err := cli.HandleList(mockClient, tt.days)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestHandleCheck(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mocks.MockClient)
		prNumber  string
		repo      string
		wantErr   bool
	}{
		{
			name: "successful check",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockClient(t)
			tt.setupMock(mockClient)

			err := cli.HandleCheck(mockClient, tt.prNumber, tt.repo)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
