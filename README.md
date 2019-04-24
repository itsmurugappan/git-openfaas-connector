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
* Build the code
```
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o release/git-faas .
```
* Build image
```
docker build -t git-faas  .
```

#### Deploy connector

1. provide the following environment variables in the deployment
2. Secret token  (some string for git webhook validation)
3. repo url (key) and fn name (value)
4. api path prefix like http://gateway:8080/function/


#### Github Set Up

1. Go to Hooks sections of settings page in the git repo
2. Add webhook
3. Enter the payload url, secret specified in the deployment
4. Content type is application/json
5. Select the appropriate git events.

This connector can be used for any protected api, just mention the correct path in the api environment variable
