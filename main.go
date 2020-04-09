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
)

func lookbackMonths(lbmonths int) time.Time {
	now := time.Now()
	endDate := now.AddDate(0, -1*lbmonths, 0)
	return endDate
}

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

	startDate := lookbackMonths(*lbMonths)
	endDate := time.Now()

	// pagination handling
	var paginationToken string = ""
	for {
		params := &costexplorer.GetCostAndUsageInput{
			TimePeriod: &costexplorer.DateInterval{
				Start: aws.String(startDate.Format("2006-01") + "-01"),
				End:   aws.String(endDate.Format("2006-01") + "-01"),
			},
			Granularity: aws.String("DAILY"),
			GroupBy: []*costexplorer.GroupDefinition{
				{
					Type: aws.String("DIMENSION"),
					Key:  aws.String("SERVICE"),
				},
				/* {
					Type: aws.String("DIMENSION"),
					Key:  aws.String("LINKED_ACCOUNT"),
				},

					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("AZ"),
					},
				*/
				{
					Type: aws.String("DIMENSION"),
					Key:  aws.String("INSTANCE_TYPE"),
				},
				/*
					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("OPERATION"),
					},
					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("USAGE_TYPE"),
					},
					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("USAGE_TYPE_GROUP"),
					},
					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("PURCHASE_TYPE"),
					},
					{
						Type: aws.String("DIMENSION"),
						Key:  aws.String("RECORD_TYPE"),
					},
				*/
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
			fmt.Printf("%s | %s | %s | %s | %s | %s | %s | %s \n", *p.TimePeriod.Start, *p.TimePeriod.End, acctID, fname, serviceName, *g.Metrics["UnblendedCost"].Amount, *g.Metrics["BlendedCost"].Amount, *g.Metrics["UsageQuantity"].Amount)
		}
	}
}
