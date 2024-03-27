# DCard Backend Assignment

This is a simple backend project for DCard backend assignment.
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
- AUTO_MIGRATION: creates table on start, (true, false)


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
### Api
API 跟作業中的說明文件一模一樣，只差在 get ads 的時候，會多回傳一個參數 `end`，來告訴前端是否還有更多的active ad 等著他去retrieve。
至於為甚麼要多這一個欄位，我們得先考慮原本的設計:
前端傳入 offset & limit ，期待得到limit 個ad，若 ad 數量小於limit，則判斷active ads已經沒了。
這樣做在第一次get的時候是可以正常運作的。  
問題在於前端嘗試取得第二個page的時候，他並不知道第一次的page有幾個ads已經被filter過了，所以會有非常高的機率讓後端去做重複的多餘運算。
所以此api改成讓前端透過 end 欄位判斷有沒有更多ad，不過此設計下前端不保證獲得limit個ad，所以必須透過loop的方式重複獲取。

### Data Storage
這次選擇relational database的原因主要是想練習一下，不然我認為nosql在這情況下開發更為方便快速。  
資料庫是使用postgresql， cache是使用redis。  
cache 的部分主要用在 get active ads 的時候，由於該api 會被大量呼叫，所以需要cache來加速查詢。  
尤其同時間active的ad數量不會超過1000筆，非常適合拿來cache。選擇使用redis而不是in mem cache的原因是，stateless的server更容易scale，若單個server的效能無法達到需求，可以簡單的增加server數量。   
cache 的方式是cache-aside，會先去查詢redis中上一次更新active ad的時間，如果超過一個小時，就會去postgres中查詢(start time < (now + cache.Interval+ cache.Tolerance)) && now < end time 的所有ad並更新redis。
這邊會發現，cache 中存的是現在active 與未來80分鐘內會active的所有 ad，比較有可能會出現問題的地方是如果active ad的active時間非常短，雖然同時不會超過1000筆active，但一小時內可能有上萬筆active ad。  
不過我推測ad的active時間應該不會太短，所以這部分是不太會出問題的，如果需要調整的話可以將cache.Interval的時間調短。  
更新的步驟為:
1. try to acquire lock (redis NX)
2. remove expired ads
3. get the largest start_at in cache
4. only insert ads that has start_at larger than the value obtained on 3rd step
5. release lock  

lock為write lock，透過redis的NX功能實作，這些步驟確保一次只會有一個redis client更新cache，
由於有tolerance的部分與redis單線程的設計，其他的client可以繼續正常的獲取active中的ads。
#### Erd

![erd](https://raw.githubusercontent.com/SpeedReach/dcard-ad-service/main/assets/erd.png)  
這邊的erd設計算是有點偷吃步，沒有將gender，country，platform各獨立成一個table。優點是查詢與開發的時候可以更快速與方便，缺點是日後如果要新增更多種condition，這個table會很難scale。不過由於這是assignment，之後並不會有新增更多種condition的需求，所以我認為是可以接受的。

### Libraries
http server 使用golang 內建，無使用框架  
sql 使用 golang 內建的 sql package 搭配 pgx driver，用純sql的方式寫，不使用orm
logging & tracing 的部分使用uber 的 zap套件，每次有新的請求時，會生成一組request id 方便debug跟日誌查詢

### Testing
test 分為 unit test與integration test，unit test主要是針對handler與domain logic，integration test則是針對整個api的行為。
unit test 會使用mock cache與 in memory sqlite， integration test則會使用真實的cloud db與redis，所以需要設定環境參數。

### Time Domain
Uses UTC time across the project, to simplify the process of handling different time zone.
