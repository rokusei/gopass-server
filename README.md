# gopass-server
Privacy oriented open source password vault/manager written in Go.

Simple design and API. Supports using popular database supported by [GORM](https://github.com/go-gorm/gorm) (sqlite, mysql, postgres, sqlserver, clickhouse, bigquery).

## Design
The server only stores the following:
- SHA256 hash of Verification Code (to be emailed to users on registration to verify their email)
- SHA256 hash of Email
- bcrypt hash (default cost) of the Authentication Hash (AuthHashHash) which is derived from a salt and the master password
- pdkdf2 encrypted blob (101101 iterations) of Vault entries encrypted by the Encryption Hash, which is also derived from a salt and the master password

### API 
`api` contains the API endpoints and all of them return `http.Handlers` so that you can wrap them in whatever middleware you'd like.

The API architecture is broken up into two parts:
- `user` handles HTTP requests for creating and fetching a user account
- `vault` handles HTTP requests for interacting with your password vault

### DB
`db` uses [GORM](https://github.com/go-gorm/gorm) as an ORM. 

The DB architecture is broken up into two parts:
- `user` handles database interactions for creating and fetching a user account
- `vault` handles database interactions for interacting with your password vault

## TODO
- Determine how to perform account verifications
- Unit Test and mock all the things
- Frontend?

## Contributing
Any and all contributions are greatly appreciated. Feel free to open an issue or send over a pull request. 

Please take the following into consideration when contributing:
- Test coverage (`go test -v -cover ./...`) should not decrease from your contribution
- Run `go mod tidy` prior to creating a pull request
- Do not upgrade dependencies (e.g. `go get -u`) as apart of feature requests, please break dependency upgrade contributions out to a separate PR