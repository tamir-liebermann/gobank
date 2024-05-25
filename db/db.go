package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tamir-liebermann/gobank/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BankAccount struct {
	
	AccountHolder string    `bson:"account_holder"`
	Balance       float64   `bson:"balance"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
	Password      string    `bson:"password"`
}

type AccManager struct {
	client *mongo.Client
}

func InitDB() (*AccManager, error) {
	mgr, err := NewManager("mongodb+srv://tamirlieb2:tamir147@cluster0.im9fd2n.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDB client created and connection verified!")

	return mgr, nil

}

var singletonClient *mongo.Client
var once sync.Once

func NewManager(uri string) (*AccManager, error) {
	clientOptions := options.Client().ApplyURI(uri)
	once.Do(func() {

		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			panic(err)
		}

		err = client.Ping(context.TODO(), nil)
		if err != nil {
			panic(err)
		}

		singletonClient = client
		
		fmt.Println("singleton init")
	})
		

	return &AccManager{
		client:  singletonClient,
	}, nil
}
func (m *AccManager) CreateAccount(name string, password string) (string,error) {
	
	hashedPw,err := utils.HashPassword(password)
	if err !=nil{
		return "", err
	}
	account := BankAccount{
		AccountHolder: name,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Balance:       0,
		Password: hashedPw ,
	}

	acc := m.client.Database("banktest").Collection("accs")
	insertResult, err := acc.InsertOne(context.TODO(), account)
	if err != nil {
		return "", err
	}
	fmt.Printf("Inserted a single document: %v\n", insertResult.InsertedID)
	oid,ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok{
		panic("mongo severe error!")
	}
	return oid.Hex(),nil
}

func (m *AccManager) DeleteAccount(accountNumber string) error {
	id, err := primitive.ObjectIDFromHex(accountNumber)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: id}}
		acc := m.client.Database("banktest").Collection("accs")

	deleteResult, err :=acc.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %v document(s)\n", deleteResult.DeletedCount)
	return nil
}

func (m *AccManager) SearchAccountByName(client *mongo.Client, name string) (*BankAccount, error) {
	filter := bson.D{{Key: "account_holder", Value: name}}
	var account BankAccount
	acc := m.client.Database("banktest").Collection("accs")

	err := acc.FindOne(context.TODO(), filter).Decode(&account)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &account, nil
}
