package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/olekukonko/tablewriter"
)

// getHeaders dynamically gets the headers from the JSON data
func getSelectedHeaders() []string {
	return []string{"transaction_details"}
}

// getSelectedRow returns only the selected columns from the record
func getSelectedRow(record map[string]interface{}, myAccountId string) ([]string, error) {
	fromAccount := fmt.Sprintf("%v", record["from_account"])
	toAccount := fmt.Sprintf("%v", record["to_account"])
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
	var details string
	if fromAccount == myAccountId {
		details = fmt.Sprintf("your account -%v to %v %v ago", amount, toAccount, timeAgoStr)
	} else if toAccount == myAccountId {
		details = fmt.Sprintf("from %v %v your account %v ago", fromAccount, amount, timeAgoStr)
	} else {
		details = fmt.Sprintf("from %v %v to %v %v ago", fromAccount, amount, toAccount, timeAgoStr)
	}

	return []string{details}, nil
}


// FormatTransactionsTable formats transactions into a table with selected columns
func FormatTransactionsTable(transactions interface{},myAccountId string) (string, error) {
	transactionsBytes, err := json.Marshal(transactions)
	if err != nil {
		return "", fmt.Errorf("error converting transactions to JSON: %v", err)
	}

	var jsonData []map[string]interface{}
	if err := json.Unmarshal(transactionsBytes, &jsonData); err != nil {
		return "", fmt.Errorf("error parsing JSON data: %v", err)
	}

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(getSelectedHeaders())
	for _, record := range jsonData {
		row, err := getSelectedRow(record, myAccountId)
		if err != nil {
			return "", fmt.Errorf("error formatting row: %v", err)
		}
		table.Append(row)
	}
	table.Render()

	return buf.String(), nil
}
