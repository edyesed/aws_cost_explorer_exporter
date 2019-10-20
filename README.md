# aws_cost_explorer_exporter
This little program exports costs from the AWS CostExplorer API

It outputs a `|` separated file, which you can import/parse into other tooling. 

Pretty rudimentary in terms of exploring cost, and also useful to me.

Many things are hard coded ( date ranges, cost granularity, facets of cost exploration )

# How to run
You'll want **go > 1.12**, methinks.

```sh
sh $ git clone git@github.com:edyesed/aws_cost_explorer_exporter.git
sh $ AWS_PROFILE=something_meaningful_to_you go run main.go | tee /var/tmp/output.csv
```

# Notes
Pagination is functional, tho probably not implemented in a golang indiomatic sort of way. I'm not a go-ologist. please PR to make that more go-like. 
