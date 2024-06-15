package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sashabaranov/go-openai"
)

const (
	TRANSFER_INTENT = "transfer"
)

// type AgentTransferRequest struct {
// 	Intent string          `json:"intent"`
// 	Body   TransferRequest `json:"body"`
// }

type GenericRequst struct {
	Intent string          `json:"intent"`
	Body   []byte `json:"body"`
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

	client := openai.NewClient("sk-proj-JKaZ7i09WvyceQo5zSCdT3BlbkFJxSsFxUuGjeqfBRPN5wlO")

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
		var transferReq TransferRequest
		err = json.Unmarshal(req.Body, &transferReq)
	if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "please specify a clear request"})

	}

		err := api.handleTransferIntent(accountId,transferReq.To,transferReq.Amount)
		if err != nil{
					ctx.JSON(http.StatusInternalServerError, gin.H{"message": "servererror"})

		}
	}
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

func TransferFunds(token string, body interface{}) {
	gptResp := "sendMoney/user123/1234"
	args := strings.Split(gptResp, "/")

	url := args[0]

	if url == "sendMoney" {
		token := "mockJWTToken"

		amount, err := strconv.Atoi(args[2])
		if err != nil {
			panic(err)
		}

		reqBody := TransferRequest{
			From:   token,
			To:     args[1],
			Amount: float64(amount),
		}

		reqBodyBytes, err := json.Marshal(reqBody)
		if err != nil {
			panic(err)
		}

		req, err := http.NewRequest("POST", "/account/transfer/:id", bytes.NewBuffer(reqBodyBytes))
		if err != nil {
			panic(err)
		}

		req.Header.Set("Content-Type", "application/json")

		client := openai.NewClient("sk-proj-JKaZ7i09WvyceQo5zSCdT3BlbkFJxSsFxUuGjeqfBRPN5wlO")

		rules := `
		You are a bank API.

		If the user wants to transfer, give them:
		{
			"url": "string",
			"body": "object"
		}

		For transfer, the body is:
		type TransferRequest struct {
			From   string  json:"from"
			To     string  json:"to"
			Amount float64 json:"amount"
		}

		The URL is "gobank/account/transfer".

		For listing transactions, the URL is "gobank/account/transactions" and the body is empty.
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
						Content: "I want to transfer 100 from John to Jane",
					},
				},
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		fmt.Println(resp.Choices[0].Message.Content)
	}
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
