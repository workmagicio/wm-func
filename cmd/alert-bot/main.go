package main

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"os"
	"strings"
	"wm-func/cmd/alert-bot/exec_sql"
)

func main() {
	sqls := GetAllDataQualityChecks()

	for _, sql := range sqls {
		tenants := GetCheckTenantIds()
		b, err := os.ReadFile("/Users/xukai/workspace/workmagic/wm-func/cmd/alert-bot/promot.txt")
		if err != nil {
			panic(err)
		}

		if strings.Contains(sql.DataSQL, "{{tenant_ids}}") {
			sql.DataSQL = strings.ReplaceAll(sql.DataSQL, "{{tenant_ids}}", strings.Join(tenants, ","))
			sql.CheckSQL = strings.ReplaceAll(sql.CheckSQL, "{{tenant_ids}}", strings.Join(tenants, ","))
		}

		prompt := string(b)
		prompt = strings.ReplaceAll(prompt, "{{MAIN_DATA_JSON}}", exec_sql.Exec(sql.DataSQL))
		prompt = strings.ReplaceAll(prompt, "{{CHECK_DATA_JSON}}", exec_sql.Exec(sql.CheckSQL))
		prompt = strings.ReplaceAll(prompt, "{{CUSTOMER_RULE}}", sql.AIPrompt)
		generateWithText(prompt)
	}

}

// generateWithText shows how to generate text using a text prompt.
func generateWithText(prompt string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:      "AIzaSyBIWCe93sSdEFu8w_X6MDgLYNQz6W8fdms",
		HTTPOptions: genai.HTTPOptions{APIVersion: "v1"},
	})
	if err != nil {
		panic(err)
		//return fmt.Errorf("failed to create genai client: %w", err)
	}

	resp, err := client.Models.GenerateContent(ctx,
		"gemini-2.5-pro",
		genai.Text(prompt),
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		panic(err)
	}

	respText := resp.Text()

	fmt.Println(respText)

	return nil
}
