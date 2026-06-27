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
[METHOD] URL [HTTP-Version]
Header-Name: header value
Another-Header: another value

raw request body
```

The request line contains a URL, an optional HTTP method, and an optional HTTP version. When the method is omitted from a URL-only request line, `send` uses `GET`.

Valid HTTP versions are `HTTP/1.1` and `HTTP/2`. `send` parses and preserves the request-line version, but the current runner does not force the wire protocol; Go's HTTP transport negotiates the actual protocol.

Supported methods:

- `GET`
- `HEAD`
- `POST`
- `PUT`
- `DELETE`
- `CONNECT`
- `PATCH`
- `OPTIONS`
- `TRACE`
- `LOCK`
- `UNLOCK`
- `PROPFIND`
- `PROPPATCH`
- `COPY`
- `MOVE`
- `MKCOL`
- `MKCALENDAR`
- `ACL`
- `SEARCH`

Methods are normalized to uppercase before sending.

Leading blank lines and leading comment lines are ignored. Comment lines start with `#` or `//`.

## Multiple Requests

Request files can contain multiple requests separated by a line starting with `###`:

```http
GET https://example.com/users/1

### createUser
POST https://example.com/users
Content-Type: application/json

{"name":"Ada"}
```

Text after `###` names the following request. Request names can also be written with portable `# @name` metadata before the request line:

```http
### Create user
# @name createUser
POST https://example.com/users
```

When both forms are present, `# @name` is the request name.

By default, `send` sends the first request in a file. To send a named request, pass the request name as the second argument:

```sh
go run ./cmd/send examples/http/multiple-requests.http createPost
```

Request names are case-sensitive. Names with spaces, such as separator titles, must be quoted for your shell:

```sh
go run ./cmd/send examples/http/multiple-requests.http 'Create user'
```

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

## Request Bodies

Bodies are optional. Add a blank line after headers, then write the raw body.

```http
POST https://example.com/users
Content-Type: application/json

{"name":"Ada"}
```

Trailing blank lines are trimmed before sending. Internal newlines and indentation are preserved.

Only raw bodies are supported. `send` does not yet support JSON helpers, file includes, multipart forms, variables, or generated content.

## Request File Selection

The command syntax is:

```sh
go run ./cmd/send <request-file> [request-name]
```

If `[request-name]` is omitted, `send` sends the first request in the file. If it is present, `send` sends the request whose name exactly matches it.

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

- Multiple requests are supported, but interactive request selection is not supported yet.
- Variables are not resolved yet.
- Environment selection is not supported yet.
