module github.com/episub/spawn/v2

go 1.12

require (
	cloud.google.com/go v0.37.2
	github.com/99designs/gqlgen v0.11.3
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/caarlos0/env/v6 v6.5.0
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/codemodus/kace v0.5.1
	github.com/episub/pqt v0.0.0-20181112131323-01e28ce941a0
	github.com/fatih/color v1.10.0 // indirect
	github.com/getsentry/sentry-go v0.2.1
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/hokaccha/go-prettyjson v0.0.0-20180920040306-f579f869bbfe
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.5.0+incompatible
	github.com/lib/pq v1.2.0
	github.com/mitchellh/mapstructure v1.3.2
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/oleiade/reflections v1.0.1 // indirect
	github.com/open-policy-agent/opa v0.16.0
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190324214902-3020fec0e66b
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/radovskyb/watcher v1.0.6
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sendgrid/rest v2.6.3+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.6.3+incompatible
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/uber-go/atomic v0.0.0-00010101000000-000000000000 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/urfave/cli v1.20.0
	github.com/vektah/gqlparser v1.3.1
	github.com/vektah/gqlparser/v2 v2.0.1
	gnorm.org/gnorm v1.0.0
	go.uber.org/atomic v1.7.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.3.0
)

replace gnorm.org/gnorm => github.com/episub/gnorm v1.1.2

replace github.com/uber-go/atomic => github.com/uber-go/atomic v1.4.0
