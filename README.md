# cron3
## Go utility program for day-to-day Mongo backups.
* General idea:
  * Execute `mongodump` command
  * Generated `.bson` file upload to S3
  * Delete files older than 3 days from S3
  
## Install

* Make **_config.yaml_** based on **_config-template.yaml_**
* `file_name` is the relative path of the `.bson` document
* `cron_time` is the cron expression (interval for executing func)

## Run

* `go build`
* `./cron3 2>> log &`
