build:
	go build -o meinberg_ltos_exporter .

run: build
	./meinberg_ltos_exporter --ltos-api-url http://localhost:8080 --log-level=debug
