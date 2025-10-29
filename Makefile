.PHONY: build clean test start_colab start_gradio

BINARIES := brief extract gradio colab

build:
	@for bin in $(BINARIES); do \
		echo "Building $$bin"; \
		go build -o $$bin ./cmd/$$bin/ || exit 1; \
	done

clean:
	@rm -f $(BINARIES)

test:
	@go test ./... -cover

# Usage: make start_colab user=test1
start_colab:
	pkill -9 colab || true
	sleep 2
	rm -f chrome-user-data/$(user)/SingletonLock chrome-user-data/$(user)/SingletonSocket
	./colab $(user) 2>&1 | tee debug/output_colab.log

# Usage: make start_gradio gradio=https://xxxxxx.gradio.live
start_gradio:
	pkill -9 gradio || true
	sleep 2
	./gradio -input cmd/gradio/example.json -gradio_url $(gradio) 2>&1 | tee debug/output_gradio.log
