# Contest

A tool for API contract testing.

## Usage

`contest --schema openapidoc.yaml --schema openapidoc2.yaml --suite suite.yaml`

### Suite YAML

The suite.yaml file describes the suite of contracts that should be tested.

A suite has the following properties:
- `headers`: global headers added to every request
- `severity`: configure the severity of failure reasons (see section [Severity](#severity))
- `contracts`: single contracts testing a single URL (see section [Contract](#contract))
- `specFiles`: OpenAPI documents to load and test (see section [Spec File](#spec-file))


#### Example

```yaml
suite:
  # headers that will be added to every request
  headers:
    Api-Key: 1234
  
  # configure the severity of failure reasons to only warn not fail the entire suite
  severity:
    unexpected.responseTime: warn
  
  # Contracts
  contracts:
    - url: https://example.com/api/test-me
      headers:
        # global headers can be overridden
      expect:
        # ... expectations
  
  # Spec files
  specFiles:
    - path: ./openapidoc.yaml
      baseUrl: https://example.com/api
      operations:
        api.list:
        api.view:
          parameters:
            id: 123
```


#### Contract

A contract describes a URL and expectations about how the response and/or data should look.
The URL can also be a local file (if the protocol is `file://`). A contract can override global
headers and have parameters.

Supported expectations:

|      Name      |                            Description                            |
| -------------- | ----------------------------------------------------------------- |
| `status`       | HTTP status code                                                  |
| `contentType`  | Content-Type header in the response (w/ or w/o extensions)        |
| `schema`       | Schema of a JSON response (can be suffixed with `[]` for an array) |
| `responseTime` | The maximum allowed response time in ms                           |


#### Spec File

A spec file describes which operations from an OpenAPI 3.0 document to test.
You need to specify a `baseUrl` for the requests, since the paths in the OpenAPI definition are
all relative.

Only operations explicitly mentioned in the suite will be executed. The resulting contracts
will always expect: `status: 200`, `contentType: application/json`, and the `schema` from the
operation in the OpenAPI definition. Additional expectations are not supported at this time.

You may pass parameters to the operation, which will be handled like normal parameters on any
contract, except that the location is not necessary and will be automatically found using
the parameter object from the OpenAPI definition.


#### Parameters

Parameters are specified as a key value map. The keys consist of two parts: `location` and `name`.
`location` specifies where in the request the parameter should be substituted and is consistent
with the [OpenAPI 3.0 Parameter.in field](https://swagger.io/specification/#parameter-object).
Currently supported locations are: `"path"`, `"header"`.

At `location`, `"{name}"` is substituted with `"value"`.

```yaml
parameters:
  path:name: value
  header:param: super_interesting
```

#### Severity

Different failure reasons can have different severities. This can be used in cases where
a particular failure should not cause the whole suite to fail.
Currently, only the severity `warn` is supported.

Supported failure reasons:

|           Name            |               Description               |
| ------------------------- | --------------------------------------- |
| `contract`                | A contract was invalid                  |
| `http`                    | HTTP request failed                     |
| `io`                      | Error loading a file                     |
| `format`                  | The data could not be parsed            |
| `unexpected.status`       | Unexpected status code                  |
| `unexpected.schema`       | Schema did not match                    |
| `unexpected.content-type` | Unexpected Content-Type response header |
| `unexpected.responseTime` | Response time was greater than expected |