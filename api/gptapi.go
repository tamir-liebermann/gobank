package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/env"
	"github.com/tamir-liebermann/gobank/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sashabaranov/go-openai"
)

const (
	TRANSFER_INTENT         = "transfer"
	FIND_ACCOUNT_INTENT     = "find this account"
	TRANSACTIONS_INTENT     = "transactions"
	SEARCH_INTENT           = "search"
	DEPOSIT_INTENT          = "deposit"
	BALANCE_CHECK_INTENT    = "balance"
	GET_ALL_ACCOUNTS_INTENT = "all accounts"
)

// type AgentTransferRequest struct {
// 	Intent string          `json:"intent"`
// 	Body   TransferRequest `json:"body"`
// }

type GenericRequst struct {
	Intent string                 `json:"intent"`
	Body   map[string]interface{} `json:"body"`
}

// @Summary Handle chat request
// @Description Handle a chat request using OpenAI GPT-3.5 model
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_text body string true "User's text"
// @Success 200 {string} string "Response"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /account/chatgpt [post]
func (api *ApiManager) handleChatGPTRequest(ctx *gin.Context) {
	spec := env.New()
	var chatReq ChatReq
	if err := ctx.ShouldBindJSON(&chatReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userInput := strings.ToLower(strings.TrimSpace(chatReq.UserText))

	
	// TODO must handle this . accountId MUST NOT BE EMPTY!!!
	accountId, ok := ctx.Get("userId") //todo get from jwt

	if !ok || accountId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "accountId is required"})
		return
	}
	var acc db.BankAccount
	twilioAcc, ok := ctx.Get(TwilioUser)
	if ok {
		b, err := json.Marshal(twilioAcc)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		json.Unmarshal(b, &acc)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		accountId = acc.ID.Hex()
	}

	client := openai.NewClient(spec.OpenaiApiKey)
	

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
				
				"amount": "float" // must be specified
			}
		}

		If the user wants to check his account balance , give them :
		{
			"intent": "balance", // must be this keyword
			"body": {
				"_id": "string, // must be the id 
				"balance": "float" // must be specified
			}
		}

		If the user is admin and wants to see the all the existing accounts, give them: 
		{
			"intent": "all accounts", // must be this keyword
			"body" :{
				"_id": "string" // must be the id 
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

	// todo use transfer req
	var response string
	switch req.Intent {
	case TRANSFER_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

		}

		var transferReq TransferRequest
		err = json.Unmarshal(bodyBytes, &transferReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

		}
		accountId := fmt.Sprintf("%v", accountId)
		err = api.handleTransferIntent(accountId, transferReq.To, transferReq.Amount)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "servererror"})

		}
		tranferAccId, _ := primitive.ObjectIDFromHex(accountId) 
		mostRecentTransfer,_ := api.accMgr.GetMostRecentTransaction(tranferAccId)
		response = fmt.Sprintf("Transfer request processed successfully : %v", mostRecentTransfer)
	case FIND_ACCOUNT_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		}

		var idRequest IdRequest
		err = json.Unmarshal(bodyBytes, &idRequest)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

		}

		account, err := api.handleFindAccountIntent(idRequest.Id.Hex())
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}
		response = fmt.Sprintf("Account found: %v", account)

	case TRANSACTIONS_INTENT:
		

		historyTable, err := api.handleTransactionsIntent(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		response = historyTable

	case SEARCH_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
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
		accountJSON, err := json.Marshal(account)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to marshal account"})
			return
		}
		response = string(accountJSON)

	case DEPOSIT_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		}

		var depositReq DepositRequest
		err = json.Unmarshal(bodyBytes, &depositReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		}

		 newBalance, err := api.handleDepositIntent(ctx, depositReq.Amount)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message": "server error"})
        return
    }

    response = fmt.Sprintf("Deposit processed successfully. New balance: %.2f", newBalance)
    
	

	case BALANCE_CHECK_INTENT:
	// Retrieve accountId from Gin context
	accountIdInterface, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "user ID not found in context"})
		return
	}

	accountId, ok := accountIdInterface.(string)
	if !ok || accountId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID in context"})
		return
	}

	// Call API method to get balance and transactions for the current account
	balance, transactions, err := api.handleCheckBalanceIntent(accountId, "")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message: err.Error()})
		return
	}

	// Prepare response with balance and transactions
	transactionInfos := make([]TransactionInfo, 0)
	for _, transaction := range transactions {
		if transaction.ToAccount.Hex() == accountId {
			transactionInfos = append(transactionInfos, TransactionInfo{
				FromAccount: transaction.FromAccount.Hex(),
				Amount:      transaction.Amount,
			})
		}
	}

	balResponse := BalanceResponse{
		Balance:      balance,
		Transactions: transactionInfos,
	}

	response =fmt.Sprintf("balance found: %v",balResponse )
	
	case GET_ALL_ACCOUNTS_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}

		var bankAccsRes BankAccsRes
		err = json.Unmarshal(bodyBytes, &bankAccsRes)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		}

		accounts, err := api.handleGetAccountsIntent()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to retrieve accounts"})
			return
		}

		// Respond with the fetched accounts
		response = fmt.Sprintf("Accounts: %v", accounts)

	}
	ctx.JSON(http.StatusOK, gin.H{"response": response})
	ctx.Set("response", response) // Set response in Gin context for retrieval
	ctx.Next()

}
func (api *ApiManager) handleTransactionsIntent(ctx *gin.Context) (string, error) {
	accountId, ok := ctx.Get("userId")
	if !ok || accountId == "" {
		return "", fmt.Errorf("accountId is required")
	}

	objectID, err := primitive.ObjectIDFromHex(accountId.(string))
	if err != nil {
		return "", fmt.Errorf("invalid account ID format: %v", err)
	}

	transactions, err := api.accMgr.GetTransactionsHistory(objectID)
	if err != nil {
		return "", fmt.Errorf("error fetching transactions: %v", err)
	}

	table, err := utils.FormatTransactionsTable(transactions)
	if err != nil {
		return "", err
	}

	return table, nil
}


func (api *ApiManager) handleTransferIntent(from, to string, amount float64) error {
	fromAccountID, err := primitive.ObjectIDFromHex(from)
	if err != nil {
		return err
	}
	var toAccountID primitive.ObjectID
    // Check if 'to' is an ObjectID (account ID)

     if toAccountID, err = primitive.ObjectIDFromHex(to); err != nil {
        // 'to' is not a valid ObjectID, assume it's a phone number
        account, err := api.accMgr.GetAccountByPhone(to)
        if err != nil {
            return fmt.Errorf("error finding account by phone: %v", err)
        }
        toAccountID = account.ID
    }

    // Perform the transfer operation
    err = api.accMgr.TransferAmountById(fromAccountID, toAccountID, amount)
    if err != nil {
        return fmt.Errorf("error transferring amount: %v", err)
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

func (api *ApiManager) handleDepositIntent( ctx *gin.Context,amount float64) (float64, error) {
	// Retrieve accountId from Gin context
	accountIdInterface, exists := ctx.Get("userId")
	if !exists {
		return 0,fmt.Errorf("user ID not found in context")
	}

	accountId, ok := accountIdInterface.(string)
	if !ok || accountId == "" {
		return 0,fmt.Errorf("invalid user ID in context")
	}

	// Convert account ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(accountId)
	if err != nil {
		return 0,fmt.Errorf("invalid account ID format: %v", err)
	}

	if amount <= 0 {
		return 0,fmt.Errorf("amount must be greater than zero")
	}

	// Perform the deposit operation
	err = api.accMgr.DepositToAccount(amount, objectID)
	if err != nil {
		return 0,fmt.Errorf("error depositing to account: %v", err)
	}
	  account, err := api.accMgr.SearchAccountById(objectID)
    if err != nil {
        return 0, fmt.Errorf("error retrieving updated account: %v", err)
    }

    return account.Balance, nil
	
}


func (api *ApiManager) handleCheckBalanceIntent(accountID, accountName string) (float64, []db.Transaction, error) {
	var account *db.BankAccount
	var err error

	if accountID != "" {
		objectID, err := primitive.ObjectIDFromHex(accountID)
		if err != nil {
			return 0, nil, fmt.Errorf("invalid account ID format: %v", err)
		}
		// Get the account by ID
		account, err = api.accMgr.SearchAccountById(objectID)
		if err != nil {
			return 0, nil, fmt.Errorf("error retrieving account by ID: %v", err)
		}
	} else if accountName != "" {
		// Search account by name or phone
		accounts, err := api.accMgr.SearchAccountByNameOrPhone(accountName)
		if err != nil {
			return 0, nil, fmt.Errorf("error searching for account: %v", err)
		}
		if len(accounts) == 0 {
			return 0, nil, fmt.Errorf("no account found with the provided name")
		}
		account = accounts[0] // Assuming we take the first matched account
	} else {
		return 0, nil, fmt.Errorf("account ID or name must be provided")
	}

	if account == nil {
		return 0, nil, fmt.Errorf("account not found")
	}

	// Retrieve the transactions
	transactions, err := api.accMgr.GetTransactionsHistory(account.ID)
	if err != nil {
		return 0, nil, fmt.Errorf("error retrieving transactions: %v", err)
	}

	return account.Balance, transactions, nil
}

func (api *ApiManager) handleGetAccountsIntent() ([]db.BankAccount, error) {
	// Call GetAccounts method to retrieve accounts
	accounts, err := api.accMgr.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("could not retrieve accounts: %v", err)
	}
	return accounts, nil
}
