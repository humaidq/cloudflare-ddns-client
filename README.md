# Cloudflare DDNS Client

## 1. Description

This is a simple Cloudflare Dynamic DNS client, which updates a record to the
current public IP address. I have created this as I couldn't find a very simple
solution to this very simple problem.

## 2. Requirements

The following packages must be installed on your system.

- Go *(tested with 1.14)*

## 3. Installation and Usage

```
$ go install git.sr.ht/~humaid/cloudflare-ddns-client
$ export APIKey=<API KEY here>
$ export Zone=<Zone ID here>
$ export RecordName=<Full record name here, e.g. foo.example.com>
$ cloudflare-ddns-client
```

You will have to [create an API token](https://dash.cloudflare.com/profile/api-tokens)
which has the permission to edit zone DNS (`Zone.DNS`).

This program will check every minute unless interrupted.
