package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/aws/aws-sdk-go/service/organizations"

	cexp "github.com/edyesed/aws_cost_explorer_exporter/internal/pkg/costexplore"
)

func main() {
	lbMonths := flag.Int("lookbackmonths", 1, "number of months to look back")
	flag.Parse()
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
	oresult, _ := osvc.ListAccounts(&organizations.ListAccountsInput{})
	// We try and keep track of friendly names for accounts
	for _, acct := range oresult.Accounts {
		acctMap[*acct.Id] = *acct.Name
	}

	svc := costexplorer.New(sess)
	ctx := context.Background()
	var results []*costexplorer.ResultByTime

	startDate := cexp.LookbackMonths(*lbMonths, time.Now())
	endDate := time.Now()

	// pagination handling
	var paginationToken string = ""
	for {
		params := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &costexplorer.DateInterval{
				Start: aws.String(startDate.Format("2006-01") + "-01"),
				End:   aws.String(endDate.Format("2006-01") + "-01"),
			},
			Granularity: aws.String("MONTHLY"),
			GroupBy: []*costexplorer.GroupDefinition{
				{
					Type: aws.String("DIMENSION"),
					Key:  aws.String("SERVICE"),
				},
				{
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
			acctID := *g.Keys[1]
			serviceName := *g.Keys[0]
			fname := acctMap[acctID]
			if fname == "" {
				fname = acctID
			}
			fmt.Printf("%s | %s | %s | %s | %s | %s\n", *p.TimePeriod.Start, *p.TimePeriod.End, acctID, fname, serviceName, *g.Metrics["UnblendedCost"].Amount)
		}
	}
}
