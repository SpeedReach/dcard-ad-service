# Dcard Backend Assignment

This is a simple backend project for Dcard backend assignment.
Spec: https://drive.google.com/file/d/1dnDiBDen7FrzOAJdKZMDJg479IC77_zT/view  
Api differs from the spec, see Design section



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
### Data Storage
這次選擇relational database的原因主要是想練習一下，不然我認為nosql在這情況下開發更為方便快速。  
資料庫是使用postgresql， cache是使用redis。  
cache 的部分主要用在 get active ads 的時候，由於該api 會被大量呼叫，所以需要cache來加速查詢。  
尤其同時間active的ad數量不會超過1000筆，非常適合拿來cache。選擇使用redis而不是in mem cache的原因是，stateless的server更容易scale，若單個server的效能無法達到需求，可以簡單的增加server數量。   
cache 的方式是cache-aside，會先去查詢redis中上一次更新active ad的時間，如果超過一個小時，就會去postgres中查詢(start time < now + 80min) && now < end time 的所有ad並更新redis。  
設定一個小時是因為我認為廣告active的時間應該不會小於一個小時，所以不會發生一次將所有ad都load進cache的情況。當然這個時間是可以再調整的。   

### Api
API 跟作業中的說明文件一模一樣，只差在 get ads 的時候，會多回傳一個參數 `end`，來告訴前端是否還有更多的active ad 等著他去retrieve。
至於為甚麼要多這一個欄位，我們得先考慮原本的設計:
前端傳入 offset & limit ，期待得到limit 個ad，若 ad 數量小於limit，則判斷active ads已經沒了。 
這樣做在第一次get的時候是可以正常運作的。  
問題在於前端嘗試取得第二個page的時候，他並不知道第一次的page有幾個ads已經被filter過了，所以會有非常高的機率讓後端去做重複的多於運算。
甚至更極端的情況下，如果前端要5個ad，而剛好 offset 後的5個ad條件都不符合，那前端就會以為已經沒有更多ad了。
所以此api改成讓前端透過 end 欄位判斷有沒有更多ad，前端不保證獲得limit個ad，必須透過loop的方式重複獲取。
### Libraries
http server 使用golang 內建，無使用框架
sql 使用 golang 內建的 sql package 搭配 pgx driver，用純sql的方式寫，不使用orm