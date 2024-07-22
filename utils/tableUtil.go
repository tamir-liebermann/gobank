package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

// getHeaders dynamically gets the headers from the JSON data
func getSelectedHeaders() []string {
	return []string{"from_account", "amount" , "to_account" , "timestamp"}
}

// getSelectedRow returns only the selected columns from the record
func getSelectedRow(record map[string]interface{}, myAccountId string) []string {
	fromAccount := fmt.Sprintf("%v", record["from_account"])
	toAccount := fmt.Sprintf("%v", record["to_account"])
	amount := fmt.Sprintf("%v", record["amount"])
	date := fmt.Sprintf("%v", record["timestamp"])

	if fromAccount == myAccountId {
		fromAccount = "your account"
		amount = fmt.Sprintf("-%v", record["amount"])
	}

	if toAccount == myAccountId {
		toAccount = "your account"
		amount = fmt.Sprintf("%v", record["amount"])
	}

	return []string{fromAccount, amount, toAccount, date}
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
		table.Append(getSelectedRow(record, myAccountId))
	}
	table.Render()

	return buf.String(), nil
}
