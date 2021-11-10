package background

import (
	"context"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

var maxCacheEntriesSize, _ = strconv.Atoi(env.Get(
	"SRC_BATCH_CHANGES_MAX_CACHE_SIZE_MB",
	"5000",
	"Maximum size of the batch_spec_execution_cache_entries.value column. Value is megabytes.",
))

const cacheCleanInterval = 1 * time.Hour

func newCacheEntryCleanerJob(ctx context.Context, s *store.Store) goroutine.BackgroundRoutine {
	maxSizeByte := int64(maxCacheEntriesSize * 1024 * 1024)

	return goroutine.NewPeriodicGoroutine(
		ctx,
		cacheCleanInterval,
		goroutine.NewHandlerWithErrorMessage("cleaning up LRU batch spec execution cache entries", func(ctx context.Context) error {
			return s.CleanBatchSpecExecutionCacheEntries(ctx, maxSizeByte)
		}),
	)
}
