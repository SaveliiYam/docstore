# docstore — порты для документоориентированных БД

`Client`, `Database`, `Collection`, `Cursor`, `Transactor`, `ErrNotFound`, `ErrDuplicateKey`.

Бизнес-логика зависит только от этого пакета. Реализации:

- [`mongostore`](https://github.com/SaveliiYam/mongostore) — MongoDB (`mongo-driver`)
- [`lungostore`](https://github.com/SaveliiYam/lungostore) — embeddable Lungo

```go
coll := client.Database("app").Collection("users")
err := coll.FindOne(ctx, bson.M{"_id": id}, &user)
```

Транзакции:

```go
err := client.WithinTx(ctx, func(ctx context.Context) error {
    _, err := coll.InsertOne(ctx, doc)
    return err
})
```

## Установка

```bash
go get github.com/SaveliiYam/docstore@latest
```

Требуется Go ≥ 1.25.
