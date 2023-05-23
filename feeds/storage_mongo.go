package feeds

import (
	"context"
	"crypto/tls"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"producer-rss/config"
	"time"
)

type storageMongo struct {
	conn *mongo.Client
	db   *mongo.Database
	coll *mongo.Collection
}

type feedRec struct {
	Url        string    `bson:"url"`
	UpdateTime time.Time `bson:"ts"`
}

const attrUrl = "url"
const attrTs = "ts"

var projRead = bson.D{
	{
		Key:   attrTs,
		Value: 1,
	},
}
var optsSrvApi = options.ServerAPI(options.ServerAPIVersion1)
var optsRead = options.
	FindOne().
	SetShowRecordID(false).
	SetProjection(projRead)
var optsUpsert = options.
	Update().
	SetUpsert(true)
var indices = []mongo.IndexModel{
	{
		Keys: bson.D{
			{
				Key:   attrUrl,
				Value: 1,
			},
		},
		Options: options.
			Index().
			SetUnique(true),
	},
}

func NewStorage(ctx context.Context, cfgDb config.DbConfig) (s Storage, err error) {
	clientOpts := options.
		Client().
		ApplyURI(cfgDb.Uri).
		SetServerAPIOptions(optsSrvApi)
	if cfgDb.Tls.Enabled {
		clientOpts = clientOpts.SetTLSConfig(&tls.Config{InsecureSkipVerify: cfgDb.Tls.Insecure})
	}
	if len(cfgDb.UserName) > 0 {
		auth := options.Credential{
			Username:    cfgDb.UserName,
			Password:    cfgDb.Password,
			PasswordSet: len(cfgDb.Password) > 0,
		}
		clientOpts = clientOpts.SetAuth(auth)
	}
	conn, err := mongo.Connect(ctx, clientOpts)
	var sm storageMongo
	if err == nil {
		db := conn.Database(cfgDb.Name)
		coll := db.Collection(cfgDb.Table.Name)
		sm.conn = conn
		sm.db = db
		sm.coll = coll
		_, err = sm.ensureIndices(ctx)
	}
	if err == nil {
		s = sm
	}
	return
}

func (sm storageMongo) ensureIndices(ctx context.Context) ([]string, error) {
	return sm.coll.Indexes().CreateMany(ctx, indices)
}

func (sm storageMongo) Close() error {
	return sm.conn.Disconnect(context.TODO())
}

func (sm storageMongo) GetUpdateTime(ctx context.Context, url string) (t time.Time, err error) {
	q := bson.M{
		attrUrl: url,
	}
	var result *mongo.SingleResult
	result = sm.coll.FindOne(ctx, q, optsRead)
	var rec feedRec
	err = result.Decode(&rec)
	if err == nil {
		t = rec.UpdateTime
	}
	switch {
	case err == mongo.ErrNoDocuments:
		err = nil
	case err != nil:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (sm storageMongo) SetUpdateTime(ctx context.Context, url string, t time.Time) (err error) {
	q := bson.M{
		attrUrl: url,
	}
	u := bson.M{
		"$set": bson.M{
			attrTs: t,
		},
	}
	_, err = sm.coll.UpdateOne(ctx, q, u, optsUpsert)
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}
