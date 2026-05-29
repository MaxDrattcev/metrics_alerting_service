package pool_test

import (
	"testing"

	"github.com/MaxDrattcev/metrics_alerting_service/internal/models"
	"github.com/MaxDrattcev/metrics_alerting_service/internal/pool"
	"github.com/stretchr/testify/require"
)

func TestPool_PutCallsReset(t *testing.T) {
	p := pool.New(func() *models.ResetableStruct {
		return &models.ResetableStruct{}
	})

	obj := p.Get()

	p.Put(obj)

	again := p.Get()
	require.NotNil(t, again)
}
