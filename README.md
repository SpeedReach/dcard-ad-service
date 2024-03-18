# Dcard Backend Assignment

This is a simple backend project for Dcard backend assignment.
Spec: https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view  
The api implementation is exactly like the spec.





## Build
Uses makefile to build and test the project
```
make all => build to build dir
make test => run unit tests
make test_all => run all tests including integration tests, requires env vars to be set
```

## Environment Variables
- POSTGRES_URI: postgres connection string
- REDIS_URI: redis connection string


## Directory Structure
```
.
├── cmd
│   └── main.go => main application
├── go.mod
├── go.sum
├── internal
│   ├── handlers => http handlers
│   ├── models => domain models
│   ├── infra => infrastructure code
│     ├── cache => redis cache
│     ├── persistent => postgres db
├── migrations => sql code
```

## Design

這次選擇relational database的原因主要是想練習一下，不然我認為nosql在這情況下開發更為方便快速。
資料庫是使用postgresql， cache是使用redis。
cache 的部分主要用在 get active ads 的時候，由於該api 會被大量呼叫，所以需要cache來加速查詢。
尤其同時間active的ad數量不會超過1000筆，非常適合拿來cache。選擇使用redis而不是in mem cache的原因是，stateless的server更容易scale，若單個server的效能無法達到需求，可以簡單的增加server數量。  
cache 的方式是cache-aside，會先去查詢redis中上一次更新active ad的時間，如果超過一個小時，就會去postgres中查詢start time < now + 80min 的所有ad並更新redis。
設定一個小時是因為我認為廣告active的時間應該不會小於一個小時，所以不會發生一次將所有ad都load進cache的情況。當然這個時間是可以再調整的。  
