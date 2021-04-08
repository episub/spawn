module github.com/episub/spawn/v2

go 1.12

require (
	cloud.google.com/go v0.37.2
	github.com/99designs/gqlgen v0.11.3
	github.com/OneOfOne/xxhash v1.2.5 // indirect
	github.com/agnivade/levenshtein v1.1.0 // indirect
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/cockroachdb/apd v1.1.0 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/codemodus/kace v0.5.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/episub/pqt v0.0.0-20181112131323-01e28ce941a0
	github.com/fatih/color v1.7.0 // indirect
	github.com/getsentry/sentry-go v0.2.1
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/h2non/filetype v1.0.10
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20180920040306-f579f869bbfe
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.5.0+incompatible
	github.com/lib/pq v1.0.0
	github.com/matryer/moq v0.0.0-20200607124540-4638a53893e6 // indirect
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mitchellh/mapstructure v1.3.2
	github.com/nbutton23/zxcvbn-go v0.0.0-20180912185939-ae427f1e4c1d
	github.com/open-policy-agent/opa v0.16.0
	github.com/opentracing-contrib/go-stdlib v0.0.0-20190324214902-3020fec0e66b
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/browser v0.0.0-20180916011732-0a3d74bf9ce4 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/radovskyb/watcher v1.0.6
	github.com/rakyll/statik v0.1.6 // indirect
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/sendgrid/rest v2.6.1+incompatible // indirect
	github.com/sendgrid/sendgrid-go v3.6.3+incompatible
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24 // indirect
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3 // indirect
	github.com/spf13/pflag v1.0.3 // indirect
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	github.com/urfave/cli v1.20.0
	github.com/urfave/cli/v2 v2.2.0 // indirect
	github.com/vektah/dataloaden v0.3.0 // indirect
	github.com/vektah/gqlparser v1.1.2
	github.com/vektah/gqlparser/v2 v2.0.1
	gnorm.org/gnorm v1.0.0
	go.uber.org/atomic v1.4.0 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/tools v0.0.0-20200717024301-6ddee64345a6 // indirect
	gopkg.in/oleiade/reflections.v1 v1.0.0
	gopkg.in/yaml.v2 v2.3.0
)

replace gnorm.org/gnorm => github.com/episub/gnorm v1.1.2