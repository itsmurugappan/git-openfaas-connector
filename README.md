## GIT - OpenFaas Connector

Trigger protected openfaas functions from github webhooks.

#### Architecture

The connector validates the secret send by github and invokes the function based
on repo/function mapping in the environment variable
```
       +++++++++++++++++++++++++++++++++++++++++
       +                                       +
  git--+--> GitOpenFaaSConnector --> Function  +
       +                                       +
       +        (K8s/Openshift Namespace)      +
       +++++++++++++++++++++++++++++++++++++++++
```
#### Build code

* All the code is in git-faas.go file.
* Download dependencies
```
dep ensure
```
* Test
```
go test
```
* Build the code
```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o release/git-faas .
```
* Build image
```
docker build -t git-faas  .
```

#### Deploy connector

Following are the configuration options as environment variables

| Key | Value | Default | Description |
|-----|-------|---------|-------------|
| secret | token string | no default - mandatory field | the secret token to validate git webhooks |
| api | api prefix | http://gateway:8080/function/ | api prefix path like http://gateway:8080/function/ |
| git url | function | no default - mandatory field | please specify your git url as key and function name as the value |
| retry | number of retries | 3 | number of time function should be retried |
| timeout | Transport dialer time out | 100s | timeout duration like 50s |


#### Github Set Up

1. Go to Hooks sections of settings page in the git repo
2. Add webhook
3. Enter the payload url, secret specified in the deployment
4. Content type is application/json
5. Select the appropriate git events.

This connector can be used for any protected api, just mention the correct path in the api environment variable
