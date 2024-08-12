# BBFS Server

A simple http server that serves the content of a Bitbucket Server server via HTTP.

This version only supports the Bitbucket Server API, latest.

It uses caching to minimize the load on the Bitbucket server. 
You need to restart the server to clear the cache.

It has no dedicated probes, but you can use / as startup, liveness and readiness probe.

```
Usage: bbfsserver
Runs a webserver on top of a bitbucket repos on a Bitbucket Server.

Environment variables
    PORT                        listen port, defaults to 8080
    BBFSSRV_LISTEN_ADDRESS      listen address, this allows you to specify the ip address to listen on, default to ":8080"
    BBFSSRV_HOST                Bitbucket server host
    BBFSSRV_PROJECT_KEY         Bitbucket project key or user id
    BBFSSRV_REPOSITORY_SLUG     Bitbucket repository name
    BBFSSRV_ACCESS_KEY          Bitbucket http access key for the repo or project
    BBFSSRV_LOG_FORMAT          log format [ text | json], defaults to json
```

## Used tools

This project uses devbox to install the tools:
* mage
* ko
* crane

It also needs go to be installed. If this is not the case, you can add it to devbox with ```devbox add go```.
