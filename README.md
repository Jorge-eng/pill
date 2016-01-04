Go tool to process pill uploads


`bin/pill-osx {command}`

```
usage: app [--version] [--help] <command> [<args>]

Available commands are:
    check       processes all zip files in current directory
    clean       warning: deletes zip directory in current directory
    download    downloads all zip files starting with given prefix from s3
    local       does something locally
    process     processes all zip files in given directory
    search      search for deviceId in all zip files in given directory
    single      processes the given zip file
```

To build: you need to have the pill factory key. Ping Tim or Jackson.

Assumes you have properly configured AWS creds with access to S3 bucket.