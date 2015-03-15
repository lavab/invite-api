# invite-api

<img src="https://mail.lavaboom.com/img/Lavaboom-logo.svg" align="right" width="300px" />

Golang API of the Lavaboom's invitation app.

Allows inviting users without having them create an account in the main
frontend app. Adds a new flow to the user registration cycle, in which user
receives an invitation code that they can later use to create an account, while
being compatible with the existing account setup wizard.

## Requirements

 - RethinkDB

## How it works

An invitation token is a document in `invites` table of the database defined by
`rethinkdb_name`. If the invitation doesn't contain an `account_id` field, then
user is able to create an account with any free username / unused email. If the
`account_id` is set, then `name` and `alt_email` of that account will be
enforced on the user. After consuming a token, it will be removed by the
backend of the setup wizard.

## Creating a new invitation

Unfortunately a tool for generating invites is not ready to be opensourced, so
in order to create a new token, you need to insert a document into the
`invites` table. Example:

```json
{
    "id": "1q2w3e4r5t6y7u8i9o0p",         // 20 characters-long token
    "account_id": "qawsedrftgyhujikolp0", // ID of the account of lavab/api
    "created_at": r.now()
}
```

## Usage

### Inside a Docker container

*This image will be soon uploaded to Docker Hub.*

```bash
git clone https://github.com/lavab/invite-api.git
cd invite-api
docker build -t "lavab/invite-api" .
docker run \
    -p 127.0.0.1:8000:8000 \
    -e "RETHINKDB_ADDRESS=172.8.0.1:28015" \
    --name invite-api \
    lavab/invite-api
```

### Directly running the app

```bash
go get github.com/lavab/invite-api
invite-api --rethinkdb_address=127.0.0.1:28015
```

## License

This project is licensed under the MIT license. Check `LICENSE` for more
information.