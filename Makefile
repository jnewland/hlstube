.PHONY: all
all:
	docker build -t dev .
	docker rm -f dev || true
	docker run --name dev --rm -i -p 8080:8080 --env-file .env dev