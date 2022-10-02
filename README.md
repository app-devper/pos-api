# POS
Point of Sale (POS).

# Feature
* CRUD API
* Authentication
* Authorization
* CORS


# Technologies
* [Gin](https://github.com/gin-gonic/gin)
* [MongoDB](https://www.mongodb.com)
* [Redis](https://redis.io)

# Set up
* Create file .env
* Set MongoDB URI and DB
  - PORT = "8586" or your port
  - MONGO_HOST = "your host/ localhost:27017"
  - MONGO_POS_DB_NAME = "your pos db name"
  - REDIS_HOST = "your redis host"
  - CLIENT_ID = "your client id"
  - SYSTEM = "your system"
  - SECRET_KEY = "your secret key"
  - LINE_TOKEN  = "line notify"

# Run
* `go mod download` for download dependencies
* `go run main.go`
* `nodemon --exec go run main.go --signal SIGTERM` for run with nodemon


