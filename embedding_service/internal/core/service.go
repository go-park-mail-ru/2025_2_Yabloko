package core

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/shota3506/onnxruntime-purego/onnxruntime"
)

type EmbeddingService struct {
	runtime      *onnxruntime.Runtime
	env          *onnxruntime.Env
	session      *onnxruntime.Session
	vocab        map[string]int64
	reverseVocab map[int64]string
	cache        map[string][]float32
	mu           sync.RWMutex
	logger       *slog.Logger
	dims         int
	maxSeqLen    int
	clsToken     int64
	sepToken     int64
	padToken     int64
	unkToken     int64
}

func NewEmbeddingService(modelPath, vocabPath string, logger *slog.Logger) (*EmbeddingService, error) {
	logger.Info("loading embedding service",
		slog.String("model_path", modelPath),
		slog.String("vocab_path", vocabPath),
	)

	rt, err := onnxruntime.NewRuntime("", 23)
	if err != nil {
		logger.Error("failed to create runtime",
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create runtime: %w", err)
	}

	env, err := rt.NewEnv("embedding_service", onnxruntime.LoggingLevelWarning)
	if err != nil {
		rt.Close()
		return nil, fmt.Errorf("failed to create env: %w", err)
	}

	session, err := rt.NewSession(env, modelPath, nil)
	if err != nil {
		env.Close()
		rt.Close()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	inputNames := session.InputNames()
	outputNames := session.OutputNames()

	logger.Info("onnx model loaded",
		slog.Any("inputs", inputNames),
		slog.Any("outputs", outputNames),
	)

	vocab, reverseVocab, err := loadVocab(vocabPath)
	if err != nil {
		session.Close()
		env.Close()
		rt.Close()
		return nil, fmt.Errorf("failed to load vocab: %w", err)
	}

	clsToken := vocab["[CLS]"]
	sepToken := vocab["[SEP]"]
	padToken := vocab["[PAD]"]
	unkToken := vocab["[UNK]"]

	logger.Info("tokenizer loaded",
		slog.Int64("cls_token", clsToken),
		slog.Int64("sep_token", sepToken),
		slog.Int64("pad_token", padToken),
		slog.Int64("unk_token", unkToken),
		slog.Int("vocab_size", len(vocab)),
	)

	return &EmbeddingService{
		runtime:      rt,
		env:          env,
		session:      session,
		vocab:        vocab,
		reverseVocab: reverseVocab,
		cache:        make(map[string][]float32),
		logger:       logger,
		dims:         384,
		maxSeqLen:    128,
		clsToken:     clsToken,
		sepToken:     sepToken,
		padToken:     padToken,
		unkToken:     unkToken,
	}, nil
}

func loadVocab(vocabPath string) (map[string]int64, map[int64]string, error) {
	vocab := make(map[string]int64)
	reverseVocab := make(map[int64]string)

	content, err := readFile(vocabPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read vocab file: %w", err)
	}

	idx := int64(0)
	for _, line := range content {
		if line != "" {
			vocab[line] = idx
			reverseVocab[idx] = line
			idx++
		}
	}

	if len(vocab) == 0 {
		return nil, nil, fmt.Errorf("vocab is empty")
	}

	return vocab, reverseVocab, nil
}

func readFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var lines []string
	content := string(data)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func (e *EmbeddingService) tokenize(text string) ([]int64, []int64, []int64) {
	tokens := []int64{e.clsToken}
	attentionMask := []int64{1}
	tokenTypeIDs := []int64{0}

	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})

	for _, word := range words {
		if len(tokens) >= e.maxSeqLen-1 {
			break
		}

		word = strings.ToLower(word)

		if id, ok := e.vocab[word]; ok {
			tokens = append(tokens, id)
		} else {
			foundSubword := false
			for j := len(word); j > 0; j-- {
				prefix := word[:j]
				if j > 1 {
					prefix = "##" + prefix
				}
				if id, ok := e.vocab[prefix]; ok {
					tokens = append(tokens, id)
					foundSubword = true
					break
				}
			}
			if !foundSubword {
				tokens = append(tokens, e.unkToken)
			}
		}

		attentionMask = append(attentionMask, 1)
		tokenTypeIDs = append(tokenTypeIDs, 0)
	}

	if len(tokens) < e.maxSeqLen {
		tokens = append(tokens, e.sepToken)
		attentionMask = append(attentionMask, 1)
		tokenTypeIDs = append(tokenTypeIDs, 0)
	}

	for len(tokens) < e.maxSeqLen {
		tokens = append(tokens, e.padToken)
		attentionMask = append(attentionMask, 0)
		tokenTypeIDs = append(tokenTypeIDs, 0)
	}

	return tokens[:e.maxSeqLen], attentionMask[:e.maxSeqLen], tokenTypeIDs[:e.maxSeqLen]
}

func (e *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	e.mu.RLock()
	if cached, ok := e.cache[text]; ok {
		e.mu.RUnlock()
		e.logger.DebugContext(ctx, "embedding from cache", slog.Int("dim", len(cached)))
		return cached, nil
	}
	e.mu.RUnlock()

	inputIDs, attentionMask, tokenTypeIDs := e.tokenize(text)

	inputIDsVal, err := onnxruntime.NewTensorValue(e.runtime, inputIDs, []int64{1, int64(len(inputIDs))})
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsVal.Close()

	attnMaskVal, err := onnxruntime.NewTensorValue(e.runtime, attentionMask, []int64{1, int64(len(attentionMask))})
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attnMaskVal.Close()

	tokenTypeVal, err := onnxruntime.NewTensorValue(e.runtime, tokenTypeIDs, []int64{1, int64(len(tokenTypeIDs))})
	if err != nil {
		return nil, fmt.Errorf("failed to create token_type_ids tensor: %w", err)
	}
	defer tokenTypeVal.Close()

	inputs := map[string]*onnxruntime.Value{
		"input_ids":      inputIDsVal,
		"attention_mask": attnMaskVal,
		"token_type_ids": tokenTypeVal,
	}

	output, err := e.session.Run(ctx, inputs)
	if err != nil {
		e.logger.ErrorContext(ctx, "onnx inference failed",
			slog.Any("err", err),
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("onnx inference failed: %w", err)
	}

	var embedding []float32

	if val, ok := output["sentence_embedding"]; ok {
		data, _, err := onnxruntime.GetTensorData[float32](val)
		if err == nil && len(data) > 0 {
			embedding = data
		}
	}

	if len(embedding) == 0 {
		if val, ok := output["last_hidden_state"]; ok {
			data, shape, err := onnxruntime.GetTensorData[float32](val)
			if err == nil && len(shape) == 3 {
				// shape: [batch, seq_len, hidden_size]
				seqLen := int(shape[1])
				hidden := int(shape[2])

				hiddenState := make([][]float32, seqLen)
				for i := 0; i < seqLen; i++ {
					start := i * hidden
					end := start + hidden
					hiddenState[i] = data[start:end]
				}
				embedding = meanPooling(hiddenState, attentionMask)
			}
		}
	}

	if len(embedding) == 0 {
		e.logger.ErrorContext(ctx, "could not extract embedding from model output",
			slog.String("text", truncate(text, 50)),
		)
		return nil, fmt.Errorf("could not extract embedding from model output")
	}

	embedding = normalizeVector(embedding)

	e.mu.Lock()
	e.cache[text] = embedding
	e.mu.Unlock()

	e.logger.DebugContext(ctx, "embedding generated",
		slog.Int("dim", len(embedding)),
		slog.Int("cache_size", len(e.cache)),
	)

	return embedding, nil
}

func (e *EmbeddingService) GetEmbeddingBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty texts slice")
	}

	results := make([][]float32, len(texts))
	var successCount, errorCount int

	for i, text := range texts {
		emb, err := e.GetEmbedding(ctx, text)
		if err != nil {
			e.logger.WarnContext(ctx, "batch item failed",
				slog.Any("err", err),
				slog.Int("index", i),
				slog.String("text", truncate(text, 50)),
			)
			errorCount++
			results[i] = nil
			continue
		}
		results[i] = emb
		successCount++
	}

	if errorCount == len(texts) {
		return nil, fmt.Errorf("all %d embeddings failed", len(texts))
	}

	e.logger.InfoContext(ctx, "batch embedding completed",
		slog.Int("total", len(texts)),
		slog.Int("success", successCount),
		slog.Int("failed", errorCount),
	)

	return results, nil
}

func (e *EmbeddingService) IsHealthy() bool {
	return e != nil && e.session != nil && len(e.vocab) > 0
}

func (e *EmbeddingService) Close() error {
	if e.session != nil {
		e.session.Close()
	}
	if e.env != nil {
		e.env.Close()
	}
	if e.runtime != nil {
		e.runtime.Close()
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache = nil
	e.vocab = nil
	e.reverseVocab = nil

	return nil
}

func meanPooling(hiddenState [][]float32, attentionMask []int64) []float32 {
	if len(hiddenState) == 0 {
		return nil
	}

	dims := len(hiddenState[0])
	sum := make([]float32, dims)
	var count float32

	for i, vec := range hiddenState {
		if i < len(attentionMask) && attentionMask[i] == 1 {
			for j := 0; j < dims; j++ {
				sum[j] += vec[j]
			}
			count++
		}
	}

	if count == 0 {
		return make([]float32, dims)
	}

	for j := 0; j < dims; j++ {
		sum[j] /= count
	}

	return sum
}

func normalizeVector(v []float32) []float32 {
	var norm float32
	for _, x := range v {
		norm += x * x
	}

	if norm == 0 {
		return v
	}

	normSqrt := float32(math.Sqrt(float64(norm)))
	for i := range v {
		v[i] /= normSqrt
	}

	return v
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
