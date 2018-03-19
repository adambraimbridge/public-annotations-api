# Public API for Annotations (public-annotations-api)
[![Circle CI](https://circleci.com/gh/Financial-Times/public-annotations-api.svg?style=shield)](https://circleci.com/gh/Financial-Times/public-annotations-api)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/public-annotations-api)](https://goreportcard.com/report/github.com/Financial-Times/public-annotations-api) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/public-annotations-api/badge.svg)](https://coveralls.io/github/Financial-Times/public-annotations-api)
__Provides a public API for Annotations stored in a Neo4J graph database__

## Installation & running locally

* `go get -u github.com/kardianos/govendor`
* `go get -u github.com/Financial-Times/public-annotations-api`
* `cd $GOPATH/src/github.com/Financial-Times/public-annotations-api`
* `govendor sync`
* `go test ./...`
* `go install`
* `$GOPATH/bin/public-annotations-api --neo-url={neo4jUrl} --port={port} --log-level={DEBUG|INFO|WARN|ERROR}--cache-duration{e.g. 22h10m3s}`   
_Optional arguments are:
--neo-url defaults to http://localhost:7474/db/data, which is the out of box url for a local neo4j instance.
--port defaults to 8080.
--cache-duration defaults to 1 hour._
* `curl http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations | json_pp`
* Or using [httpie](https://github.com/jkbrzt/httpie) `http GET http://localhost:8080/content/143ba45c-2fb3-35bc-b227-a6ed80b5c517/annotations`

## Build & deployment
Continuosly built be CircleCI. The docker image of the service is built by Dockerhub based on the git release tag. 
To prepare a new git release, go to the repo page on GitHub and create a new release.
* Docker Hub builds: [coco/public-annotations-api](https://hub.docker.com/r/coco/public-annotations-api/)
* Cluster deployment:  [public-annotations-api](https://upp-k8s-jenkins.in.ft.com/job/k8s-deployment/job/apps-deployment/job/public-annotations-api-auto-deploy/)
* CI provided by CircleCI: [public-annotations-api](https://circleci.com/gh/Financial-Times/public-annotations-api)
* Code coverage provided by Coverall: [public-annotations-api](https://coveralls.io/github/Financial-Times/public-annotations-api)

## API definition
Based on the following [google doc](https://docs.google.com/a/ft.com/document/d/1kQH3tk1GhXnupHKdDhkDE5UyJIHm2ssWXW3zjs3g2h8/edit?usp=sharing)

### GET content/{uuid}/annotations endpoint
Returns all annotations for a given uuid of a piece of content in json format.

*Please note* that   
* the `public-annotations-api` will return more brands than the ones the article has been annotated with. 
This is because it will return also the parent of the brands from any brands annotations. 
If those brands have parents, then they too will be brought into the result.

* the `public-annotations-api` curated (tag-me) annotations (life cycle pac) for a piece of content take precedence, if present they are returned, all non-pac lifecycle annotations are omitted .
If there are no pac life cycle annotations, non-pac annotations will be returned. The filtering described in the next paragraph relates to non-pac annotations.

* the `public-annotations-api` will filter out less important annotations if a more important annotation is also present for the same concept.  
_For example_, if a piece of content is annotated with a concept with "About", "Major Mentions" and "Mentions" relationships 
only the annotation with "About" relationship will be returned.    
Similarly if a piece of content is annotated with a Concept "Is Classified By" and "Is Primarily Classified By"
only the annotation with "Is Primarily Classified By" relationship will be returned.

### GET content/{uuid}/annotations/{platformVersion} endpoint

This endpoint returns all the existing annotations for a specific platformVersion - if any.
Note:
The response here is an enriched format of the simple /content/{uuid}/annotations response, containing fields like `platformVersion`, and the referenced concepts' identifiers.
This endpoint does not show inferred Brands annotations, as the other endpoint does.

### Response Example
The structure of the the response is the same both endpoints and would look like this:
```
[...
    {
        predicate: "http://www.ft.com/ontology/classification/isClassifiedBy",
        id: "http://api.ft.com/things/{concepts_canonical_uuid}",
        apiUrl: "http://api.ft.com/things/{concepts_canonical_uuid}",
        types: [
            "http://www.ft.com/ontology/core/Thing",
            "http://www.ft.com/ontology/concept/Concept",
            "http://www.ft.com/ontology/classification/Classification",
            "http://www.ft.com/ontology/Subject"
        ],
        prefLabel: "Company News",
        leiCode: "leicode_value",
        FIGI: "figi_value",
        factsetID: "factsetID_value"
        tmeIDs: [
            "tmeid__value"
        ],
        uuids: [
            "uuid1","uuid2","canonical_uuid"
        ],
        platformVersion: "v1",
    },
...
]
```

## Admin endpoints

Healthchecks: [http://localhost:8080/__health](http://localhost:8080/__health)  
Build Info: [http://localhost:8080/__build-info](http://localhost:8080/__build-info)  
GTG: [http://localhost:8080/__gtg](http://localhost:8080/__gtg)

### Logging
The application uses logrus, the logfile is initialised in main.go.

Logging requires an env app parameter: for all environments other than local, logs are written to file. When running locally logging
is written to console (if you want to log locally to file you need to pass in an env parameter that is != local).

NOTE: http://localhost:8080/__gtg end point is not logged as it is called every second from varnish and this information is not needed in logs/splunk
