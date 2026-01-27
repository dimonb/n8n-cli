package unit

import (
	"bytes"
	"errors"
	"testing"

	"github.com/edenreich/n8n-cli/n8n"
	"github.com/edenreich/n8n-cli/n8n/clientfakes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestListWorkflowsCommand(t *testing.T) {
	fakeClient := &clientfakes.FakeClientInterface{}
	var stdout, stderr bytes.Buffer

	setupTestCommand := func() *cobra.Command {
		viper.Set("api_key", "test-api-key")
		viper.Set("instance_url", "http://localhost:5678")

		cmd := &cobra.Command{
			Use: "list",
			RunE: func(cmd *cobra.Command, args []string) error {
				var pageLimit *int
				if pageLimitVal, _ := cmd.Flags().GetInt("page-limit"); pageLimitVal > 0 {
					pageLimit = &pageLimitVal
				}

				workflowList, err := fakeClient.GetWorkflows(pageLimit)
				if err != nil {
					return err
				}

				if workflowList == nil || workflowList.Data == nil || len(*workflowList.Data) == 0 {
					cmd.Println("No workflows found")
					return nil
				}

				for _, workflow := range *workflowList.Data {
					cmd.Printf("ID: %s, Name: %s\n", *workflow.Id, workflow.Name)
				}
				return nil
			},
		}

		cmd.SetOut(&stdout)
		cmd.SetErr(&stderr)
		cmd.Flags().StringP("output", "o", "table", "Output format")
		cmd.Flags().IntP("page-limit", "l", 0, "Page size for API request (default: 100, max: 250)")

		stdout.Reset()
		stderr.Reset()

		return cmd
	}

	createSampleWorkflowList := func(count int) *n8n.WorkflowList {
		workflows := make([]n8n.Workflow, count)
		for i := 0; i < count; i++ {
			id := string(rune('1' + i))
			name := "Workflow " + id
			active := true
			workflows[i] = n8n.Workflow{
				Id:     &id,
				Name:   name,
				Active: &active,
			}
		}
		return &n8n.WorkflowList{Data: &workflows}
	}

	t.Run("calls GetWorkflows with nil when no page-limit specified", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetWorkflowsReturns(createSampleWorkflowList(3), nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Equal(t, 1, fakeClient.GetWorkflowsCallCount())
		pageLimit := fakeClient.GetWorkflowsArgsForCall(0)
		assert.Nil(t, pageLimit, "Expected pageLimit to be nil when not specified")
	})

	t.Run("calls GetWorkflows with pageLimit when --page-limit is specified", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetWorkflowsReturns(createSampleWorkflowList(5), nil)

		err := cmd.Flags().Set("page-limit", "5")
		assert.NoError(t, err)

		err = cmd.Execute()

		assert.NoError(t, err)
		pageLimit := fakeClient.GetWorkflowsArgsForCall(fakeClient.GetWorkflowsCallCount() - 1)
		assert.NotNil(t, pageLimit, "Expected pageLimit to be set")
		assert.Equal(t, 5, *pageLimit)
	})

	t.Run("returns error when client fails", func(t *testing.T) {
		cmd := setupTestCommand()
		fakeClient.GetWorkflowsReturns(nil, errors.New("API error"))

		err := cmd.Execute()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "API error")
	})

	t.Run("shows empty message when no workflows", func(t *testing.T) {
		cmd := setupTestCommand()
		emptyData := []n8n.Workflow{}
		fakeClient.GetWorkflowsReturns(&n8n.WorkflowList{Data: &emptyData}, nil)

		err := cmd.Execute()

		assert.NoError(t, err)
		assert.Contains(t, stdout.String(), "No workflows found")
	})

	t.Run("MaxPageLimit constant matches n8n API max of 250", func(t *testing.T) {
		assert.Equal(t, 250, n8n.MaxPageLimit, "MaxPageLimit should be 250 per n8n API docs")
	})
}
