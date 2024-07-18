package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

// getHeaders dynamically gets the headers from the JSON data
func getSelectedHeaders() []string {
	return []string{"from_account", "amount" , "to_account"}
}

// getSelectedRow returns only the selected columns from the record
func getSelectedRow(record map[string]interface{}) []string {
	return []string{
		fmt.Sprintf("%v", record["from_account"]),
		fmt.Sprintf("%v", record["amount"]),
		fmt.Sprintf("%v", record["to_account"]),

	}
}

// FormatTransactionsTable formats transactions into a table with selected columns
func FormatTransactionsTable(transactions interface{}) (string, error) {
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
		table.Append(getSelectedRow(record))
	}
	table.Render()

	return buf.String(), nil
}
