package ratelimiter

import (
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Log struct {
	gorm.Model
	Identifier int
	Timestamp  int
	Parameter  string
}

var (
	db *gorm.DB
)

func init() {
	dbSession, err := gorm.Open(sqlite.Open("log.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db = dbSession

	// Migrate the schema
	db.AutoMigrate(&Log{})
}

type RateLimiter struct {
	Period   int
	MaxCalls int
}

func (r *RateLimiter) CheckLimit(identifier int, parameter string) bool {
	timeframe := int(time.Now().Unix()) - r.Period
	var result int64
	db.Table("logs").Where("Identifier = ? AND Parameter = ? AND Timestamp > ?", identifier, parameter, timeframe).Count(&result)
	if int(result) >= r.MaxCalls {
		return true
	} else {
		r.AddLog(identifier, parameter)
		return false
	}
}

func (r *RateLimiter) AddLog(identifier int, parameter string) {
	timestamp := time.Now().Unix()
	db.Create(&Log{Identifier: identifier, Timestamp: int(timestamp), Parameter: parameter})
}

// def check_limit(self, identifier):
//     """The __call__ function allows the RateLimiter object to be used as a
//     regular function decorator.
//     """
//     logging.info(f"Function is called {identifier}")
//     current_time = time.time()

//     print(db.get_count(twitter_user_id=identifier, period=self.period))
//     if db.get_count(twitter_user_id=identifier, period=self.period) >= self.max_calls:
//         raise RateLimitExceeded

//     print(identifier)
//     lock.acquire()
//     db.add_log(identifier, current_time)
//     lock.release()
