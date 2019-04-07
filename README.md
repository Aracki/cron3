# cron3

Go utility program for day-to-day Mongo backups. Used by [countgo](https://github.com/Aracki/countgo). <br>

* General idea:
  * Execute `mongodump` command
  * Generated `.bson` file upload to S3
  * Delete files older than 3 days from S3
  
## Install

* Make **_config.yaml_** based on **_config-template.yaml_**
* `file_name` is the relative path of the `.bson` document
* `cron_time` is the [cron expression](https://godoc.org/github.com/robfig/cron#hdr-Predefined_schedules) (eg. `0 30 6 * * *` for 6:30 AM)

## Usage

### Run

* `go build`
* `./cron3 2>> log &`

### Import a .bson file into a mongo database

* `mongorestore -d aracki -c visitors /path/file.bson`

### Run mongo in a docker container 

docker run --name mongo --rm -p 27017:27017 mongo:latest