# Hachyboop! 👉🐘

Testing utility to poke hachyderm.io from various angles. Writes data to parquet for later bulk analysis.

## Overview

Hachyboop performs the following types of tests:

- DNS checks - how does DNS resolve for hachyderm.io, a geographically-dependent A/AAAA record?
- (coming one day) HTTP checks - how does hachyderm.io respond and perform for basic connection checks?
- (coming another day, not pictured) Websocket checks - how does the streaming portion of Mastodon perform?

```
┌─────────────────┐                                                                                           
│                 │                                                                                           
│  Hachyderm.io   │                                                                                           
│  Authoritative  │◄─────┐        ┌───────────────────────────────────┐       ┌──────────────────────────────┐
│  DNS Server     │      │        │                                   │       │                              │
│                 │      │        │  Agent Node                       │       │  Hachyderm Cloud             │
└─────────────────┘      │        │                                   │       │                              │
                    DNS  │        │   ┌─────────────┐                 │       │  ┌────────────────────────┐  │
                    boop ├────────┼───┤             │  write results  │       │  │                        │  │
┌─────────────────┐      │        │   │  Hachyboop  ├─────────┬───────┼───────┼─►│  Hachyderm S3 Storage  │  │
│                 │      │  ┌─────┼───┤             │         │       │       │  │                        │  │
│  Hachyderm.io   │      │  │     │   └─────────────┘         │       │       │  └────────────────────────┘  │
│  DNS Migration  │◄─────┘  │     │                           │       │       │              ▲               │
│  Target Server  │         │     │   ┌─────────────────┐     │       │       │              │               │
│                 │         │     │   │                 │     │       │       └──────────────┼───────────────┘
└─────────────────┘         │     │   │  Local Parquet  │◄────┘       │                      │                
                            │     │   │                 │             │              analyze │                
                            │     │   └─────────────────┘             │              (duckdb)│                
┌─────────────────┐         │     │                                   │                      │                
│                 │    HTTP │     │                                   │          ┌───────────┴────────────┐   
│                 │    boop │     └───────────────────────────────────┘          │                        │   
│  Hachyderm.io   │◄────────┘                                                    │  Hachyderm Infra Team  │   
│                 │                                                              │                        │   
│                 │                                                              └────────────────────────┘   
└─────────────────┘                                                                                           
```

> [!IMPORTANT]  
> Hachyboop DOES NOT and WILL NEVER collect any personal data or machine data beyond the data you explicitly provide us, the answers to DNS queries, and performance data related to the tests. By default, Hachyboop does not write results to Hachyderm S3 Storage.

## Configuration

Hachyboop provides several environment variables to configure how it behaves. If you use `direnv`, you can copy the [.envrc.example](./envrc.example), rename it to `.envrc`, then fill it with your values.

```bash
##################################################
# Hachyboop Common Environment Variables
##################################################

## Metadata

### Identifying metadata - helps us understand who you are (anonymized OK) and where you're coming from (broadly)
export HACHYBOOP_OBSERVER_ID=esk  # unique (or not) identifier for you. recommendation is you provide a pseudonym or generated unique ID.
export HACHYBOOP_OBSERVER_REGION=exandria  # see below, region code that most closely matches where you are

## DNS Configuration

### Resolvers and questions - what DNS entries are you poking, and who are you asking?
export HACHYBOOP_RESOLVERS=91.200.176.1:53,8.8.8.8:53  # comma-separated, requires port number, e.g. 8.8.8.8:53
export HACHYBOOP_QUESTIONS=hachyderm.io  # which records do you want to test

## Writers

### S3 writer - writes parquet to the specified S3 bucket
export HACHYBOOP_S3_WRITER_ENABLED=true  # should we write to S3?
export HACHYBOOP_S3_ENDPOINT=fsn1.your-objectstorage.net  # S3 endpoint
export HACHYBOOP_S3_BUCKET=bag-of-holding  # S3 bucket name
export HACHYBOOP_S3_PATH=some-stuff/some-subfolder # what path to write to on S3
export HACHYBOOP_S3_ACCESS_KEY_ID=AKIAEXAMPLE  # access key
export HACHYBOOP_S3_SECRET_ACCESS_KEY=secret  # secret

### Local file writer - writes parquet to a local file on disk
export HACHYBOOP_LOCAL_WRITER_ENABLED=true  # should we write to the local disk
export HACHYBOOP_LOCAL_RESULTS_PATH=data  # path to write to, relative to where hachyboop is running
```

### `HACHYBOOP_OBSERVER_REGION` values

When choosing a value, choose the value that most closely matches how you would describe your location.

- North America
  - namer-east
  - namer-central
  - namer-west
- South America
  - samer-east
  - samer-west
- Europe
  - eu-east
  - eu-central
  - eu-west
- Africa
  - africa-north
  - africa-south
- Middle East
  - me
- Asia
  - ap-south
  - ap-southeast
  - ap-southwest
  - ap-north
  - ap-northeast
  - ap-northwest
  - china
  - japan
- Australia
  - australia-east
  - australia-west

## Example of collected data

> [!IMPORTANT]  
> Repeating this, because it's important. Hachyboop DOES NOT and WILL NEVER collect any personal data or machine data beyond the data you explicitly provide us, the answers to DNS queries, and performance data related to the tests. By default, Hachyboop does not write results to Hachyderm S3 Storage.

Here's a screenshot of the data we collect with `hachyboop`. The `observedby` and `observationregion` fields were supplied via the `HACHYBOOP_*` environment variables above.

![image](https://github.com/user-attachments/assets/3e1a8ddf-7777-4336-8139-b233e53839c6)

  

