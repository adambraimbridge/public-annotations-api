module github.com/Financial-Times/public-annotations-api/v3

go 1.12

require (
	github.com/Financial-Times/annotations-rw-neo4j/v3 v3.2.0
	github.com/Financial-Times/base-ft-rw-app-go v0.0.0-20170831124905-0800bf5938ec
	github.com/Financial-Times/concepts-rw-neo4j v0.0.0-20180530124205-d49a23fd8c0b
	github.com/Financial-Times/content-rw-neo4j v1.0.3-0.20170901141716-315eb09dc04b
	github.com/Financial-Times/go-fthealth v0.0.0-20180807113633-3d8eb430d5b5
	github.com/Financial-Times/go-logger/v2 v2.0.1
	github.com/Financial-Times/http-handlers-go v0.0.0-20180517120644-2c20324ab887 // indirect
	github.com/Financial-Times/http-handlers-go/v2 v2.1.0
	github.com/Financial-Times/neo-model-utils-go v0.0.0-20180712095719-aea1e95c8305
	github.com/Financial-Times/neo-utils-go v0.0.0-20180807105745-1fe6ae2f38f3
	github.com/Financial-Times/service-status-go v0.0.0-20160323111542-3f5199736a3d
	github.com/bradfitz/slice v0.0.0-20180809154707-2b758aa73013 // indirect
	github.com/cyberdelia/go-metrics-graphite v0.0.0-20161219230853-39f87cc3b432 // indirect
	github.com/gorilla/handlers v1.4.2 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/jawher/mow.cli v1.0.4
	github.com/jmcvetta/neoism v1.3.1
	github.com/joho/godotenv v1.2.0
	github.com/mitchellh/hashstructure v1.0.0 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20190706150252-9beb055b7962
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/stretchr/testify v1.3.0
	go4.org v0.0.0-20190313082347-94abd6928b1d // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
)

replace gopkg.in/stretchr/testify.v1 => github.com/stretchr/testify v1.4.0
