package utils

import (
	"bytes"

	"encoding/json"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
	
)
type SearchResult struct {
    AccountHolder string `json:"account_holder"`
    PhoneNumber   string `json:"phone_number"`
}
// getHeaders dynamically gets the headers from the JSON data
func getSelectedHeaders() []string {
	return []string{"sender_name", "amount", "receiver_name", "timestamp"}}

// getSelectedRow returns only the selected columns from the record
func getSelectedRow(record map[string]interface{}, myAccountId string,  ) ([]string, error) {
	fromAccount := fmt.Sprintf("%v", record["from_account"])
	senderName :=  fmt.Sprintf("%v", record["sender_name"])
	toAccount := fmt.Sprintf("%v", record["to_account"])
	receiverName :=  fmt.Sprintf("%v", record["receiver_name"])
	amount := fmt.Sprintf("%v", record["amount"])
	date := fmt.Sprintf("%v", record["timestamp"])
	timestamp, err := time.Parse(time.RFC3339, date)
	if err != nil {
		return nil, fmt.Errorf("error parsing timestamp: %v", err)
	}

	timeAgo := time.Since(timestamp)
	var timeAgoStr string

	switch {
	case timeAgo.Hours() >= 24:
		timeAgoStr = fmt.Sprintf("%v days ago", int(timeAgo.Hours()/24))
	case timeAgo.Hours() >= 1:
		timeAgoStr = fmt.Sprintf("%v hours ago", int(timeAgo.Hours()))
	default:
		timeAgoStr = fmt.Sprintf("%v minutes ago", int(timeAgo.Minutes()))
	}
	
	
	
	

	if fromAccount == myAccountId {
		fromAccount = "your account"
		senderName = "your account"
		amount = fmt.Sprintf("-%v", record["amount"])
	}

	if toAccount == myAccountId {
		toAccount = "your account"
		receiverName= "your account"
		amount = fmt.Sprintf("%v", record["amount"])
	}
	

	return []string{senderName,amount,receiverName, timeAgoStr}, nil
}

func reverseSlice(slice []map[string]interface{}) []map[string]interface{} {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
	return slice
}


// FormatTransactionsTable formats transactions into a table with selected columns
func FormatTransactionsTable(transactions interface{},myAccountId string,  ) (string, error) {
	transactionsBytes, err := json.Marshal(transactions)
	if err != nil {
		return "", fmt.Errorf("error converting transactions to JSON: %v", err)
	}

	var jsonData []map[string]interface{}
	if err := json.Unmarshal(transactionsBytes, &jsonData); err != nil {
		return "", fmt.Errorf("error parsing JSON data: %v", err)
	}

	jsonData = reverseSlice(jsonData)

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(getSelectedHeaders())
	for _, record := range jsonData {
		row, err := getSelectedRow(record, myAccountId )
		if err != nil {
			return "", fmt.Errorf("error formatting row: %v", err)
		}
		table.Append(row)
	}
	table.Render()

	return buf.String(), nil
}
func getSearchHeaders() []string {
    return []string{"Account Holder", "Phone Number"}
}

// getSearchRow returns the row data for the search results table
func getSearchRow(result SearchResult) []string {
    return []string{
        result.AccountHolder,
        result.PhoneNumber,
    }
}

// FormatSearchResultsTable formats search results into a table with selected columns
func FormatSearchResultsTable(results []SearchResult) (string, error) {
    var buf bytes.Buffer
    table := tablewriter.NewWriter(&buf)
    table.SetHeader(getSearchHeaders())
    
    for _, result := range results {
        row := getSearchRow(result)
        table.Append(row)
    }
    
    table.Render()

    return buf.String(), nil
}

