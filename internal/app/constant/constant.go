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
	EnvNameKey             = "KEY"

	ShortLen = 8

	APIRoute     = "/api"
	ShortenRoute = "/shorten"
	BatchRoute   = "/batch"
	UserRoute    = "/user"
	URLsRoute    = "/urls"

	DBTableName      = "shortener"
	DBUsersTableName = "users"

	CookieAuthName              = "AuthShortener"
	ContextUserValueName CtxKey = "userID"
)

type CtxKey string

func (c CtxKey) String() string {
	return string(c)
}
