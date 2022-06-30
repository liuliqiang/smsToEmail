# smsToEmail
A http server recieved SMS from phone, and send to email

# usage

## run the server

```
# export SRC_EMAIL_ADDR=<you email addr>
# export SRC_EMAIL_PASS=<you email pass>
# export DEST_EMAIL_ADDR=<email addr receive sms>
# make runforever
```

## request with http client

```
# curl --location --request GET 'http://127.0.0.1:8080/sms' \
--header 'Content-Type: application/json' \
--data-raw '{
    "from": "13800138002",
    "sms": "See you tomorrow!"
}'
```

and your email will recieved the sms information.