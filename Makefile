.PHONY: build vet test list clean run-gh run-git run-js run-python run-go run-ruby run-dotnet run-java run-terraform run-all

CONFORM := go run ./cmd/conform

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

list:
	$(CONFORM) --list

# ── per-runner targets (need GITHOME_HOST + GITHOME_TOKEN) ────────────────────

run-git:
	bash runners/shell/git_test.sh

run-gh:
	bash runners/shell/gh_cli_test.sh

run-js:
	cd runners/js && npm ci && cd ../.. && node runners/js/octokit_test.mjs

run-python:
	cd runners/python && pip install -q -r requirements.txt && cd ../.. && python3 runners/python/pygithub_test.py

run-go:
	cd runners/go && go run . && cd ../..

run-ruby:
	cd runners/ruby && bundle install -q && cd ../.. && ruby runners/ruby/octokit_rb_test.rb

run-dotnet:
	dotnet run --project runners/dotnet/CompatTest.csproj

run-java:
	mvn -pl runners/java -q test -Dsurefire.useFile=false

run-terraform:
	bash runners/shell/terraform_test.sh

# Run all runners via the conform binary (respects -parallel)
run-all:
	$(CONFORM) --report report.json --tap

# ── setup targets ─────────────────────────────────────────────────────────────

install-js-deps:
	cd runners/js && npm ci

install-python-deps:
	pip install -r runners/python/requirements.txt

install-ruby-deps:
	cd runners/ruby && bundle install

go-runner-deps:
	cd runners/go && go mod tidy

# ── cassettes ─────────────────────────────────────────────────────────────────

scrub-cassettes:
	find cassettes -name '*.yaml' -exec \
	  sed -i 's/Bearer [A-Za-z0-9_-]*/Bearer SCRUBBED/g; s/token [A-Za-z0-9_-]*/token SCRUBBED/g' {} +

clean:
	rm -f conform report.json *.tap
	find . -name '*.tap' -delete
