module github.com/taipm/go-deep-agent

go 1.25.2

// Retract v0.7.1 due to invalid file name causing module proxy errors
retract v0.7.1

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/alicebob/miniredis/v2 v2.35.0
	github.com/joho/godotenv v1.5.1
	github.com/openai/openai-go v1.12.0
	github.com/openai/openai-go/v3 v3.8.1
	github.com/redis/go-redis/v9 v9.16.0
	github.com/stretchr/testify v1.11.1
	gonum.org/v1/gonum v0.16.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tidwall/gjson v1.14.4 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	golang.org/x/time v0.14.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
