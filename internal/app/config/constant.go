package config

const (
	ServerShutdownTimeout = 30

	scheme          = "http://"
	serverAddress   = "localhost:8080"
	baseURL         = "localhost:8080"
	fileStoragePath = "/tmp/short-url-db.json"

	envServerAddressName   = "SERVER_ADDRESS"
	envBaseURLName         = "BASE_URL"
	envFileStoragePathName = "FILE_STORAGE_PATH"
	envNameDBDSN           = "DATABASE_DSN"

	ShortLen = 8

	APIRoute     = "/api"
	ShortenRoute = "/shorten"
	BatchRoute   = "/batch"

	DBTableName = "shortener"
)
