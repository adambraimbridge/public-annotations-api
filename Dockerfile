FROM alpine:3.5

ENV SOURCE_DIR /public-annotations-api-src

COPY *.go .git $SOURCE_DIR/
COPY annotations/*.go $SOURCE_DIR/annotations/
COPY vendor/vendor.json $SOURCE_DIR/vendor/

RUN apk add --no-cache  --update bash ca-certificates \
  && apk --no-cache --virtual .build-dependencies add git go libc-dev \
  && cd $SOURCE_DIR \
  && BUILDINFO_PACKAGE="github.com/Financial-Times/service-status-go/buildinfo." \
  && VERSION="version=$(git describe --tag --always 2> /dev/null)" \
  && DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)" \
  && REPOSITORY="repository=$(git config --get remote.origin.url)" \
  && REVISION="revision=$(git rev-parse HEAD)" \
  && BUILDER="builder=$(go version)" \
  && LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'" \
  && cd .. \
  && export GOPATH=/gopath \
  && REPO_PATH="github.com/Financial-Times/public-annotations-api" \
  && mkdir -p $GOPATH/src/${REPO_PATH} \
  && cp -r $SOURCE_DIR/* $GOPATH/src/${REPO_PATH} \
  && cd $GOPATH/src/${REPO_PATH} \
  && echo ${LDFLAGS} \
  && go get -u github.com/kardianos/govendor \
  && $GOPATH/bin/govendor sync \
  && go build -ldflags="${LDFLAGS}" \
  && mv public-annotations-api / \
  && apk del .build-dependencies \
  && rm -rf $GOPATH /var/cache/apk/*

CMD [ "/public-annotations-api" ]
