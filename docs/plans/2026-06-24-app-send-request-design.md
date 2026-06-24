# App Send Request Design

## Goal

Wire the parsed request flow to the core runner through an app-level send process.

## Approach

Add `app.SendRequest(ctx, input)` as the app boundary for direct execution. It will reuse `LoadRequest` to scan and parse the selected request, then call `runner.Run` with the parsed method and URL.

The app result will include the request metadata needed for debug output and the response data needed for normal output: path, method, URL, status code, and body bytes.

## CLI Behavior

The CLI will call `app.SendRequest` with `cmd.Context()`.

Normal output will print the response body.

Debug output will keep the existing request path and parsed request line, and add the response status code as `Status: <code>`.

## Error Handling

Scanner, parser, request construction, network, and body-read errors will return normally as command errors.

Non-2xx HTTP responses will not be command errors in this slice. The status code and body will be returned and printed according to the normal/debug output rules.

## Testing

Use `httptest.Server` so tests do not hit the real network. Add app tests for `SendRequest`, then CLI tests for body output and debug status output.
