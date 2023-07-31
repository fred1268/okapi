# okapi :giraffe:

API tests made as easy as table driven tests.

## Introduction

okapi is a program allowing you to test your APIs by using tests files and test cases, pretty much like the `go test` command with table-driven tests. okapi will iterate on all `.test.json` files in the specified directory and runs every test case containted within the files, sequentially or in parallel.

The response of each test case is then compared to the expected response (both the HTTP Response Status Code, as well as the payload). Success or failure are reported.

## Installation

Pretty easy to install:

```shell
go install github.com/fred1268/okapi@latest
```

> Please note that okapi does have a single dependency: [**clap :clap:, the Command Line Argument Parser**](https://github.com/fred1268/go-clap), which makes it very easy to parse okapi command line. clap is a lightweight, non intrusive command line parsing library you may want to try out in your own projects. Feel free to give it a try!

## Configuring okapi :giraffe:

In order to run, okapi requires the following files:

- a configuration file
- one or more test files
- zero or more payload files
- zero or more response files

### Configuration file

The configuration file looks like the following:

```json
{
  "exampleserver1": {
    "host": "http://localhost:8080",
    "auth": {
      "login": {
        "method": "POST",
        "endpoint": "http://localhost:8080/login",
        "payload": "{\"email\":\"test@test.com\",\"password\":\"${env:MY_PASSWORD}\"}",
        "expected": {
          "statuscode": 200
        }
      },
      "session": {
        "cookie": "jsessionid"
      }
    }
  },
  "exampleserver2": {
    "host": "http://localhost:9090",
    "auth": {
      "login": {
        "method": "POST",
        "endpoint": "http://localhost:8080/login",
        "payload": "{\"email\":\"test@test.com\",\"password\":\"${env:MY_PASSWORD}\"}",
        "expected": {
          "statuscode": 200
        }
      },
      "session": {
        "jwt": "header"
      }
    }
  },
  "exampleserver3": {
    "host": "http://localhost:8088",
    "auth": {
      "apikey": {
        "apikey": "Bearer: ${env:MY_APIKEY}",
        "header": "Authorization"
      }
    }
  },
  "hackernews": {
    "host": "https://hacker-news.firebaseio.com"
  }
}
```

A Server description contains various fields:

- `host` (mandatory): the URL of the server (including port and everything you don't want to repeat on every test)

- `auth.login`: used for login/password authentication, using the same format as a test (see below)

- `auth.session.cookie`: used for cookie session management, name of the cookie maintaining the session

- `auth.apikey`: used for API Key authentication, contains both the API Key and the required header

- `auth.session.jwt`: used for JWT session management (`header`, `payload` or `payload.xxx`)

Here `exampleserver1` uses the `/login` endpoint on the same HTTP server than the one used for the tests. Both `email` and `password` are submitted in the `POST`, and `200 OK` is expected upon successful login. The session is maintained by a session cookie called `jsessionid`.

The second server, `exampleserver2` also uses the `/login` endpoint, but on a different server, hence the endpoint with a different server. The sesssion is maintained using a JWT (JSON Web Token) which is obtained though a header (namely `Authorization`). Should your JWT be returned as a payload, you can specify `"payload"` instead of `"header"`. You can even use `payload.token` for instance, if your JWT is returned in a `token` field of a JSON object. JWT is always sent back using the `Authorization` header in the form of `Authorization: Bearer my_jwt`.

> Please note that in the case of the server definition, `endpoint` must be an fully qualified URL, not a relative endpoint like in the test files. Thus `endpoint` must start with `http://` or `https://`.

The third server, `exampleserver3` uses API Keys for authentication. The apikey field contains the key itself, whereas the header field contains the field used to send the API Key back (usually `Authorization`). Please note that session is not maintained in this example, since the API Key is sent with each request.

The last server, `hackernews`, is a server which doesn't require any authentication.

> _Environment variable substitution_: please note that `host`, `apikey`, `endpoint` and `payload` can use environment variable substitution. For example, instead of hardcoding your API Key in your server configuration file, you can use `${env:MY_APIKEY}` instead. Upon startup, the `${env:MY_APIKEY}` text will be replaced by the value of `MY_APIKEY` environment variable (i.e. `$MY_APIKEY` or `%MY_APIKEY%`).

### Test files

A test file looks like the following:

```json
{
  "tests": [
    {
      "name": "121003",
      "server": "hackernews",
      "method": "GET",
      "endpoint": "/v0/item/121003.json",
      "expected": {
        "statuscode": 200
      }
    },
    {
      "name": "cap121004",
      "server": "hackernews",
      "method": "GET",
      "endpoint": "/v0/item/121004.json",
      "capture": true,
      "expected": {
        "statuscode": 200
      }
    },
    {
      "name": "121004",
      "server": "hackernews",
      "method": "GET",
      "endpoint": "/v0/item/${cap121004.id}.json",
      "expected": {
        "statuscode": 200,
        "response": "{\"id\":${cap121004.id}}"
      }
    },
    {
      "name": "121005",
      "server": "hackernews",
      "method": "GET",
      "endpoint": "/v0/item/121005.json",
      "expected": {
        "statuscode": 200,
        "response": "@file"
      }
    },
    {
      "name": "doesnotwork",
      "server": "hackernews",
      "method": "POST",
      "endpoint": "/v0/item",
      "payload": "@custom_filename.json",
      "expected": {
        "statuscode": 200
      }
    },
    {
      "name": "121007",
      "server": "hackernews",
      "method": "GET",
      "endpoint": "/v0/item/121007.json",
      "expected": {
        "statuscode": 200,
        "response": "\\s+[wW]eight"
      }
    }
  ]
}
```

> The test files must end with `.test.json` in order for okapi to find them. A good pratice is to name them based on your routes. For example, in this case, since we are testing hackernews' `item` route, the file could be named `item.test.json` or `item.get.test.json` if you need to be more specific.

A test file contains an array of tests, each of them containing:

- `name` (mandatory): a unique name to globally identify the test (test name must not contain the `. (period)` character)

- `server` (mandatory): the name of the server used to perform the test (declared in the configuration file)

- `method` (mandatory): the method to perform the operation (`GET`, `POST`, `PUT`, etc.)

- `endpoint` (mandatory): the endpoint of the operation (usually a ReST API of some sort)

- `capture` (default false): true if you want to capture the response of this test so that it can be used in another test in this file (fileParallel mode only)

- `skip` (default false): true to have okapi skip this test (useful when debugging a script file)

- `payload` (default none): the payload to be sent to the endpoint (usually with a POST, PUT or PATCH method)

- `expected`: this section contains:

  - `statuscode` (mandatory): the expected status code returned by the endpoint (200, 401, 403, etc.)

  - `response` (default none): the expected payload returned by the endpoint.

> Please note that `payload` and `response` can be either a string (including json, as shown in 121004), or `@file` (as shown in 121005) or even a `@custom_filename.json` (as shown in doesnotwork). This is useful if you prefer to separate the test from its `payload` or expected `response` (for instance, it is handy if the `payload` or `response` are complex JSON structs that you can easily copy and paste from somewhere else, or simply prefer to avoid escaping double quotes). However, keeping the names for `payload` and `response` like `test_name.payload.json`and `test_name.expected.json` is still a good practice.

> Please also note that `endpoint` and `payload` can use environment variable substitution using the ${env:XXX} syntax (see previous note about environment variable substitution).

> Lastly, please note that `endpoint` and `response` can use captured variable (i.e. variables inside a captured response, see `"capture":true`). For instance, to use the `id` field returned inside of a `user` object in test `mytest`, you will use `${mytest.user.id}`. In the example above, we used `${cap121004.id}` to retrieve the ID of the returned response in test `cap121004`. Captured response also works with arrays.

### Payload and Response files

Payload and response files don't have a specific format, since they represent whatever the server you are testing is expecting from or returns to you. The only important things to know about the payload and response files, is that they must be placed in the test directory, and must be named `<name_of_test>.payload.json` and `<name_of_test>.expected.json` (`121005.expected.json` in the example above) respectively if you specify `@file`. Alternatively, they can also be put in a `payload/` or `expected/` subdirectory of the test directory, and, in that case, be named `<name_of_test>.json`. If you decide to use a custom filename for your `payload` and/or `response`, then you can specify the name of your choice prefixed by `@` (`@custom_filename.json` in the example above).

## Expected response

As we saw earlier, for each test, you will have to define the expected response. okapi will always compare the HTTP Response Status Code with the one provided, and can optionally, compare the returned payload. The way it works is pretty simple:

- if the response is in JSON format:
  - if a field is present in `expected`, okapi will also check for its presence in the response
  - if the response contains other fields not mentioned in `expected`, they will be ignored
  - success or failure is reporting accordingly
- if the response is a non-JSON string:
  - the response is compared to `expected` and success or failure is reported

> Please note that, in the case of non-JSON responses, you can use regular expressions (see test 121007). In that case, make sure the `expected.response` field is set to a proper, compilable, regular expression. Be mindful that you will need to escape the `\ (backslash)` character using `\\`. For instance `\s+[wW]eight` will be written `\\s+[wW]eight`, in order to match one or more whitespace characters, followed by weight or Weight.

## Running okapi :giraffe:

To launch okapi, please run the following:

```shell
    okapi [options] <test_directory>
```

where options are one or more of the following:

- `--servers-file`, `-s` (mandatory): point to the configuration file's location

- `--verbose`, `-v` (default no): enable verbose mode

- `--file-parallel` (default no): run the test files in parallel (instead of the tests themselves)

- `--file`, `-f` (default none): only run the specified test file

- `--test`, `-t` (default none): only run the specified standalone test

- `--timeout` (default 30s): set a default timeout for all HTTP requests

- `--no-parallel` (default parallel): prevent tests from running in parallel

- `--user-agent` (default okapi UA): set the default user agent

- `--content-type` (default application/json): set the default content type for requests

- `--accept` (default application/json): set the default accept header for responses

- `test_directory` (mandatory): point to the directory where all the test files are located

> Please note that the `--file-parallel` mode is particularly handy if you want to have a sequence of tests that needs to run in a specific order. For instance, you may want to create a resource, update it, and delete it. Placing these three tests in the same file and in the right order, and then running okapi with `--file-parallel` should do the trick. The default mode is used for unit tests, whereas the `--file-parallel` mode is used for (complex) test scenarios.

## Output example

To try the included examples (located in `./assets/tests`), you need to run the following command:

```shell
$ okapi --servers-file ./assets/config.json --verbose ./assets/tests
```

You should then see the following output:

```shell
--- PASS:       hackernews.items.test.json
    --- PASS:   121014 (0.35s)
    --- PASS:   121012 (0.36s)
    --- PASS:   121010 (0.36s)
    --- PASS:   121004 (0.36s)
    --- PASS:   121007 (0.36s)
    --- PASS:   121009 (0.36s)
    --- PASS:   121006 (0.36s)
    --- PASS:   121011 (0.36s)
    --- PASS:   121008 (0.36s)
    --- PASS:   121013 (0.36s)
    --- PASS:   121005 (0.36s)
    --- PASS:   121003 (0.36s)
PASS
ok      hackernews.items.test.json                      0.363s
--- PASS:       hackernews.users.test.json
    --- PASS:   jk (0.36s)
    --- PASS:   jo (0.36s)
    --- PASS:   jc (0.36s)
    --- PASS:   401 (0.36s)
    --- PASS:   jl (0.37s)
PASS
ok      hackernews.users.test.json                      0.368s
okapi total run time: 0.368s
```

## Integrating okapi :giraffe: with your own software

okapi exposes a pretty simple and straightforward API that you can use within your own Go programs.

You can find more information about the exposed API on [The Official Go Package](https://pkg.go.dev/github.com/fred1268/okapi/testing).

## Feedback and contribution

Feel free to send feedback, PR, issues, etc.
