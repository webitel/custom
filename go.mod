module github.com/webitel/custom

go 1.22.5

replace github.com/webitel/proto/gen => ./proto/gen

require (
	github.com/Masterminds/squirrel v1.5.4
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/jackc/pgx/v5 v5.7.2
	github.com/webitel/proto/gen v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.36.5
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/text v0.22.0 // indirect
)
