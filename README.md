# server

Go packages you will need
```
go get gopkg.in/yaml.v2
go get github.com/gorilla/mux
go get github.com/dgrijalva/jwt-go
go get github.com/auth0/go-jwt-middleware
go get github.com/mongodb/mongo-go-driver/bson
go get github.com/mongodb/mongo-go-driver/mongo
```

You will also need mongo installed on your machine.


Start mongo (if not already started)
```
mongod
```

If its your first time youll also need to create the database (and fill it)

1. Open a mongo shell

```
mongo
```
2. Create DB and Collections
```
use ShopTrac

db.createCollection("users")
```


Starting
```
go build
./server
