# Contest

A simple tool for contract testing web APIs.

## Installation

You can find the binaries of the latest release [here](https://github.com/cadast/contest/releases/latest).

There is also a homebrew package for macOS which you can install by running `brew install cadast/contest/contest`. 

## Usage

`contest --schema openapidoc.yaml --schema openapidoc2.yaml --suite suite.yaml`

### Contest YAML

The contest.yaml file describes the suite of contracts that should be tested.

A suite has the following properties:
- `headers`: global headers added to every request
- `severity`: configure the severity of failure reasons (see section [Severity](#severity))
- `contracts`: single contracts testing a single URL (see section [Contract](#contract))
- `specFiles`: OpenAPI documents to load and test (see section [Spec File](#spec-file))

#### Linting

The contestSchema.json file describes the format of the contest.yaml file. You can add this schema
to your editor to have linting in your contestSchema.json.

##### VSCode

In VSCode using the [YAML Extension from RedHat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)
you can easily add the schema, by modifying your workspace or user config:

```json5
{
  // ... other config entries
  "yaml.schemas": {
    "<PATH_TO_CONTESTSCHEMA.JSON>": [
      "contest.yaml",
      "*.contest.yaml"
    ],
  }
}
```

##### JetBrains

To add the schema to a JetBrains IDE, follow this guide: [Using custom JSON schemas](https://www.jetbrains.com/help/idea/json.html#ws_json_schema_add_custom).


#### Example

You can find a complete example in the `example.contest.yaml`.

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

A contract can have the `anyOf` parameter, which is a list of contracts. If set, the response will be validated against
all of those and if at least one subcontract does not fail, the contract will return that verdict.

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

**ParameterSets:** ParameterSets are a list of Parameters. For every ParameterSet a copy of the contract is created with
the parameters of that set.

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

### Supported Validations

The following OpenAPI Schema attributes are currently validated:

- type: object, array, integer, string, number, boolean
- properties
- items
- oneOf, anyOf, allOf
- nullable
- required
- format: string.uri
