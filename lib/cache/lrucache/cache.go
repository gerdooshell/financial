package lrucache

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

type LRUCache[K comparable] interface {
	Add(K, interface{}) (interface{}, error)
	Read(key K) (interface{}, error)
	ReadSafe() (any, error)
}

type lruCache[K comparable] struct {
	queue      *list.List
	dictionary map[K]*list.Element
	size       int
	mu         sync.Mutex
}

func NewLRUCache[K comparable](bufferSize int) LRUCache[K] {
	return &lruCache[K]{
		queue:      list.New(),
		dictionary: make(map[K]*list.Element, bufferSize),
		size:       bufferSize,
	}
}

func (l *lruCache[K]) Add(key K, value interface{}) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.dictionary[key]; ok {
		message := fmt.Sprintf("cannot add an already existing element to lru cache. element %v already exists", key)
		return nil, errors.New(message)
	}
	cNode := CacheNode[K, any]{key, value}
	element := l.queue.PushFront(cNode)
	l.dictionary[key] = element
	var removedValue interface{} = nil
	if l.queue.Len() > l.size {
		removedValue = l.removeLeastUsed()
	}
	return removedValue, nil
}

func (l *lruCache[K]) Read(key K) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	element, ok := l.dictionary[key]
	if !ok {
		return nil, errors.New("failed reading lruCache element. key does not exist")
	}
	l.queue.MoveToFront(element)
	return element.Value.(CacheNode[K, any]).Value, nil
}

func (l *lruCache[K]) removeLeastUsed() any {
	element := l.queue.Back()
	node := l.queue.Remove(element)
	delete(l.dictionary, node.(CacheNode[K, any]).key)
	return element.Value.(CacheNode[K, any]).Value
}

func (l *lruCache[K]) ReadSafe() (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.dictionary) == 0 {
		return nil, errors.New("empty cache")
	}
	return l.queue.Front().Value.(CacheNode[K, any]).Value, nil
}

type CacheNode[K comparable, V any] struct {
	key   K
	Value V
}
