# How to use it

Write:

```go
db := New("path/to/data")

err := db.Save("log", "something")
```
	
Read:

```go
db := New("path/to/data")

scanner := db.Query("logs", time.Now(), time.Now())

for scanner.Scan() {
	datapoint := scanner.Data()
}	
```
