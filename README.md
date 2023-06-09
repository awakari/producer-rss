# Contents

1. [Overview](#1-overview)<br/>
   1.1. [Purpose](#11-purpose)<br/>
   1.2. [Definitions](#12-definitions)<br/>
2. [Configuration](#2-configuration)<br/>
3. [Deployment](#3-deployment)<br/>
   3.1. [Prerequisites](#31-prerequisites)<br/>
   3.2. [Bare](#32-bare)<br/>
   3.3. [Docker](#33-docker)<br/>
   3.4. [K8s](#34-k8s)<br/>
   &nbsp;&nbsp;&nbsp;3.4.1. [Helm](#341-helm) <br/>
4. [Usage](#4-usage)<br/>
5. [Design](#5-design)<br/>
   5.1. [Requirements](#51-requirements)<br/>
   5.2. [Approach](#52-approach)<br/>
   5.3. [Limitations](#53-limitations)<br/>
6. [Contributing](#6-contributing)<br/>
   6.1. [Versioning](#61-versioning)<br/>
   6.2. [Issue Reporting](#62-issue-reporting)<br/>
   6.3. [Building](#63-building)<br/>
   6.4. [Testing](#64-testing)<br/>
   &nbsp;&nbsp;&nbsp;6.4.1. [Functional](#641-functional)<br/>
   &nbsp;&nbsp;&nbsp;6.4.2. [Performance](#642-performance)<br/>
   6.5. [Releasing](#65-releasing)<br/>

# 1. Overview

## 1.1. Purpose

The example Awakari Producer implementation that reads new items from RSS feeds and converts these to the 
[cloud events](https://cloudevents.io/) for the further processing by the Awakari Core system.

## 1.2. Definitions

The component is [RSS](https://en.wikipedia.org/wiki/RSS)-specific messages producer. It works with RSS feeds and items.
An RSS feed is a subject of periodic fetching to get new RSS items. RSS items are being converted to the Cloud Events.

# 2. Configuration

The service is configurable using the environment variables:

| Variable                    | Example value                                            | Description                                                                                 |
|-----------------------------|----------------------------------------------------------|---------------------------------------------------------------------------------------------|
| API_WRITER_BACKOFF          | `10s`                                                    | Time to sleep before a retry when resolver fails to accept all messages                     |
| API_WRITER_BATCH_SIZE       | `64`                                                     | Defines the max size of the messages batch to be flushed to the writer                      |
| API_WRITER_URI              | `writer:50051`                                           | [Writer](https://github.com/awakari/writer) dependency service URI                          |
| DB_URI                      | `mongodb://localhost:27017/?retryWrites=true&w=majority` | DB URI                                                                                      |
| DB_NAME                     | `producer-rss`                                           | DB name                                                                                     |
| DB_USERNAME                 | `root`                                                   | DB authentication: user name                                                                |
| DB_PASSWORD                 | `********`                                               | DB authentication: passwod                                                                  |
| DB_TABLE_NAME               | `feeds`                                                  | Table name for the feeds update timestamps                                                  |
| DB_TLS_ENABLED              | `false`                                                  | Defines whether to use TLS to connect the DB. Should be `true` when cloud DB is used.       |
| DB_TLS_INSECURE             | `false`                                                  | Defines whether to skip the server TLS certificate check when TLS is used to connect the DB |
| LOG_LEVEL                   | `-4`                                                     | [Logging level](https://pkg.go.dev/golang.org/x/exp/slog#Level)                             |
| FEED_URL                    | `https://techcrunch.com/feed `                           | Feed URL to fetch and update                                                                |
| FEED_TLS_SKIP_VERIFY        | `true`                                                   | Defines whether producer should skip the TLS certificate check when fetching the RSS feed   |
| FEED_UPDATE_INTERVAL_MIN    | `10s`                                                    | Minimum possible feed update interval                                                       |
| FEED_UPDATE_INTERVAL_MAX    | `10m`                                                    | Maximum pssible feed update interval                                                        |
| FEED_UPDATE_TIMEOUT         | `1m`                                                     | Timeout to fetch the RSS feed                                                               |
| FEED_USER_AGENT             | `awakari-producer-rss/0.0.1`                             | HTTP user agent to use to fetch any RSS feed                                                |
| MSG_MD_KEY_FEED_CATEGORIES  | `feedcategories`                                         | Cloud Event attribute name to use for the feed categories                                   |
| MSG_MD_KEY_FEED_DESCRIPTION | `feeddescription`                                        | Cloud Event attribute name to use for the feed description                                  |
| MSG_MD_KEY_FEED_IMAGE_TITLE | `feedimagetitle`                                         | Cloud Event attribute name to use for the feed image title                                  |
| MSG_MD_KEY_FEED_IMAGE_URL   | `feedimageurl`                                           | Cloud Event attribute name to use for the feed image URL                                    |
| MSG_MD_KEY_FEED_TITLE       | `feedtitle`                                              | Cloud Event attribute name to use for the feed title                                        |
| MSG_MD_KEY_AUTHOR           | `author`                                                 | Cloud Event attribute name to use for the RSS item author                                   |
| MSG_MD_KEY_CATEGORIES       | `categories`                                             | Cloud Event attribute name to use for the RSS item categories                               |
| MSG_MD_KEY_GUID             | `rssitemguid`                                            | Cloud Event attribute name to use for the RSS item GUID                                     |
| MSG_MD_KEY_IMAGE_TITLE      | `imagetitle`                                             | Cloud Event attribute name to use for the RSS item image title                              |
| MSG_MD_KEY_IMAGE_URL        | `imageurl`                                               | Cloud Event attribute name to use for the RSS item image URL                                |
| MSG_MD_KEY_TITLE            | `title`                                                  | Cloud Event attribute name to use for the RSS item title                                    |
| MSG_MD_KEY_LANGUAGE         | `language`                                               | Cloud Event attribute name to use for the RSS item language                                 |
| MSG_MD_KEY_SUMMARY          | `summary`                                                | Cloud Event attribute name to use for the RSS item summary                                  |
| MSG_CONTENT_TYPE            | `text/plain`                                             | Cloud Event attribute name to use for the message content type                              |

The only command line argument is the path to the file that is used to load the list of the feed URLs.
Example file is located at [config/feed-urls.txt](config/feed-urls.txt).

# 3. Deployment

## 3.1. Prerequisites

The Awakari Core system should be deployed.
The producer uses the [writer](https://github.com/awakari/writer) service as an entry point.
The API gateway is not required as far as the producer-rss is internal and owned component.

## 3.2. Bare

Preconditions:
1. Build patterns executive using ```make build```
2. Port-forward the writer service to local from a test cluster

Then run the command:
```shell
API_WRITER_URI=localhost:50051 \
FEED_URL=https://hnrss.org/newest \
./producer-rss
```

## 3.3. K8s

Note the producer generally requires the custom network policy to be able to fetch the specified feeds.
See the [helm/producer-rss/templates/egress.yaml](helm/producer-rss/templates/egress.yaml) source file for the details.

### 3.3.1. Helm

Create a helm package from the sources:
```shell
helm package helm/producer-rss/
```

Install the helm chart:
```shell
helm install -n awakari producer-rss ./producer-rss-<CHART_VERSION>.tgz \
  --values helm/api/values-db-uri.yaml
```
where `<CHART_VERSION>` is the helm chart version

This deploys a set of predefined cronjobs, each per specific RSS feed.

# 4. Usage

The component is a job, so it doesn't serve any API.

# 5. Design

## 5.1. Requirements

| #     | Summary                                        | Description    |
|-------|------------------------------------------------|----------------|
| REQ-1 | Produce Cloud Events                           | From RSS items |
| REQ-2 | Store the latest update time                   | Per RSS feed   |
| REQ-3 | Use only new RSS items to produce the messages | Best effort    |

## 5.2. Approach

To produce the messages based on the new RSS items only, it's necessary to save the latest timestamp between producer
job runs. The state

### 5.2.1. Data Schema

| Attribute | Type    | Description                    |
|-----------|---------|--------------------------------|
| url       | String  | RSS feed URL, unique           |
| ts        | Integer | Last RSS feed update time, UTC |


## 5.3. Limitations

TODO

# 6. Contributing

## 6.1. Versioning

The service uses the [semantic versioning](http://semver.org/).
The single source of the version info is the git tag:
```shell
git describe --tags --abbrev=0
```

## 6.2. Issue Reporting

TODO

## 6.3. Building

```shell
make build
```
Generates the sources from proto files, compiles and creates the `producer-rss` executable.

## 6.4. Testing

### 6.4.1. Functional

```shell
make test
```

### 6.4.2. Performance

TODO

## 6.5. Releasing

To release a new version (e.g. `1.2.3`) it's enough to put a git tag:
```shell
git tag -v1.2.3
git push --tags
```

The corresponding CI job is started to build a docker image and push it with the specified tag (+latest).
