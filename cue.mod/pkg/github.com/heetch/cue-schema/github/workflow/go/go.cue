// Go-centric defaults for github actions.
//
// This defines a single test job with appropriate defaults.
// Override the Platforms, Versions, RunTest and Services definitions
// for easy modification of some parameters.
package workflow

on:   _ | *["push", "pull_request"]
name: _ | *"Test"
jobs: test: {
	strategy: matrix: {
		"go-version": _ | *[ "\(v).x" for v in Versions ]
		platform:     _ | *[ "\(p)-latest" for p in Platforms ]
	}
	"runs-on": "${{ matrix.platform }}"
	steps: [{
		name: "Install Go"
		uses: "actions/setup-go@v1"
		with: "go-version": "${{ matrix.go-version }}"
	}, {
		name: "Checkout code"
		uses: "actions/checkout@v1"
	}, _ | *{
		name: "Test"
		run:  RunTest
	}]
}

// Include all named services.
for name in Services {
	jobs: test: services: "\(name)": ServiceConfig[name]
}

jobs: test: services: kafka?: _ | *{
	image: "confluentinc/cp-kafka:latest"
	env: {
		KAFKA_BROKER_ID:                        "1"
		KAFKA_ZOOKEEPER_CONNECT:                "zookeeper:2181"
		KAFKA_ADVERTISED_LISTENERS:             "PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092"
		KAFKA_LISTENER_SECURITY_PROTOCOL_MAP:   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT"
		KAFKA_INTER_BROKER_LISTENER_NAME:       "PLAINTEXT"
		KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: "1"
	}
}

// Platforms configures what platforms to run the tests on.
Platforms :: *["ubuntu"] | [ ... "ubuntu" | "macos" | "windows"]

// Versions configures what Go versions to run the tests on.
// TODO regexp.Match("^1.[0-9]+$")
Versions :: *["1.13"] | [ ... string]

// RunTest configures the command used to run the tests.
RunTest :: *"go test ./..." | string

// Service configures which services to make available.
// The configuration the service with name N is taken from
// ServiceConfig[N]
Services :: [... string]

// ServiceConfig holds the default configuration for services that
// can be started by naming them in Services.
ServiceConfig :: [_]: _

ServiceConfig :: kafka: {
	image: "confluentinc/cp-kafka:latest"
	env: {
		KAFKA_BROKER_ID:                        "1"
		KAFKA_ZOOKEEPER_CONNECT:                "zookeeper:2181"
		KAFKA_ADVERTISED_LISTENERS:             "PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092"
		KAFKA_LISTENER_SECURITY_PROTOCOL_MAP:   "PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT"
		KAFKA_INTER_BROKER_LISTENER_NAME:       "PLAINTEXT"
		KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: "1"
	}
}

ServiceConfig :: postgres: {
	env: {
		POSTGRES_DB:       "postgres"
		POSTGRES_PASSWORD: "postgres"
		POSTGRES_USER:     "postgres"
	}
	image:   "postgres:10.8"
	options: "--health-cmd pg_isready --health-interval 10s --health-timeout 5s --health-retries 5"
	ports: [
		"5432:5432",
	]
}
