package cli

import (
	"context"
	"doglog/log"
	"doglog/options"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/briandowns/spinner"
)

// CommandListMessages Print out the log messages that match the search criteria.
func listMessages(opts *options.Options, cursor *string) (nextId *string, success bool) {
	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: opts.ServerConfig.ApiKey(),
			},
			"appKeyAuth": {
				Key: opts.ServerConfig.ApplicationKey(),
			},
		},
	)

	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query:   &opts.Query,
			From:    &opts.StartDate,
			To:      &opts.EndDate,
			Indexes: opts.Indexes,
		},
		Options: &datadogV2.LogsQueryOptions{
			Timezone: datadog.PtrString("GMT"),
		},
		Page: &datadogV2.LogsListRequestPage{
			Limit:  datadog.PtrInt32(int32(opts.Limit)),
			Cursor: cursor,
		},
		Sort: datadogV2.LOGSSORT_TIMESTAMP_ASCENDING.Ptr(),
	}

	configuration := datadog.NewConfiguration()
	configuration.Debug = opts.Debug
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewLogsApi(apiClient)

	items, _ := api.ListLogsWithPagination(ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))

	for paginationResult := range items {
		if paginationResult.Error != nil {
			log.Error(*opts, "Error when calling `LogsApi.ListLogs`: %v", paginationResult.Error)
			return nil, false
		} else {
			printMessage(opts, &paginationResult.Item)
		}
	}

	return body.Page.Cursor, true
}

// CommandListMessages Print out the log messages that match the search criteria.
func CommandListMessages(opts *options.Options, s *spinner.Spinner) bool {
	var nextId *string
	nextId = nil
	result := false
	for {
		if s != nil {
			s.Stop()
		}
		nextId, result = listMessages(opts, nextId)
		if s != nil {
			s.Start()
		}
		if nextId == nil {
			break
		} else {
			DelayForSeconds(0.2)
		}
	}
	return result
}
