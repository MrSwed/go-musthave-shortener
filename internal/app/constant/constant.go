package constant

const (
	ServerShutdownTimeout  = 30
	ServerOperationTimeout = 30

	Scheme          = "http://"
	ServerAddress   = "localhost:8080"
	BaseURL         = "localhost:8080"
	FileStoragePath = "/tmp/short-url-db.json"

	EnvServerAddressName   = "SERVER_ADDRESS"
	EnvBaseURLName         = "BASE_URL"
	EnvFileStoragePathName = "FILE_STORAGE_PATH"
	EnvNameDBDSN           = "DATABASE_DSN"

	ShortLen = 8

	APIRoute     = "/api"
	ShortenRoute = "/shorten"
	BatchRoute   = "/batch"

	DBTableName = "shortener"
)
