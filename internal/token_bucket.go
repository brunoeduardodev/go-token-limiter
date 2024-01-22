package token_bucket

import (
	"errors"
	"math"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens        float64
	lastAccess    int64
	totalAttempts int64
}

type TokenMachine struct {
	bucketsPool     map[string]*TokenBucket
	maxTokens       int
	tokensPerMinute int
	tokenLocker     sync.Mutex
}

func (m *TokenMachine) CreateFullTokenBucket() TokenBucket {
	return TokenBucket{
		tokens:        float64(m.maxTokens),
		lastAccess:    time.Now().UnixMilli(),
		totalAttempts: 0,
	}
}

func (m *TokenMachine) RecalculateTokenBucketTokens(b *TokenBucket) {
	now := time.Now().UnixMilli()
	elapsedMinutes := float64(now-b.lastAccess) / 60000
	tokensToInsert := elapsedMinutes * float64(m.tokensPerMinute)

	newTokens := math.Min(b.tokens+tokensToInsert, float64(m.maxTokens))
	b.tokens = newTokens
	b.lastAccess = now
	b.totalAttempts++
}

func (m *TokenMachine) InsertToken(userId string) bool {
	m.tokenLocker.Lock()
	userBucket, exists := m.bucketsPool[userId]

	if !exists {
		newBucket := m.CreateFullTokenBucket()
		m.bucketsPool[userId] = &newBucket
		userBucket = m.bucketsPool[userId]
	}
	m.tokenLocker.Unlock()

	m.RecalculateTokenBucketTokens(userBucket)
	if userBucket.tokens < 1 {
		return false
	}

	userBucket.tokens = userBucket.tokens - 1
	return true
}

type BucketInformation struct {
	Tokens        float64
	LastAccess    int64
	TotalAttempts int64
}

func (m *TokenMachine) GetBucketInformation(userId string) (*BucketInformation, error) {
	m.tokenLocker.Lock()

	bucket, exist := m.bucketsPool[userId]
	m.tokenLocker.Unlock()
	if !exist {
		return nil, errors.New("could not find token bucket for specified user")
	}

	return &BucketInformation{
		Tokens:        bucket.tokens,
		LastAccess:    bucket.lastAccess,
		TotalAttempts: bucket.totalAttempts,
	}, nil
}

func MakeTokenMachine(maxTokens, tokensPerMinute int) *TokenMachine {
	return &TokenMachine{
		bucketsPool:     make(map[string]*TokenBucket),
		maxTokens:       maxTokens,
		tokensPerMinute: tokensPerMinute,
	}
}
