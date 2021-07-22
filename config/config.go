package config

type Properties struct {
	AppPort             string `env:"APP_PORT" env-default:"8080"`
	AppHost             string `env:"APP_HOST" env-default:"localhost"`
	DBPort              string `env:"DB_PORT" env-default:"27017"`
	DBHost              string `env:"DB_HOST" env-default:"localhost"`
	DBName              string `env:"DB_NAME" env-default:"tronics"`
	DBCollection        string `env:"DB_COLLECTION" env-default:"products"`
	DBProductCollection string `env:"DB_PRODUCTS_COLLECTION" env-default:"products"`
	DBUsersCollection   string `env:"DB_USERS_COLLECTION" env-default:"users"`
	JwtSecret           string `env:"JWT_SECRET" env-default:"secretsecretsecret"`
}
