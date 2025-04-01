package embeddings

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	"github.com/copilot-extensions/rag-extension/embedding"
)

type Service struct {
	chunkSize     int
	overlap       int
	integrationID string
	apiToken      string
}

func NewService(chunkSize, overlap int, integrationID, apiToken string) *Service {
	return &Service{
		chunkSize:     chunkSize,
		overlap:       overlap,
		integrationID: integrationID,
		apiToken:      apiToken,
	}
}

func (s *Service) ProcessAndStore(filePath, text, embeddingDir string) error {
	chunks := chunkText(text, s.chunkSize, s.overlap)
	for i, chunk := range chunks {
		emb, err := embedding.Create(context.Background(), s.integrationID, s.apiToken, chunk)
		if err != nil {
			return fmt.Errorf("erro ao gerar embedding para chunk %d: %w", i, err)
		}

		baseName := filepath.Base(filePath)
		embeddingFile := fmt.Sprintf("%s_chunk_%d.embedding", baseName, i)
		textFile := fmt.Sprintf("%s_chunk_%d.txt", baseName, i)

		embPath := filepath.Join(embeddingDir, embeddingFile)
		textPath := filepath.Join(embeddingDir, textFile)

		embFile, err := os.Create(embPath)
		if err != nil {
			return fmt.Errorf("erro ao criar arquivo de embedding: %w", err)
		}
		defer embFile.Close()

		if err := json.NewEncoder(embFile).Encode(emb); err != nil {
			return fmt.Errorf("erro ao salvar embedding: %w", err)
		}

		if err := os.WriteFile(textPath, []byte(chunk), 0644); err != nil {
			return fmt.Errorf("erro ao salvar chunk: %w", err)
		}
	}
	return nil
}

func chunkText(text string, size int, overlap int) []string {
	var chunks []string
	start := 0
	for start < len(text) {
		end := start + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
		start += size - overlap
	}
	return chunks
}

func (s *Service) RetrieveRelevantChunks(query string, k int) ([]string, error) {
	embeddingDir := "embeddings/embeddings"
	files, err := filepath.Glob(filepath.Join(embeddingDir, "*.embedding"))
	if err != nil {
		return nil, fmt.Errorf("erro ao listar arquivos de embeddings: %w", err)
	}

	queryEmbedding, err := embedding.Create(context.Background(), s.integrationID, s.apiToken, query)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar embedding para a query: %w", err)
	}

	var results []struct {
		Chunk      string
		Similarity float32
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var emb []float32
		if err := json.Unmarshal(content, &emb); err != nil {
			continue
		}

		sim := cosineSimilarityVectors(queryEmbedding, emb)
		results = append(results, struct {
			Chunk      string
			Similarity float32
		}{file, sim})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	var topChunks []string
	for i := 0; i < k && i < len(results); i++ {
		topChunks = append(topChunks, results[i].Chunk)
	}

	return topChunks, nil
}

func cosineSimilarityVectors(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, magA, magB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(magA))) * float32(math.Sqrt(float64(magB))))
}
