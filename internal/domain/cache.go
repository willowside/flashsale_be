package domain

type CacheRepository interface {
	Set(key string, value interface{}, ttlSeconds int) error
	Get(key string) (string, error)
	Del(key string) error
}
