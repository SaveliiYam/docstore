// Package docstore — порты (интерфейсы) для документоориентированных БД.
//
// Ключевой принцип заменяемости: бизнес-логика и репозитории зависят только
// от Client / Database / Collection / Cursor из этого пакета, а не от
// mongo-driver или lungo. Адаптеры: mongostore, lungostore.
//
//	coll := client.Database("app").Collection("users")
//	err := coll.FindOne(ctx, bson.M{"email": email}, &u)
//
// Транзакции — через единый Transactor (как в github.com/SaveliiYam/database):
//
//	err := client.WithinTx(ctx, func(ctx context.Context) error {
//		return coll.InsertOne(ctx, doc)
//	})
//
// Фильтры/апдейты типизированы как any — обычно bson.M / bson.D из
// go.mongodb.org/mongo-driver/bson (общий формат для mongo и lungo).
package docstore

import (
	"context"
	"errors"
)

var (
	ErrNotFound     = errors.New("docstore: not found")
	ErrDuplicateKey = errors.New("docstore: duplicate key")
)

func IsNotFound(err error) bool     { return errors.Is(err, ErrNotFound) }
func IsDuplicateKey(err error) bool { return errors.Is(err, ErrDuplicateKey) }

// Client — точка входа: базы, ping, shutdown, транзакции.
type Client interface {
	Database(name string) Database
	Ping(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Transactor
}

// Database — именованная БД и фабрика коллекций.
type Database interface {
	Name() string
	Collection(name string) Collection
	Drop(ctx context.Context) error
	ListCollectionNames(ctx context.Context) ([]string, error)
}

// Collection — CRUD над документами.
type Collection interface {
	Name() string

	InsertOne(ctx context.Context, document any) (*InsertOneResult, error)
	InsertMany(ctx context.Context, documents []any) (*InsertManyResult, error)

	FindOne(ctx context.Context, filter any, dest any, opts ...*FindOneOptions) error
	Find(ctx context.Context, filter any, opts ...*FindOptions) (Cursor, error)

	UpdateOne(ctx context.Context, filter, update any, opts ...*UpdateOptions) (*UpdateResult, error)
	UpdateMany(ctx context.Context, filter, update any, opts ...*UpdateOptions) (*UpdateResult, error)
	ReplaceOne(ctx context.Context, filter, replacement any, opts ...*ReplaceOptions) (*UpdateResult, error)

	DeleteOne(ctx context.Context, filter any) (*DeleteResult, error)
	DeleteMany(ctx context.Context, filter any) (*DeleteResult, error)

	CountDocuments(ctx context.Context, filter any, opts ...*CountOptions) (int64, error)
	Drop(ctx context.Context) error
}

// Cursor — итерация по результату Find.
type Cursor interface {
	Next(ctx context.Context) bool
	Decode(dest any) error
	All(ctx context.Context, results any) error
	Close(ctx context.Context) error
	Err() error
}

// Transactor — единый паттерн транзакций (сессии mongo / lungo).
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// HealthChecker / Closer — для app.Component / database.Component-подобных обёрток.
type HealthChecker interface {
	Ping(ctx context.Context) error
}

type Closer interface {
	Disconnect(ctx context.Context) error
}

// DB — удобный агрегат для DI.
type DB interface {
	Client
	HealthChecker
	Closer
}

// --- results ---

type InsertOneResult struct {
	InsertedID any
}

type InsertManyResult struct {
	InsertedIDs []any
}

type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    any
}

type DeleteResult struct {
	DeletedCount int64
}

// --- options (без зависимости от mongo options) ---

type FindOptions struct {
	Limit      *int64
	Skip       *int64
	Sort       any
	Projection any
}

type FindOneOptions struct {
	Sort       any
	Projection any
}

type UpdateOptions struct {
	Upsert *bool
}

type ReplaceOptions struct {
	Upsert *bool
}

type CountOptions struct {
	Limit *int64
	Skip  *int64
}

func Limit(n int64) *int64  { return &n }
func Skip(n int64) *int64   { return &n }
func Upsert(v bool) *bool   { return &v }

// Component — лёгкая обёртка под app.Component (Name/Start/Stop).
type Component struct {
	name   string
	client Client
}

func NewComponent(name string, client Client) *Component {
	return &Component{name: name, client: client}
}

func (c *Component) Name() string { return c.name }

func (c *Component) Start(ctx context.Context) error {
	if err := c.client.Ping(ctx); err != nil {
		return err
	}
	<-ctx.Done()
	return nil
}

func (c *Component) Stop(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
