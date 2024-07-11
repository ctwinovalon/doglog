package cli

import (
	"context"
	"doglog/log"
	"doglog/options"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/briandowns/spinner"
)

// Print out a page of log messages that match the search criteria.
func listMessages(ctx context.Context, logsApi *datadogV2.LogsApi,
	opts *options.Options, cursor *string) (*string, bool) {
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

	items, _ := logsApi.ListLogsWithPagination(ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))

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

// Construct a datadog api client.
func apiClient(opts *options.Options) *datadog.APIClient {
	configuration := datadog.NewConfiguration()
	configuration.Debug = opts.Debug
	return datadog.NewAPIClient(configuration)
}

// Build the datadog context required for all api calls.
// The context includes the api keys.
func constructDatadogContext(opts *options.Options) context.Context {
	return context.WithValue(
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
}

// CommandListMessages Print out the log messages that match the search criteria.
// Will continue until all pages of output are displayed.
func CommandListMessages(opts *options.Options, s *spinner.Spinner) bool {
	ctx := constructDatadogContext(opts)
	logsApi := datadogV2.NewLogsApi(apiClient(opts))

	var nextId *string
	nextId = nil
	result := false
	for {
		if s != nil {
			s.Stop()
		}
		nextId, result = listMessages(ctx, logsApi, opts, nextId)
		if s != nil {
			s.Start()
		}
		if nextId == nil {
			break
		} else {
			DelayForSeconds(0.2, true)
		}
	}
	return result
}
