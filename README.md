# send

`send` is a command-line tool for working with HTTP requests defined in `.http` and `.rest` files.

## Usage

```sh
go run ./cmd/send <request-file>
```

For now, `send` resolves the request file selector and prints the matched file path.

## Request File Selection

Selectors with `.http` or `.rest` extensions are exact relative paths:

```sh
go run ./cmd/send requests/users.http
```

This does not recursively search for `users.http`.

Bare names search recursively within the lookup depth:

```sh
go run ./cmd/send users
```

This can match `users.http` or `users.rest` under the current working directory.

Extensionless paths are exact relative stems:

```sh
go run ./cmd/send requests/users
```

This can match `requests/users.http` or `requests/users.rest`.
