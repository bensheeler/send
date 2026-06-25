# send

`send` is a command-line tool for sending HTTP requests defined in `.http` and `.rest` files.

Today, `send` executes one request from one request file and prints the response body to stdout. It is intentionally small and is growing toward a fuller HTTP file workflow.

## Quick Start

Create a request file:

```http
GET https://jsonplaceholder.typicode.com/users/1
Accept: application/json
```

Run it:

```sh
go run ./cmd/send examples/http/single-line.http
```

The response body is written to stdout.

Use `--debug` to see the resolved request file, parsed request, parsed headers, and response status:

```sh
go run ./cmd/send --debug examples/http/single-line.http
```

Debug output is formatted like this:

```text
/path/to/examples/http/single-line.http
GET https://jsonplaceholder.typicode.com/users/1
Accept: application/json
Status: 200
{response body}
```

## Request Files

Request files use `.http` or `.rest` extensions.

The supported request format is:

```http
METHOD URL
Header-Name: header value
Another-Header: another value
```

The request line must contain exactly two fields: an HTTP method and a URL.

Supported methods:

- `GET`
- `POST`
- `PUT`
- `PATCH`
- `DELETE`
- `HEAD`
- `OPTIONS`

Methods are normalized to uppercase before sending.

Leading blank lines and leading comment lines are ignored. Comment lines start with `#` or `//`.

## Headers

Headers are optional. Header lines come immediately after the request line and continue until the first blank line or the end of the file.

Header syntax is:

```http
Name: value
```

Any non-empty header name is accepted. There is no allowlist of supported header names, so common headers like `Accept`, `Authorization`, `Content-Type`, and custom headers are all parsed the same way.

Duplicate headers are preserved and sent in order. For example:

```http
GET https://example.com
X-Trace: one
X-Trace: two
```

Malformed header lines are rejected. A header line must include `:` and must have a non-empty name.

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

## Current Limitations

`send` currently supports a small direct execution flow only:

- Only the first request in a request file is parsed and sent.
- Request bodies are not supported yet.
- Variables are not resolved yet.
- Environment selection is not supported yet.
- Interactive request selection is not supported yet.
