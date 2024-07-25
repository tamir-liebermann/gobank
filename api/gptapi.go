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
	FIND_ACCOUNT_BY_PHONE_INTENT     = "find this account"
	TRANSACTIONS_INTENT     = "transactions"
	SEARCH_INTENT           = "search"
	DEPOSIT_INTENT          = "deposit"
	BALANCE_CHECK_INTENT    = "balance"
	GET_ALL_ACCOUNTS_INTENT = "all accounts"
	CHANGE_ACC_NAME_INTENT  = "change name"
)

// type AgentTransferRequest struct {
// 	Intent string          `json:"intent"`
// 	Body   TransferRequest `json:"body"`
// }

type GenericRequest struct {
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
		You are a bank API, you reply in json objects only, if unsure ask for clarification ,
		 try to guide the first time user present him the intent map without the get accounts intent in a human readable bullet pointes
		 and not json.	

		If the user's intent is clear but the input does not match the spec please guide him on the correct request parameters
	    Make it feel like a natrual conversation , you can use humor.


		If the user wants to transfer to phone number or account holder's name, give them:
		{
			"intent": "transfer", // must be this keyword
			"body":{
				
				to:"string", // must be the phone number or name only,
				amount:"float" // must be specified
			}
			
		}

		If the user wants to search for another account by spesific phone number, give them: 
		{
			"intent": "find this account", // must be this keyword
			"body": {
				"phone_number":"string", // must be the phone number 
			}
		}

		If the user wants to see his transactions history , give them a formatted table of the transactions history :
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
				
				"balance": "float" // must be specified
			}
		}

		If the user wants to change his account name, give them:
		{
			"intent": "change name" // must be this keyword
			"body": {
				"account_name": "string" // must be sepcified 
			}
		}

		If the user is admin and wants to see the all the existing accounts, give them: 
		{	
			"intent": "all accounts", // must be this keyword
			"body" :{
				
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

	var req GenericRequest
	var response string

	err = json.Unmarshal([]byte(textResp), &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": textResp})
		response = fmt.Sprintf("%v", textResp)
		ctx.Set("response",response)
		return
	}
	
	
	errorMsgMap :=  map[string]string{
		TRANSACTIONS_INTENT:"Transactions not found",
		TRANSFER_INTENT:"Please provide a valid name or phone number and amount",
		FIND_ACCOUNT_BY_PHONE_INTENT:"Please provide a valid phone number",
		SEARCH_INTENT: "Please provide a valid account name or phone",
		DEPOSIT_INTENT:"Please provide a valid amount",
		BALANCE_CHECK_INTENT:"Check for typos",
		GET_ALL_ACCOUNTS_INTENT:"You are not the Admin! ",
		CHANGE_ACC_NAME_INTENT: "Please provide a valid input",
	}
   
	// todo use transfer req
	
	switch req.Intent {
	case TRANSFER_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}

		var transferReq TransferRequest
		err = json.Unmarshal(bodyBytes, &transferReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return 
		}
		accountId := fmt.Sprintf("%v", accountId)
		
		err = api.handleTransferIntent(accountId, transferReq.To, transferReq.Amount)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": errorMsgMap[req.Intent]})
			response = errorMsgMap[req.Intent]
		    ctx.Set("response", response)
			return
		}
		
		response = fmt.Sprintf("Transfer request processed successfully : %v To  %v", transferReq.Amount, transferReq.To)
	case FIND_ACCOUNT_BY_PHONE_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
	    if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		return
	}

	var phoneRequest PhoneRequest
	err = json.Unmarshal(bodyBytes, &phoneRequest)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
		return
	}

	accountName, err := api.handleFindAccountByPhoneIntent(phoneRequest.PhoneNumber)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message":  errorMsgMap[req.Intent]})
		response = errorMsgMap[req.Intent]
		ctx.Set("response", response)
		return
	}

	response = fmt.Sprintf("Account found: %v", accountName)

	case TRANSACTIONS_INTENT:
		

		historyTable, err := api.handleTransactionsIntent(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message":  errorMsgMap[req.Intent]})
			response = errorMsgMap[req.Intent]
		    ctx.Set("response", response)
			return
		}

		response = historyTable

	case SEARCH_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}

		var accNameReq AccNameReq
		err = json.Unmarshal(bodyBytes, &accNameReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}
		account, err := api.handleSearchAccountByNameIntent(accNameReq.AccountHolder)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": errorMsgMap[req.Intent]})
			response = errorMsgMap[req.Intent]
		    ctx.Set("response", response)
			return
		}
		// accountJSON, err := json.Marshal(account)
		// if err != nil {
		// 	ctx.JSON(http.StatusInternalServerError, gin.H{"message": "failed to marshal account"})
		// 	return
		// }
	   response = fmt.Sprintf("Account found: %v ", account )

	case DEPOSIT_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return 
		}

		var depositReq DepositRequest
		err = json.Unmarshal(bodyBytes, &depositReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}

		 newBalance, err := api.handleDepositIntent(ctx, depositReq.Amount)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message": errorMsgMap[req.Intent]})
		response = errorMsgMap[req.Intent]
		ctx.Set("response", response)
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
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Message:  errorMsgMap[req.Intent]})
		response = errorMsgMap[req.Intent]
		ctx.Set("response", response)
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

	response =fmt.Sprintf("balance found: %v , ",balResponse.Balance )
	
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
			return
		}

		accounts, err := api.handleGetAccountsIntent(ctx)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"message":  errorMsgMap[req.Intent]})
			response = errorMsgMap[req.Intent]
		    ctx.Set("response", response)
			return
		}

		// Respond with the fetched accounts
		response = fmt.Sprintf("Accounts: %v", accounts)

	case CHANGE_ACC_NAME_INTENT:
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}
		var changeNameReq ChangeAccNameReq
		err = json.Unmarshal(bodyBytes, &changeNameReq)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})
			return
		}
		if changeNameReq.AccountHolder == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Account name is required"})
        return
    }
	accountIdInterface, exists := ctx.Get("userId")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{"message": "User ID not found in context"})
        return
    }

    accountId, ok := accountIdInterface.(string)
    if !ok || accountId == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID in context"})
        return
    }

    // Call the handler function
    updatedName, err := api.handleChangeAccNameIntent(accountId, changeNameReq.AccountHolder)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"message":  errorMsgMap[req.Intent]})
        response = errorMsgMap[req.Intent]
		    ctx.Set("response", response)
		    return
    }
		response = fmt.Sprintf("Username changed to : %v", updatedName)

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

	table, err := utils.FormatTransactionsTable(transactions,accountId.(string))
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
    // Check if 'to' is an ObjectID (account ID))
     if toAccountID, err = primitive.ObjectIDFromHex(to); err != nil {
		// 'to' is not a valid ObjectID, check if it's a phone number
		account, err := api.accMgr.GetAccountByPhone(to)
		if err != nil {
			// Not a phone number, assume it's an account holder's name
			account, err = api.accMgr.GetAccountByName(to)
			if err != nil {
				return fmt.Errorf("error finding account by phone or name: %v", err)
			}
			toAccountID = account.ID
		} else {
			toAccountID = account.ID
		}
	}

    // Perform the transfer operation
    err = api.accMgr.TransferAmountById(fromAccountID, toAccountID, amount)
    if err != nil {
        return fmt.Errorf("error transferring amount: %v", err)
    }

    return nil
}

func (api *ApiManager) handleSearchAccountByNameIntent(name string) ([]string, error) {
	// Call the updated SearchAccountByNameOrPhone function
	accounts, err := api.accMgr.SearchAccountByNameOrPhone(name)
	if err != nil {
		return nil, fmt.Errorf("error searching for account: %v", err)
	}

	// Check if no accounts were found
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no accounts found matching the provided name or phone number")
	}

	var accountNames []string
	for _, account := range accounts {
		accountNames = append(accountNames, account.AccountHolder)
	}

	return accountNames, nil
}

func (api *ApiManager) handleFindAccountByPhoneIntent(phone string) (string, error) {
	
	account, err := api.accMgr.GetAccountByPhone(phone)
	if err != nil {
		return "", fmt.Errorf("error searching for account by phone: %v", err)
	}

	if account == nil {
		return "", fmt.Errorf("account not found")
	}
	
	return account.AccountHolder, nil
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


func (api *ApiManager) handleCheckBalanceIntent(accountID, accountName string) (string, []db.Transaction, error) {
	var account *db.BankAccount
	var err error

	if accountID != "" {
		objectID, err := primitive.ObjectIDFromHex(accountID)
		if err != nil {
			return "", nil, fmt.Errorf("invalid account ID format: %v", err)
		}
		// Get the account by ID
		account, err = api.accMgr.SearchAccountById(objectID)
		if err != nil {
			return "", nil, fmt.Errorf("error retrieving account by ID: %v", err)
		}
	} else if accountName != "" {
		// Search account by name or phone
		accounts, err := api.accMgr.SearchAccountByNameOrPhone(accountName)
		if err != nil {
			return "", nil, fmt.Errorf("error searching for account: %v", err)
		}
		if len(accounts) == 0 {
			return "", nil, fmt.Errorf("no account found with the provided name")
		}
		account = accounts[0] // Assuming we take the first matched account
	} else {
		return "", nil, fmt.Errorf("account ID or name must be provided")
	}

	if account == nil {
		return "", nil, fmt.Errorf("account not found")
	}

	// Retrieve the transactions
	transactions, err := api.accMgr.GetTransactionsHistory(account.ID)
	if err != nil {
		return "", nil, fmt.Errorf("error retrieving transactions: %v", err)
	}
	balanceStr := fmt.Sprintf("%.2f", account.Balance)

	return balanceStr, transactions, nil
	
}

func (api *ApiManager) handleGetAccountsIntent(ctx *gin.Context) ([]db.BankAccount, error) {
    // Retrieve the user ID from the context
    userId, exists := ctx.Get("userId")
    if !exists {
        return nil, fmt.Errorf("user ID not found")
    }

    // Fetch the account based on userId
    objectID, err := primitive.ObjectIDFromHex(userId.(string))
    if err != nil {
        return nil, fmt.Errorf("invalid user ID format: %v", err)
    }

    account, err := api.accMgr.SearchAccountById(objectID)
    if err != nil {
        return nil, fmt.Errorf("could not find account: %v", err)
    }

    // Check the role of the account
    if account.Role != "admin" {
        return nil, fmt.Errorf("you are not authorized to perform this action")
    }

    // Call GetAccounts method to retrieve accounts
    accounts, err := api.accMgr.GetAccounts()
    if err != nil {
        return nil, fmt.Errorf("could not retrieve accounts: %v", err)
    }
    return accounts, nil
}

// handleChangeAccNameIntent updates the account name for the given account ID
func (api *ApiManager) handleChangeAccNameIntent(accountId string, newAccountName string) (string, error) {
    // Convert the accountId from string to ObjectID
    

    // Call the ChangeAccName method of AccManager
    updatedName, err := api.accMgr.ChangeAccName(accountId, newAccountName)
    if err != nil {
        return "", fmt.Errorf("error updating account name: %v", err)
    }

    return updatedName, nil
}
