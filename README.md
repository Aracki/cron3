# cron3
## Go utility program for day-to-day Mongo backups.
* General idea:
  * Execute `mongodump` command
  * Generated `.bson` file upload to S3
  * Delete files older than 3 days from S3

