package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	
	"sync"
	"time"

	"github.com/tamir-liebermann/gobank/env"
	"github.com/tamir-liebermann/gobank/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BankAccount represents a bank account.
// swagger:model
type BankAccount struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	AccountHolder string             `bson:"account_holder"`
	Balance       float64            `bson:"balance"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
	Password      string             `bson:"password"`
	PhoneNumber   string 			 `bson:"phone_number"`
}

type Transaction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FromAccount primitive.ObjectID `bson:"from_account" json:"from_account"`
	ToAccount   primitive.ObjectID `bson:"to_account" json:"to_account"`
	Amount      float64            `bson:"amount" json:"amount"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
}

type AccManager struct {
	client       *mongo.Client
	transactions *mongo.Collection
	accounts     *mongo.Collection
}

func InitDB() (*AccManager, error) {
	spec := env.New()
	 mgr, err := NewManager(spec.MongoSecret)
	// mgr, err := NewManager("mongodb://localhost:27017")
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
		
	db := singletonClient.Database("banktest")	
	return &AccManager{
		client:  singletonClient,
		transactions: db.Collection("transactions"),
		accounts: db.Collection("accs"),
	}, nil
}


func (m *AccManager) CreateAccount(name string, password  string, balance float64 , phoneNumber string) (string, error) {

	hashedPw,err := utils.HashPassword(password)
	if err !=nil{
		return "", err
	}
	account := BankAccount{
		AccountHolder: name,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Balance:       balance,
		Password:      hashedPw ,
		PhoneNumber:   phoneNumber,
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

	deleteResult, err := acc.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %v document(s)\n", deleteResult.DeletedCount)
	return nil
}

func (m *AccManager) DeleteAccountById(id primitive.ObjectID) error {
	filter := bson.D{{Key: "_id", Value: id}}
	acc := m.client.Database("banktest").Collection("accs")

	deleteResult, err := acc.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	if deleteResult.DeletedCount == 0 {
		return fmt.Errorf("no documents found with id: %v", id)
	}

	fmt.Printf("Deleted %v document(s)\n", deleteResult.DeletedCount)
	return nil
}

func (m *AccManager) SearchAccountByNameOrPhone(query string) ([]*BankAccount, error) {
    regexPattern := primitive.Regex{Pattern: query, Options: "i"}		
	filter := bson.M{
        "$or": []bson.M{
            {"account_holder": regexPattern},
            {"phone_number": regexPattern},
        },
    }   
	 var accounts []*BankAccount

    acc := m.client.Database("banktest").Collection("accs")
    cursor, err := acc.Find(context.TODO(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.TODO())

    for cursor.Next(context.TODO()) {
        var account BankAccount
        if err := cursor.Decode(&account); err != nil {
            return nil, err
        }
        accounts = append(accounts, &account)
    }

    if err := cursor.Err(); err != nil {
        return nil, err
    }

    return accounts, nil
}

func (m *AccManager) SearchAccountById(id primitive.ObjectID) (*BankAccount, error) {
	filter := bson.M{"_id": id}
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

func (m *AccManager) GetAccounts() ([]BankAccount, error) {
	var accounts []BankAccount
	collection := m.client.Database("banktest").Collection("accs")

	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var account BankAccount
		err := cursor.Decode(&account)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (m *AccManager) TransferAmountById(fromAccountId, toAccountId primitive.ObjectID, amount float64) error {
	session, err := m.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.TODO())

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		collection := m.accounts

		// Find the from account and ensure sufficient funds
		var fromAccount BankAccount
		err := collection.FindOne(sessCtx, bson.M{"_id": fromAccountId}).Decode(&fromAccount)
		if err != nil {
			return nil, err
		}

		if fromAccount.Balance < amount {
			return nil, errors.New("insufficient funds")
		}

		// Find the to account
		var toAccount BankAccount
		err = collection.FindOne(sessCtx, bson.M{"_id": toAccountId}).Decode(&toAccount)
		if err != nil {
			return nil, err
		}

		// Perform the transfer
		fromAccount.Balance -= amount
		toAccount.Balance += amount

		// Update the from account
		_, err = collection.UpdateOne(
			sessCtx,
			bson.M{"_id": fromAccountId},
			bson.M{"$set": bson.M{"balance": fromAccount.Balance, "updated_at": time.Now()}},
		)
		if err != nil {
			return nil, err
		}

		// Update the to account
		_, err = collection.UpdateOne(
			sessCtx,
			bson.M{"_id": toAccountId},
			bson.M{"$set": bson.M{"balance": toAccount.Balance, "updated_at": time.Now()}},
		)
		if err != nil {
			return nil, err
		}

		// Record the transaction
		transaction := Transaction{
			FromAccount: fromAccountId,
			ToAccount:   toAccountId,
			Amount:      amount,
			Timestamp:   time.Now(),
		}
		_, err = m.transactions.InsertOne(sessCtx, transaction)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	_, err = session.WithTransaction(context.Background(), callback)
	if err != nil {
		return err
	}

	return nil
}

func (m *AccManager) GetTransactionsHistory(accountId primitive.ObjectID) ([]Transaction, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"from_account": accountId},
			{"to_account": accountId},
		},
	}

	cursor, err := m.transactions.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var transactions []Transaction
	for cursor.Next(context.TODO()) {
		var transaction Transaction
		err := cursor.Decode(&transaction)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}


func (m *AccManager) DepositToAccount(amount float64, accountId primitive.ObjectID) error {
    // Ensure the amount is positive
    if amount <= 0 {
        return errors.New("deposit amount must be greater than zero")
    }

    // Define the filter and update
    filter := bson.M{"_id": accountId}
    update := bson.M{"$inc": bson.M{"balance": amount}}

    // Get the collection
    accCollection := m.client.Database("banktest").Collection("accs")

    // Perform the update
    result, err := accCollection.UpdateOne(context.TODO(), filter, update)
    if err != nil {
        return err
    }

    // Check if any document was modified
    if result.ModifiedCount == 0 {
        return errors.New("account not found or no update made")
    }

    return nil
}

func (m *AccManager) GetAccountBalance(accountId primitive.ObjectID) (float64, error) {
    filter := bson.M{"_id": accountId}
    var account BankAccount
    acc := m.client.Database("banktest").Collection("accs")
    err := acc.FindOne(context.TODO(), filter).Decode(&account)
    if errors.Is(err, mongo.ErrNoDocuments) {
        return 0, fmt.Errorf("account not found")
    } else if err != nil {
        return 0, err
    }

    return account.Balance, nil
}





