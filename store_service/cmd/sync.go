package cmd

import (
	"apple_backend/store_service/internal/repository"
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type EmbeddingClient interface {
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
	GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error)
}

func SyncAllEmbeddings(
	ctx context.Context,
	db *pgxpool.Pool,
	embeddingClient EmbeddingClient,
	log *slog.Logger,
) error {
	storeRepo := repository.NewStoreRepoPostgres(db)
	itemRepo := repository.NewItemRepoPostgres(db)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := syncStoresEmbeddings(ctx, storeRepo, embeddingClient, log); err != nil {
			log.Error("store sync failed", "err", err)
			errCh <- fmt.Errorf("store sync failed: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := syncItemsEmbeddings(ctx, itemRepo, embeddingClient, log); err != nil {
			log.Error("item sync failed", "err", err)
			errCh <- fmt.Errorf("item sync failed: %w", err)
		}
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errs []error
	for err := range errCh {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		log.Error("embedding synchronization had errors",
			"total_errors", len(errs),
			"errors", fmt.Sprintf("%v", errs),
		)
		return fmt.Errorf("embedding sync completed with %d errors: %v", len(errs), errs)
	}

	return nil
}

func syncStoresEmbeddings(
	ctx context.Context,
	storeRepo *repository.StoreRepoPostgres,
	embeddingClient EmbeddingClient,
	log *slog.Logger,
) error {
	log.Info("syncing store embeddings")

	stores, err := storeRepo.GetStoresWithoutEmbedding(ctx)
	if err != nil {
		log.Error("failed to get stores", "err", err)
		return err
	}

	if len(stores) == 0 {
		log.Info("all stores have embeddings")
		return nil
	}

	log.Info("found stores without embeddings", "count", len(stores))

	successCount := 0
	failCount := 0

	for i, store := range stores {
		select {
		case <-ctx.Done():
			log.Warn("sync interrupted",
				"processed", i,
				"total", len(stores),
				"success", successCount,
				"failed", failCount,
			)
			return ctx.Err()
		default:
		}

		embedding, err := embeddingClient.GetEmbedding(ctx, store.Description)
		if err != nil {
			log.Warn("failed to get embedding for store",
				"id", store.ID,
				"name", store.Name,
				"err", err,
			)
			failCount++
			continue
		}

		if len(embedding) == 0 {
			log.Warn("received empty embedding for store",
				"id", store.ID,
				"name", store.Name,
			)
			failCount++
			continue
		}

		if err := storeRepo.UpdateStoreEmbedding(ctx, store.ID, embedding); err != nil {
			log.Warn("failed to update store embedding",
				"id", store.ID,
				"err", err,
			)
			failCount++
			continue
		}

		successCount++

		if (i+1)%10 == 0 {
			log.Info("store embeddings progress",
				"processed", i+1,
				"total", len(stores),
				"success", successCount,
				"failed", failCount,
			)
		}
	}

	log.Info("store embeddings sync completed",
		"total", len(stores),
		"success", successCount,
		"failed", failCount,
	)

	if successCount == 0 && len(stores) > 0 {
		return fmt.Errorf("failed to sync any stores: %d/%d", failCount, len(stores))
	}

	return nil
}

func syncItemsEmbeddings(
	ctx context.Context,
	itemRepo *repository.ItemRepoPostgres,
	embeddingClient EmbeddingClient,
	log *slog.Logger,
) error {
	log.Info("syncing item embeddings")

	items, err := itemRepo.GetItemsWithoutEmbedding(ctx)
	if err != nil {
		log.Error("failed to get items", "err", err)
		return err
	}

	if len(items) == 0 {
		log.Info("all items have embeddings")
		return nil
	}

	log.Info("found items without embeddings", "count", len(items))

	successCount := 0
	failCount := 0

	for i, item := range items {
		select {
		case <-ctx.Done():
			log.Warn("sync interrupted",
				"processed", i,
				"total", len(items),
				"success", successCount,
				"failed", failCount,
			)
			return ctx.Err()
		default:
		}

		text := fmt.Sprintf("%s %s", item.Name, item.Description)
		embedding, err := embeddingClient.GetEmbedding(ctx, text)
		if err != nil {
			log.Warn("failed to get embedding for item",
				"id", item.ID,
				"name", item.Name,
				"err", err,
			)
			failCount++
			continue
		}

		if len(embedding) == 0 {
			log.Warn("received empty embedding for item",
				"id", item.ID,
				"name", item.Name,
			)
			failCount++
			continue
		}

		if err := itemRepo.UpdateItemEmbedding(ctx, item.ID, embedding); err != nil {
			log.Warn("failed to update item embedding",
				"id", item.ID,
				"err", err,
			)
			failCount++
			continue
		}

		successCount++

		if (i+1)%10 == 0 {
			log.Info("item embeddings progress",
				"processed", i+1,
				"total", len(items),
				"success", successCount,
				"failed", failCount,
			)
		}
	}

	log.Info("item embeddings sync completed",
		"total", len(items),
		"success", successCount,
		"failed", failCount,
	)

	if successCount == 0 && len(items) > 0 {
		return fmt.Errorf("failed to sync any items: %d/%d", failCount, len(items))
	}

	return nil
}
