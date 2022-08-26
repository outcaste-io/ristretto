package ristretto

import (
	"testing"
	"time"

	"github.com/dgraph-io/ristretto/z"
	"github.com/stretchr/testify/require"
)

func TestStoreSetGet(t *testing.T) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	i := Item{
		Key:      key,
		Conflict: conflict,
		Value:    2,
	}
	s.Set(&i)
	val, ok := s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 2, val.(int))

	i.Value = 3
	s.Set(&i)
	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 3, val.(int))

	key, conflict = z.KeyToHash(2)
	i = Item{
		Key:      key,
		Conflict: conflict,
		Value:    2,
	}
	s.Set(&i)
	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 2, val.(int))
}

func TestStoreDel(t *testing.T) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	i := Item{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	s.Del(key, conflict)
	val, ok := s.Get(key, conflict)
	require.False(t, ok)
	require.Nil(t, val)

	s.Del(2, 0)
}

func TestStoreClear(t *testing.T) {
	s := newShardedMap(nil)
	for i := uint64(0); i < 1000; i++ {
		key, conflict := z.KeyToHash(i)
		it := Item{
			Key:      key,
			Conflict: conflict,
			Value:    i,
		}
		s.Set(&it)
	}
	s.Clear(nil)
	for i := uint64(0); i < 1000; i++ {
		key, conflict := z.KeyToHash(i)
		val, ok := s.Get(key, conflict)
		require.False(t, ok)
		require.Nil(t, val)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	i := Item{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	i.Value = 2
	_, ok := s.Update(&i)
	require.True(t, ok)

	val, ok := s.Get(key, conflict)
	require.True(t, ok)
	require.NotNil(t, val)

	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 2, val.(int))

	i.Value = 3
	_, ok = s.Update(&i)
	require.True(t, ok)

	val, ok = s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 3, val.(int))

	key, conflict = z.KeyToHash(2)
	i = Item{
		Key:      key,
		Conflict: conflict,
		Value:    2,
	}
	_, ok = s.Update(&i)
	require.False(t, ok)
	val, ok = s.Get(key, conflict)
	require.False(t, ok)
	require.Nil(t, val)
}

func TestStoreCollision(t *testing.T) {
	s := newShardedMap(nil)
	s.shards[1].Lock()
	s.shards[1].data[1] = storeItem{
		key:      1,
		conflict: 0,
		value:    1,
	}
	s.shards[1].Unlock()
	val, ok := s.Get(1, 1)
	require.False(t, ok)
	require.Nil(t, val)

	i := Item{
		Key:      1,
		Conflict: 1,
		Value:    2,
	}
	s.Set(&i)
	val, ok = s.Get(1, 0)
	require.True(t, ok)
	require.NotEqual(t, 2, val.(int))

	_, ok = s.Update(&i)
	require.False(t, ok)
	val, ok = s.Get(1, 0)
	require.True(t, ok)
	require.NotEqual(t, 2, val.(int))

	s.Del(1, 1)
	val, ok = s.Get(1, 0)
	require.True(t, ok)
	require.NotNil(t, val)
}

func TestStoreExpiration(t *testing.T) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	expiration := time.Now().Add(time.Second)
	i := Item{
		Key:        key,
		Conflict:   conflict,
		Value:      1,
		Expiration: expiration,
	}
	s.Set(&i)
	val, ok := s.Get(key, conflict)
	require.True(t, ok)
	require.Equal(t, 1, val.(int))

	ttl := s.Expiration(key)
	require.Equal(t, expiration, ttl)

	s.Del(key, conflict)

	_, ok = s.Get(key, conflict)
	require.False(t, ok)
	require.True(t, s.Expiration(key).IsZero())

	// missing item
	key, _ = z.KeyToHash(4340958203495)
	ttl = s.Expiration(key)
	require.True(t, ttl.IsZero())
}

func BenchmarkStoreGet(b *testing.B) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	i := Item{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Get(key, conflict)
		}
	})
}

func BenchmarkStoreSet(b *testing.B) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			i := Item{
				Key:      key,
				Conflict: conflict,
				Value:    1,
			}
			s.Set(&i)
		}
	})
}

func BenchmarkStoreUpdate(b *testing.B) {
	s := newShardedMap(nil)
	key, conflict := z.KeyToHash(1)
	i := Item{
		Key:      key,
		Conflict: conflict,
		Value:    1,
	}
	s.Set(&i)
	b.SetBytes(1)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Update(&Item{
				Key:      key,
				Conflict: conflict,
				Value:    2,
			})
		}
	})
}
