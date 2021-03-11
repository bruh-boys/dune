# SQL Query Builder

```go
q, err := sqx.Parse("select * from foo where id = ?")
if err != nil {
    log.Fatal(err)
}
```