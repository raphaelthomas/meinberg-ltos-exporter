USER         := admin
PASS         := password
NEXT_VERSION := $(shell svu next)

.PHONY: build run test-certs mock-api release test clean
build:
	go build -o meinberg_ltos_exporter .

run: build
	AUTH_PASS=$(PASS) ./meinberg_ltos_exporter --ltos-api-url https://localhost:8080 --log-level=debug --ignore-ssl-verify --auth-user $(USER)

test-certs:
	 openssl req -x509 -newkey rsa:4096 -keyout tests/mock-key.pem -out tests/mock-cert.pem -sha256 -days 5 -nodes -subj "/C=CH/ST=State/L=City/O=Organization/OU=Department/CN=localhost"

mock-api: test-certs
	go run tests/mock-server.go -ssl-cert tests/mock-cert.pem -ssl-key tests/mock-key.pem -file tests/mock-api-status-response.json -user $(USER) -pass $(PASS)

release:
	@echo "Current version: $(shell svu current)"
	@echo "Next version:    $(NEXT_VERSION)"
	@read -p "Press enter to confirm release or Ctrl+C to cancel"
	git tag -a $(NEXT_VERSION) -m "Release $(NEXT_VERSION)"
	git push origin $(NEXT_VERSION)

test:
	go test -v ./...

clean:
	rm -rv dist/
	rm -v meinberg_ltos_exporter
