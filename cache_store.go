package main

type MemoryStore interface {
	AddValue(valueName string)
	//AddSeries(valueName string)
	GetValues() []string
	//GetSeries(valueName string) []int64
}

type memoryStore struct {
	values map[string]bool
}

func NewMemoryStore() MemoryStore {
	return &memoryStore{
		make(map[string]bool),
	}
}

func (ms *memoryStore) AddValue(valueName string) {
	ms.values[valueName] = true
}

func (ms *memoryStore) GetValues() []string {
	keys := make([]string, len(ms.values))

	i := 0
	for k := range ms.values {
		keys[i] = k
		i++
	}

	return keys
}
