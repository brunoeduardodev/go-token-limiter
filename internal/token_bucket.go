package token_bucket

import (
	"math"
	"time"
)

type TokenBucket struct {
	tokens     float64
	lastAccess int64
}

var usersBucketPool = map[string]TokenBucket{}

type TokenMachine struct {
	bucketsPool     map[string]TokenBucket
	maxTokens       int
	tokensPerMinute int
}

func (m *TokenMachine) CreateFullTokenBucket() TokenBucket {
	return TokenBucket{
		tokens:     float64(m.maxTokens),
		lastAccess: time.Now().UnixMilli(),
	}
}

func (m *TokenMachine) RecalculateTokenBucketTokens(b *TokenBucket) {
	now := time.Now().UnixMilli()
	elapsedMinutes := (float64)(now-b.lastAccess) / 1000
	tokensToInsert := elapsedMinutes * float64(m.tokensPerMinute)

	newTokens := math.Max(b.tokens+tokensToInsert, float64(m.maxTokens))
	b.tokens = newTokens
}

func (m *TokenMachine) InsertToken(userId string) bool {
	userBucket, exists := m.bucketsPool[userId]

	if !exists {
		userBucket := m.CreateFullTokenBucket()
		usersBucketPool[userId] = userBucket
	}

	m.RecalculateTokenBucketTokens(&userBucket)
	if userBucket.tokens < 1 {
		return false
	}

	userBucket.tokens--

	return true

}

func MakeTokenMachine(maxTokens, tokensPerMinute int) *TokenMachine {
	return &TokenMachine{
		bucketsPool:     make(map[string]TokenBucket),
		maxTokens:       maxTokens,
		tokensPerMinute: tokensPerMinute,
	}
}
