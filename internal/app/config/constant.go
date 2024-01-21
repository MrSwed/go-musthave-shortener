package config

const (
	scheme          = "http://"
	serverAddress   = "localhost:8080"
	baseURL         = "localhost:8080"
	fileStoragePath = "/tmp/short-url-db.json"

	envServerAddressName   = "SERVER_ADDRESS"
	envBaseURLName         = "BASE_URL"
	envFileStoragePathName = "FILE_STORAGE_PATH"

	ShortLen = 8

	APIRoute     = "/api"
	ShortenRoute = "/shorten"
)
