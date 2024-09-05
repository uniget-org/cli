package cache

type Cache interface {
	WriteDataToCache(data []byte, key string) error
	CheckDataInCache(key string) bool
	ReadDataFromCache(key string) ([]byte, error)
}