# Cassettes

This directory holds go-vcr v2 YAML cassettes recording exact GitHub API
request/response pairs. They serve as the oracle for contract testing —
Githome's output is diffed against them.

## Format

```yaml
version: 2
interactions:
- request:
    body: ""
    headers:
      Accept: ["application/vnd.github+json"]
      Authorization: ["Bearer SCRUBBED"]
    method: GET
    url: https://api.github.com/repos/owner/repo
  response:
    body: '{"id": 1, "name": "repo", ...}'
    headers:
      Content-Type: ["application/json; charset=utf-8"]
      ETag: ['W/"abc123"']
    status: 200 OK
    code: 200
```

## Recording

```sh
CASSETTE_RECORD=1 GITHUB_TOKEN=your_real_token \
  go test ./cassettes/...
```

Tokens are scrubbed before committing. Run `make scrub-cassettes` to
replace all auth values with `SCRUBBED`.

## Directory layout

```
cassettes/
├── rest/
│   ├── GET_user.yaml
│   ├── GET_repos_owner_repo.yaml
│   ├── POST_repos_owner_repo_issues.yaml
│   └── ...
└── graphql/
    ├── viewer.yaml
    ├── repository.yaml
    └── ...
```
