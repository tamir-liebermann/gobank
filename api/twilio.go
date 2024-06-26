package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (api *ApiManager) sendWhatsAppMessage(to, message string) error {
    accountSid := "AC009bc3c85f2212a8f4cdc0c32be81ef8"
    authToken := "f4349fb88ade7717d074d8ff1c47a74f"
    from := "whatsapp:+14155238886"

    urlStr := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", accountSid)

    // Construct the payload using url.Values
    data := url.Values{}
    data.Set("To", to)
    data.Set("From", from)
    data.Set("Body", message)

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
        fmt.Println("Response from Twilio:", string(bodyBytes))
        return nil
    } else {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("twilio API error: %s", string(bodyBytes))
    }
}

func (api *ApiManager) handleTwilioWebhook(ctx *gin.Context) {
    var twilioReq struct {
        From string `form:"From"`
        Body string `form:"Body"`
    }

    if err := ctx.ShouldBind(&twilioReq); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
    err = api.sendWhatsAppMessage(twilioReq.From, responseStr)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Request processed and response sent via WhatsApp"})
}