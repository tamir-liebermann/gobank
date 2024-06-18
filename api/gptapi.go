package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tamir-liebermann/gobank/db"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sashabaranov/go-openai"
)

const (
	TRANSFER_INTENT = "transfer"
	FIND_ACCOUNT_INTENT= "find this account"
	TRANSACTIONS_INTENT= "transactions"
	SEARCH_INTENT="search"
	DEPOSIT_INTENT= "deposit"
)

// type AgentTransferRequest struct {
// 	Intent string          `json:"intent"`
// 	Body   TransferRequest `json:"body"`
// }

type GenericRequst struct {
	Intent string          `json:"intent"`
	Body   map[string]interface{} `json:"body"`
}

var urlMap = map[string]string{
	"create account":                "http://localhost:8080/create",
	"login":                         "http://localhost:8080/login",
	"get account by ID":             "http://localhost:8080/account/:id",
	"get account by name":           "http://localhost:8080/account/name/:account_holder",
	"delete account by ID":          "http://localhost:8080/account/:id",
	"transfer funds":                "http://localhost:8080/account/transfer/:id",
	"get transactions history":      "http://localhost:8080/account/transactions/:id",
	"get all accounts (admin only)": "http://localhost:8080/admin/accounts",
}

func processGPTResponse(gptResponse *GPTResponse) string {
	if len(gptResponse.Choices) > 0 {
		return gptResponse.Choices[0].Text
	}
	return ""
}

func (api *ApiManager) handleChatGPTRequest(ctx *gin.Context) {
	var chatReq ChatReq
	if err := ctx.ShouldBindJSON(&chatReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userInput := strings.ToLower(strings.TrimSpace(chatReq.UserText))

	client := openai.NewClient("sk-proj-4uI52DxoSVvkMSlx7vAqT3BlbkFJdk9t7YDPID1Ome55jWCL")

	rules := `
		You are a bank API, you reply in json objects only, if unsure ask for clarification.

		If the user wants to transfer to an account id, give them:
		{
			"intent": "transfer", // must be this keyword
			"body":{
				from:"string", // may leave empty string
				to:"string", // must be the id
				amount:"float" // must be specified
			}
			
		}

		If the user wants to search for another account by id, give them: 
		{
			"intent": "find this account", // must be this keyword
			"body": {
				"_id":"string", // must be the id 
			}
		}

		If the user wants to see his transactions history , give them:
		{
			"intent": "transactions", // must be this keyword 
			"body": {
				"_id": "string", // must be the id 
			}
		}

		If the user wants to search for account by his name or his phone   , give them:
		{
			"intent": "search", // must be this keyword
			"body": {
				"account_holder": "string", // must be the account holder
				"phone_number": "string", // must be the account's phone number
			}
		}

		If the user wants to deposit money to his account , give them:
		{
			"intent": "deposit", // must be this keyword
			"body": {
				"_id": "string", // must be the id 
				"amount": "float" // must be specified
			}
		}
	`

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: rules,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userInput,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	textResp := resp.Choices[0].Message.Content

	var req GenericRequst
	err = json.Unmarshal([]byte(textResp), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
	}

	accountId := ctx.GetString("userId")
	// todo use transfer req

	switch req.Intent {
	case TRANSFER_INTENT:
		bodyBytes,err  := json.Marshal(req.Body)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}


		var transferReq TransferRequest
		err = json.Unmarshal(bodyBytes, &transferReq)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}
		
		err = api.handleTransferIntent(accountId,transferReq.To,transferReq.Amount)
		if err != nil{
					ctx.JSON(http.StatusInternalServerError, gin.H{"message": "servererror"})

		}
	case FIND_ACCOUNT_INTENT:
		bodyBytes,err  := json.Marshal(req.Body)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
	}

		var idRequest  IdRequest
		err = json.Unmarshal(bodyBytes, &idRequest)
		if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}

		account, err := api.handleFindAccountIntent(idRequest.Id.Hex())
		if err != nil {
   		 ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    	return
	}
	ctx.JSON(http.StatusOK, gin.H{"account found": account})
	
	case TRANSACTIONS_INTENT:
		bodyBytes,err  := json.Marshal(req.Body)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
	}
	
	var transactionsRequest TransactionsReq
	err = json.Unmarshal(bodyBytes, &transactionsRequest)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}

	history, err := api.handleTransactionsIntent(transactionsRequest.AccountID)
	if err != nil {
   		 ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    	return
}
	ctx.JSON(http.StatusOK, gin.H{"Transaction history found": history})

 	case SEARCH_INTENT: 
		bodyBytes,err  := json.Marshal(req.Body)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
	}

	var accNameReq AccNameReq
		err = json.Unmarshal(bodyBytes, &accNameReq)
		if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}
		account, err := api.handleSearchAccountByNameIntent(accNameReq.AccountHolder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
    		return
		}
			ctx.JSON(http.StatusOK, gin.H{"Account   found": account})

	case DEPOSIT_INTENT:
		bodyBytes,err  := json.Marshal(req.Body)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
	}

	var depositReq DepositRequest
	    err = json.Unmarshal(bodyBytes, &depositReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		}

		err = api.handleDepositIntent(depositReq.AccountID, depositReq.Amount)
		if err != nil{
					ctx.JSON(http.StatusInternalServerError, gin.H{"message": "servererror"})

		}
}	
}
func (api *ApiManager) handleTransactionsIntent(id string ) (*[]db.Transaction ,error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID format: %v", err)
	}

	transactions, err := api.accMgr.GetTransactionsHistory(objectID)
	if err != nil {
		return nil, fmt.Errorf("error fetching transactions: %v", err)
	}

	return &transactions , nil 


}

func (api *ApiManager) handleTransferIntent(from, to string, amount float64) error {
	fromAccountID, err := primitive.ObjectIDFromHex(from)
	if err != nil {
		return err
	}
	toAccountID, err := primitive.ObjectIDFromHex(to)
	if err != nil {
		return err
	}

	err = api.accMgr.TransferAmountById(fromAccountID, toAccountID, amount)
	if err != nil {
		return err
	}

	return nil

}

func (api *ApiManager) handleSearchAccountByNameIntent(name string) ([]*db.BankAccount, error) {
    // Call the updated SearchAccountByNameOrPhone function
    accounts, err := api.accMgr.SearchAccountByNameOrPhone(name)
    if err != nil {
        return nil, fmt.Errorf("error searching for account: %v", err)
    }

    // Check if no accounts were found
    if len(accounts) == 0 {
        return nil, fmt.Errorf("no accounts found matching the provided name or phone number")
    }

    return accounts, nil
}

func (api *ApiManager) handleFindAccountIntent(id string) (*db.BankAccount, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID format: %v", err)
	}

	account, err := api.accMgr.SearchAccountById(objectID)
	if err != nil {
		return nil, fmt.Errorf("error searching for account: %v", err)
	}

	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	return account, nil
}

func (api *ApiManager) handleDepositIntent(id string , amount float64) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return  fmt.Errorf("invalid account ID format: %v", err)
	}
	if amount <= 0 {
        return fmt.Errorf("amount must be greater than zero")
    }

    // Perform the deposit operation
    err = api.accMgr.DepositToAccount(amount, objectID)
    if err != nil {
        return fmt.Errorf("error depositing to account: %v", err)
    }

    return nil


	
}

// func GetAccount(token string, body interface{}) {
// 	gptResp := "getAccount/user123"
// 	args := strings.Split(gptResp, "/")

// 	url := args[0]

// 	if url == "getAccount" {
// 		req, err := http.NewRequest("GET", "gobank/account/name/"+args[1], nil)
// 		if err != nil {
// 			panic(err)
// 		}

// 		req.Header.Set("Authorization", "Bearer "+token)

// 		client := openai.NewClient("your_openai_api_key")

// 		rules := `
// 	You are a bank API.

// 	If the user wants to get account information, give them:
// 	{
// 		"url": "string",
// 		"body": "object"
// 	}

// 	For getting account information, the URL is "gobank/account/name/{account_holder}" and the body is empty.
// 	`

// 		resp, err := client.CreateChatCompletion(
// 			context.Background(),
// 			openai.ChatCompletionRequest{
// 				Model: openai.GPT3Dot5Turbo,
// 				Messages: []openai.ChatCompletionMessage{
// 					{
// 						Role:    openai.ChatMessageRoleSystem,
// 						Content: rules,
// 					},
// 					{
// 						Role:    openai.ChatMessageRoleUser,
// 						Content: "I want to get account information for user123",
// 					},
// 				},
// 			},
// 		)
// 		if err != nil {
// 			fmt.Printf("ChatCompletion error: %v\n", err)
// 			return
// 		}

// 	}
// }
