package storage

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"

	"github.com/lex-rio/omni-knowledge-base-assistant/internal/domain"
)

type vecEntry struct {
	id    string
	orgID string
	vec   []float32
}

type MemoryVectorStore struct {
	mu      sync.RWMutex
	entries []vecEntry
}

func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{}
}

func (m *MemoryVectorStore) Add(id string, orgID string, embedding []float32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, vecEntry{id: id, orgID: orgID, vec: embedding})
}

func (m *MemoryVectorStore) Search(query []float32, orgID string, topK int) []domain.VectorResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type scored struct {
		id    string
		score float32
	}
	var results []scored

	for _, e := range m.entries {
		if e.orgID != orgID {
			continue
		}
		s := cosineSimilarity(query, e.vec)
		results = append(results, scored{id: e.id, score: s})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	out := make([]domain.VectorResult, len(results))
	for i, r := range results {
		out[i] = domain.VectorResult{ChunkID: r.id, Score: r.score}
	}
	return out
}

func (m *MemoryVectorStore) Remove(ids []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	filtered := m.entries[:0]
	for _, e := range m.entries {
		if _, found := idSet[e.id]; !found {
			filtered = append(filtered, e)
		}
	}
	m.entries = filtered
}

func (m *MemoryVectorStore) Save(path string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	f, err := os.Create(path + ".tmp")
	if err != nil {
		return fmt.Errorf("create tmp: %w", err)
	}
	defer f.Close()

	if err := binary.Write(f, binary.LittleEndian, uint32(len(m.entries))); err != nil {
		return err
	}

	for _, e := range m.entries {
		if err := writeString(f, e.id); err != nil {
			return err
		}
		if err := writeString(f, e.orgID); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, uint32(len(e.vec))); err != nil {
			return err
		}
		if err := binary.Write(f, binary.LittleEndian, e.vec); err != nil {
			return err
		}
	}

	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(path+".tmp", path)
}

func (m *MemoryVectorStore) Load(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	var count uint32
	if err := binary.Read(f, binary.LittleEndian, &count); err != nil {
		return err
	}

	m.entries = make([]vecEntry, 0, count)
	for range count {
		id, err := readString(f)
		if err != nil {
			return err
		}
		orgID, err := readString(f)
		if err != nil {
			return err
		}
		var dim uint32
		if err := binary.Read(f, binary.LittleEndian, &dim); err != nil {
			return err
		}
		vec := make([]float32, dim)
		if err := binary.Read(f, binary.LittleEndian, &vec); err != nil {
			return err
		}
		m.entries = append(m.entries, vecEntry{id: id, orgID: orgID, vec: vec})
	}
	return nil
}

func writeString(f *os.File, s string) error {
	b := []byte(s)
	if err := binary.Write(f, binary.LittleEndian, uint16(len(b))); err != nil {
		return err
	}
	_, err := f.Write(b)
	return err
}

func readString(f *os.File) (string, error) {
	var length uint16
	if err := binary.Read(f, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	b := make([]byte, length)
	if _, err := f.Read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		ai, bi := float64(a[i]), float64(b[i])
		dot += ai * bi
		normA += ai * ai
		normB += bi * bi
	}
	denom := math.Sqrt(normA) * math.Sqrt(normB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}
