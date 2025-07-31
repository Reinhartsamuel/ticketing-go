package connection

import (
	"database/sql"
	"fmt"
	"log"
	"ticketing-go/internal/config"

	_ "github.com/lib/pq"
)

func GetDatabase(conf config.Database) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable Timezone=%s",
		conf.Host,
		conf.Port,
		conf.User,
		conf.Pass,
		conf.Name,
		conf.Tz,
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("error opening connection to database:", err.Error())
	}
	err = db.Ping()
	if err != nil {
		log.Fatal("error pinging database:", err.Error())
	}

	return db
}
