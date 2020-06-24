package ldclient

import (
	"testing"

	"gopkg.in/launchdarkly/go-server-sdk-evaluation.v1/ldbuilders"

	"github.com/stretchr/testify/assert"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldlog"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
	"gopkg.in/launchdarkly/go-server-sdk.v5/interfaces"
	"gopkg.in/launchdarkly/go-server-sdk.v5/internal"
	"gopkg.in/launchdarkly/go-server-sdk.v5/ldcomponents"
	"gopkg.in/launchdarkly/go-server-sdk.v5/sharedtest"
)

type clientExternalUpdatesTestParams struct {
	client  *LDClient
	store   interfaces.DataStore
	mockLog *sharedtest.MockLoggers
}

func withClientExternalUpdatesTestParams(callback func(clientExternalUpdatesTestParams)) {
	p := clientExternalUpdatesTestParams{}
	p.store = internal.NewInMemoryDataStore(ldlog.NewDisabledLoggers())
	p.mockLog = sharedtest.NewMockLoggers()
	config := Config{
		DataSource: ldcomponents.ExternalUpdatesOnly(),
		DataStore:  singleDataStoreFactory{p.store},
		Logging:    ldcomponents.Logging().Loggers(p.mockLog.Loggers),
	}
	p.client, _ = MakeCustomClient("sdk_key", config, 0)
	defer p.client.Close()
	callback(p)
}

func TestClientExternalUpdatesMode(t *testing.T) {
	t.Run("is initialized", func(t *testing.T) {
		withClientExternalUpdatesTestParams(func(p clientExternalUpdatesTestParams) {
			assert.True(t, p.client.Initialized())
			assert.Equal(t, interfaces.DataSourceStateValid,
				p.client.GetDataSourceStatusProvider().GetStatus().State)
		})
	})

	t.Run("reports non-offline status", func(t *testing.T) {
		withClientExternalUpdatesTestParams(func(p clientExternalUpdatesTestParams) {
			assert.False(t, p.client.IsOffline())
		})
	})

	t.Run("logs appropriate message at startup", func(t *testing.T) {
		withClientExternalUpdatesTestParams(func(p clientExternalUpdatesTestParams) {
			assert.Contains(
				t,
				p.mockLog.GetOutput(ldlog.Info),
				"LaunchDarkly client will not connect to Launchdarkly for feature flag data",
			)
		})
	})

	t.Run("uses data from store", func(t *testing.T) {
		flag := ldbuilders.NewFlagBuilder("flagkey").SingleVariation(ldvalue.Bool(true)).Build()

		withClientExternalUpdatesTestParams(func(p clientExternalUpdatesTestParams) {
			upsertFlag(p.store, &flag)
			result, err := p.client.BoolVariation(flag.Key, evalTestUser, false)
			assert.NoError(t, err)
			assert.True(t, result)
		})
	})
}
