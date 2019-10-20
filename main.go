package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/organizations"
)

func main() {
	metrics := []string{
		"BlendedCost",
		"UnblendedCost",
		"UsageQuantity",
	}
	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		fmt.Println("ERROR IN SESSION", err)
		os.Exit(1)
	}

	acctMap := make(map[string]string)

	osvc := organizations.New(sess)
	oresult, err := osvc.ListAccounts(&organizations.ListAccountsInput{})
        // We try and keep track of friendly names for accounts
	for _, acct := range oresult.Accounts {
		acctMap[*acct.Id] = *acct.Name
	}

	svc := costexplorer.New(sess)
	ctx := context.Background()
	var results []*costexplorer.ResultByTime

	// pagination handling
	var paginationToken string = ""
	for {
		params := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &costexplorer.DateInterval{
				Start: aws.String("2019-01-01"),
				End:   aws.String("2019-10-17"),
			},
			Granularity: aws.String("MONTHLY"),
			GroupBy: []*costexplorer.GroupDefinition{
				&costexplorer.GroupDefinition{
					Type: aws.String("DIMENSION"),
					Key:  aws.String("SERVICE"),
				},
				&costexplorer.GroupDefinition{
					Type: aws.String("DIMENSION"),
					Key:  aws.String("LINKED_ACCOUNT"),
				},
			},
			Metrics: aws.StringSlice(metrics),
		}
		if paginationToken != "" {
			params.NextPageToken = aws.String(paginationToken)
		}

		result, err := svc.GetCostAndUsageWithContext(
			ctx,
			params,
		)
		if err != nil {
			fmt.Println("Error happened.", err)
			os.Exit(1)
		}
		results = append(results, result.ResultsByTime...)
		if result.NextPageToken == nil {
			break
		}
		paginationToken = *result.NextPageToken
	}
	for _, p := range results {
		for _, g := range p.Groups {
			acctId := *g.Keys[1]
			serviceName := *g.Keys[0]
			fname := acctMap[acctId]
			if fname == "" {
				fname = acctId
			}
			fmt.Println(fmt.Sprintf("%s | %s | %s | %s | %s | %s", *p.TimePeriod.Start, *p.TimePeriod.End, acctId, fname, serviceName, *g.Metrics["UnblendedCost"].Amount))
		}
	}
}