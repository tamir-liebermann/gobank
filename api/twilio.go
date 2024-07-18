package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tamir-liebermann/gobank/db"
	"github.com/tamir-liebermann/gobank/env"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/twilio/twilio-go/client"

)

const TwilioUser = "twilio_user"

func (api *ApiManager) sendWhatsAppMessage(to, message string) error {
	spec := env.New()
	accountSid := spec.TwilioAccSid
	authToken := spec.TwilioAuth
    from := spec.TwilioPhoneNum
    

	urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)
    
   

	maxCharsPerMessage := 1600
	messageParts := splitMessage(message, maxCharsPerMessage)

	// Iterate over each message part and send it
	for i, part := range messageParts {
		// Construct payload
		data := url.Values{}
		data.Set("To", to)
		data.Set("From", from)
		data.Set("Body", part)

		req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
		if err != nil {
			return err
		}
		req.SetBasicAuth(accountSid, authToken)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Sent part %d/%d: Response from Twilio: %s\n", i+1, len(messageParts), string(bodyBytes))
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("twilio API error: %s", string(bodyBytes))
		}
	}
	fmt.Println("All message parts sent successfully.")
	return nil
}

func splitMessage(msg string, maxLen int) []string {
	var parts []string
	for i := 0; i < len(msg); i += maxLen {
		end := i + maxLen
		if end > len(msg) {
			end = len(msg)
		}
		parts = append(parts, msg[i:end])
	}
	return parts
}

type TwilioReq struct {
	From string `form:"From" json:"From"`
	Body string `form:"Body" json:"Body"`
   
}

var PhoneNumberRegexp = regexp.MustCompile(`^\+\d{1,3}[-.\s]?\d{1,4}[-.\s]?\d{1,4}[-.\s]?\d{1,9}$`)

func (api *ApiManager) getAccountFromTwilioReq( ctx *gin.Context, req TwilioReq) (*db.BankAccount, error) {
	
    phone := strings.TrimPrefix(req.From, "whatsapp:")
	ok := PhoneNumberRegexp.Match([]byte(phone))
	if !ok {
		return nil, errors.New("this is not twilio number")
	}

	account, err := api.accMgr.GetAccountByPhone(phone) // todo make specific call
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			//  return create account
			account, err = api.accMgr.CreateAccount("guest", "abc", 1000, phone ,"user")

			if err != nil {
				return nil, err
			}
		}else{
            return nil, err
        }
	}
	ctx.Set("userId", account.ID.Hex())
	return account, nil

}

func (api *ApiManager) handleTwilioWebhook(ctx *gin.Context) {
	var twilioReq TwilioReq // find out if twillio also sends some token to ensure this in auth, and not impersonator
    


	if err := ctx.ShouldBind(&twilioReq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	

    // Fetch or create the Twilio secret
   
	_, err := api.getAccountFromTwilioReq(ctx,twilioReq)

    if err != nil {
        	ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
    }


	userInput := strings.ToLower(strings.TrimSpace(twilioReq.Body))

	chatReq := ChatReq{
		UserText: userInput,
	}

	chatReqBytes, err := json.Marshal(chatReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process request"})
		return
	}

	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(chatReqBytes))

	api.handleChatGPTRequest(ctx)

	response, exists := ctx.Get("response")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no response from ChatGPT"})
		return
	}

	responseStr, ok := response.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "response is not a string"})
		return
	}

	// Prepare and send the HTTP POST request to Twilio API
	err = api.sendWhatsAppMessage( twilioReq.From, responseStr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request processed and response sent via WhatsApp"})
}





func validateTwilioRequest(ctx *gin.Context) bool {
	spec := env.New()
	authToken := spec.TwilioAuth
	requestValidator := client.NewRequestValidator(authToken)
	url := spec.AppWebhookUrl // Replace with your actual webhook URL

	// Extract the Twilio signature from the header
	signature := ctx.GetHeader("X-Twilio-Signature")
	if signature == "" {
		return false
	}

	// Extract the parameters from the request
	params := make(map[string]string)
	for key, values := range ctx.Request.PostForm {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}

	// Validate the request
	return requestValidator.Validate(url, params, signature)
}

