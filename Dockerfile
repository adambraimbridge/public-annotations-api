FROM golang:1.10-alpine

# Set up our extra bits in the image
RUN apk --no-cache add git curl
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

# Set up app information in the environment
ENV PROJECT="public-annotations-api"
ENV ORG_PATH="github.com/Financial-Times"
ENV SRC_FOLDER="${GOPATH}/src/${ORG_PATH}/${PROJECT}"
ENV BUILDINFO_PACKAGE="${ORG_PATH}/${PROJECT}/vendor/${ORG_PATH}/service-status-go/buildinfo."
ENV VERSION="version=$(git describe --tag --always 2> /dev/null)"
ENV DATETIME="dateTime=$(date -u +%Y%m%d%H%M%S)"
ENV REPOSITORY="repository=$(git config --get remote.origin.url)"
ENV REVISION="revision=$(git rev-parse HEAD)"
ENV BUILDER="builder=$(go version)"
ENV LDFLAGS="-X '"${BUILDINFO_PACKAGE}$VERSION"' -X '"${BUILDINFO_PACKAGE}$DATETIME"' -X '"${BUILDINFO_PACKAGE}$REPOSITORY"' -X '"${BUILDINFO_PACKAGE}$REVISION"' -X '"${BUILDINFO_PACKAGE}$BUILDER"'"
RUN echo "Build flags: $LDFLAGS"

# Include code
COPY . ${SRC_FOLDER}
WORKDIR ${SRC_FOLDER}

# Install dependancies and build app
RUN $GOPATH/bin/dep ensure -vendor-only
RUN go build -o /artifacts/${PROJECT} -ldflags="${LDFLAGS}"



# Multi-stage build - copy only the certs and the binary into the image
FROM scratch
WORKDIR /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=0 /artifacts/public-annotations-api /

CMD [ "/public-annotations-api" ]
