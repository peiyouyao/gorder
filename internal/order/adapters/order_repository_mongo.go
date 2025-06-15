package adapters

import (
	"context"
	"time"

	_ "github.com/peiyouyao/gorder/common/config"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/peiyouyao/gorder/order/entity"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepositoryMongo struct {
	db *mongo.Client
}

type orderModel struct {
	MongoID     primitive.ObjectID `bson:"_id"`
	ID          string             `bson:"id"`
	CustomerID  string             `bson:"customer_id"`
	Status      string             `bson:"status"`
	PaymentLink string             `bson:"payment_link"`
	Items       []*entity.Item     `bson:"items"`
}

var (
	dbName   = viper.GetString("mongo.db-name")
	collName = viper.GetString("mongo.coll-name")
)

func NewOrderRepositoryMongo(db *mongo.Client) *OrderRepositoryMongo {
	return &OrderRepositoryMongo{db: db}
}

func (r *OrderRepositoryMongo) Create(ctx context.Context, order *domain.Order) (created *domain.Order, err error) {
	defer r.logWithTag("create", err, created)

	write := r.marshalToModel(order)
	res, err := r.collection().InsertOne(ctx, write)
	if err != nil {
		return
	}
	created = order
	created.ID = res.InsertedID.(primitive.ObjectID).Hex()
	return
}

func (r *OrderRepositoryMongo) Get(ctx context.Context, id, customerID string) (got *domain.Order, err error) {
	defer r.logWithTag("get", err, got)

	read := &orderModel{}
	mongoID, _ := primitive.ObjectIDFromHex(id)

	cond := bson.M{"_id": mongoID}
	if err = r.collection().FindOne(ctx, cond).Decode(read); err != nil {
		return
	}
	got = r.unmarshal(read)
	return
}

func (r *OrderRepositoryMongo) Update(
	ctx context.Context, order *domain.Order,
	updateFn func(context.Context, *domain.Order) (*domain.Order, error),
) (err error) {
	if order == nil {
		panic("nil order")
	}

	session, err := r.db.StartSession()
	if err != nil {
		return
	}
	defer session.EndSession(ctx)

	if err = session.StartTransaction(); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_ = session.CommitTransaction(ctx)
		} else {
			_ = session.AbortTransaction(ctx)
		}
	}()

	oldOrder, err := r.Get(ctx, order.ID, order.CustomerID)
	if err != nil {
		return
	}

	updated, err := updateFn(ctx, order)
	if err != nil {
		return
	}

	mongoID, _ := primitive.ObjectIDFromHex(oldOrder.ID)

	res, err := r.collection().UpdateOne(
		ctx,
		bson.M{"_id": mongoID, "customer_id": oldOrder.CustomerID}, // can't add condition: `"id": mongoID"`, because id need mongoID.Hex()
		bson.M{"$set": bson.M{
			"status":       updated.Status,
			"payment_link": updated.PaymentLink,
		}},
	)
	if err != nil {
		return
	}
	logrus.Info(res)
	r.logWithTag("update", err, res)
	return
}

func (r *OrderRepositoryMongo) collection() *mongo.Collection {
	return r.db.Database(dbName).Collection(collName)
}

func (r *OrderRepositoryMongo) marshalToModel(order *domain.Order) orderModel {
	mongoID := primitive.NewObjectID()
	return orderModel{
		MongoID:     mongoID,
		ID:          mongoID.Hex(),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
}

func (r *OrderRepositoryMongo) unmarshal(read *orderModel) *domain.Order {
	return &domain.Order{
		ID:          read.MongoID.Hex(),
		CustomerID:  read.CustomerID,
		Status:      read.Status,
		PaymentLink: read.PaymentLink,
		Items:       read.Items,
	}
}

func (r *OrderRepositoryMongo) logWithTag(tag string, err error, result interface{}) {
	l := logrus.WithFields(logrus.Fields{
		"timestamp": time.Now().Unix(),
		"err":       err,
		"result":    result,
	})
	if err != nil {
		l.Infof("order_repo_mongo_%s_fail", tag)
	} else {
		l.Infof("order_repo_mongo_%s_success", tag)
	}
}
