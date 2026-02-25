build:
	go build -o meinberg_ltos_exporter .

run: build
	./meinberg_ltos_exporter --ltos-api-url https://localhost:8080 --log-level=debug --ignore-ssl-verify

test-certs:
	 openssl req -x509 -newkey rsa:4096 -keyout tests/mock-key.pem -out tests/mock-cert.pem -sha256 -days 5 -nodes -subj "/C=CH/ST=State/L=City/O=Organization/OU=Department/CN=localhost"

mock-api: test-certs
	go run tests/mock-server.go -ssl-cert tests/mock-cert.pem -ssl-key tests/mock-key.pem -file tests/mock-api-status-response.json
