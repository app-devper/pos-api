package db

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"strconv"
	"time"
)

type Resource struct {
	Client *mongo.Client
	PosDb  *mongo.Database
	RdDb   *redis.Client
}

// Close use this method to close database connection
func (r *Resource) Close() {
	logrus.Warning("Closing all db connections")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if r.Client != nil {
		if err := r.Client.Disconnect(ctx); err != nil {
			logrus.WithError(err).Warn("disconnect mongo failed")
		}
	}
	if r.RdDb != nil {
		if err := r.RdDb.Close(); err != nil {
			logrus.WithError(err).Warn("close redis failed")
		}
	}
}

func InitResource() (*Resource, error) {
	if !isCloudRunRuntime() {
		err := godotenv.Load(".env")
		if err != nil {
			log.Print(err)
		}
	}

	host := os.Getenv("MONGO_HOST")
	posDbName := os.Getenv("MONGO_POS_DB_NAME")
	if host == "" {
		return nil, fmt.Errorf("missing MONGO_HOST")
	}
	if posDbName == "" {
		return nil, fmt.Errorf("missing MONGO_POS_DB_NAME")
	}
	mongoOptions := options.Client().
		ApplyURI(host).
		SetMaxPoolSize(getEnvUint64("MONGO_MAX_POOL_SIZE", 10)).
		SetMinPoolSize(getEnvUint64("MONGO_MIN_POOL_SIZE", 0)).
		SetMaxConnIdleTime(getEnvDuration("MONGO_MAX_CONN_IDLE_TIME_SEC", 120)).
		SetConnectTimeout(getEnvDuration("MONGO_CONNECT_TIMEOUT_SEC", 3)).
		SetServerSelectionTimeout(getEnvDuration("MONGO_SERVER_SELECTION_TIMEOUT_SEC", 3))

	mongoClient, err := mongo.NewClient(mongoOptions)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), getEnvDuration("MONGO_INIT_TIMEOUT_SEC", 5))
	defer cancel()
	err = mongoClient.Connect(ctx)
	if err != nil {
		return nil, err
	}
	if err = mongoClient.Ping(ctx, nil); err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		_ = mongoClient.Disconnect(context.Background())
		return nil, fmt.Errorf("missing REDIS_HOST")
	}
	redisOp, err := redis.ParseURL(redisHost)
	if err != nil {
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}
	if redisOp.PoolSize == 0 {
		redisOp.PoolSize = getEnvInt("REDIS_POOL_SIZE", 10)
	}
	if redisOp.MinIdleConns == 0 {
		redisOp.MinIdleConns = getEnvInt("REDIS_MIN_IDLE_CONNS", 0)
	}
	if redisOp.DialTimeout == 0 {
		redisOp.DialTimeout = getEnvDuration("REDIS_DIAL_TIMEOUT_SEC", 3)
	}
	if redisOp.ReadTimeout == 0 {
		redisOp.ReadTimeout = getEnvDuration("REDIS_READ_TIMEOUT_SEC", 3)
	}
	if redisOp.WriteTimeout == 0 {
		redisOp.WriteTimeout = getEnvDuration("REDIS_WRITE_TIMEOUT_SEC", 3)
	}
	if redisOp.PoolTimeout == 0 {
		redisOp.PoolTimeout = getEnvDuration("REDIS_POOL_TIMEOUT_SEC", 4)
	}
	if redisOp.IdleTimeout == 0 {
		redisOp.IdleTimeout = getEnvDuration("REDIS_IDLE_TIMEOUT_SEC", 120)
	}
	rdb := redis.NewClient(redisOp)
	if err = rdb.Ping(context.Background()).Err(); err != nil {
		_ = rdb.Close()
		_ = mongoClient.Disconnect(context.Background())
		return nil, err
	}

	return &Resource{
		Client: mongoClient,
		PosDb:  mongoClient.Database(posDbName),
		RdDb:   rdb,
	}, nil
}

func isCloudRunRuntime() bool {
	return os.Getenv("K_SERVICE") != "" || os.Getenv("K_REVISION") != ""
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	result, err := strconv.Atoi(value)
	if err != nil || result < 0 {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Warn("invalid int env, using fallback")
		return fallback
	}
	return result
}

func getEnvUint64(key string, fallback uint64) uint64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"key":   key,
			"value": value,
		}).Warn("invalid uint env, using fallback")
		return fallback
	}
	return result
}

func getEnvDuration(key string, fallbackSeconds int) time.Duration {
	return time.Duration(getEnvInt(key, fallbackSeconds)) * time.Second
}
