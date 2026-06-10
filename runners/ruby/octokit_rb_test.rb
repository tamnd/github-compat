#!/usr/bin/env ruby
# Octokit.rb conformance runner. Gates RB-1 through RB-8.

require "bundler/setup"
require "octokit"

HOST          = ENV.fetch("GITHOME_HOST")         { abort "need GITHOME_HOST" }
TOKEN         = ENV.fetch("GITHOME_TOKEN")        { abort "need GITHOME_TOKEN" }
OWNER         = ENV.fetch("GITHOME_OWNER",  "test-owner")
REPO          = ENV.fetch("GITHOME_REPO",   "test-repo")
ISSUE         = ENV.fetch("GITHOME_ISSUE_NUMBER", "1").to_i
INSECURE      = ENV["GITHOME_INSECURE"] == "1"
SCHEME        = INSECURE ? "http" : "https"
API_ENDPOINT  = "#{SCHEME}://#{HOST}/api/v3/"

TOTAL = 8
results = []
n = 0

def tap_pass(results, n, gate, desc)  = results << "ok #{n} - #{gate}: #{desc}"
def tap_fail(results, n, gate, desc, err = nil)
  results << "not ok #{n} - #{gate}: #{desc}"
  results << "  # #{err.to_s.slice(0, 200)}" if err
end
def tap_skip(results, n, gate, desc, reason) = results << "ok #{n} - #{gate}: #{desc} # SKIP #{reason}"

client = Octokit::Client.new(
  access_token:  TOKEN,
  api_endpoint:  API_ENDPOINT,
  auto_paginate: false,
  connection_options: {
    ssl: { verify: !INSECURE }
  }
)

# RB-1: client.user
n += 1
begin
  u = client.user
  raise "no login" unless u[:login]
  tap_pass(results, n, "RB-1", "client.user")
rescue => e
  tap_fail(results, n, "RB-1", "client.user", e)
end

# RB-2: client.repo
n += 1
begin
  r = client.repo("#{OWNER}/#{REPO}")
  raise "no name"           unless r[:name]
  raise "no default_branch" unless r[:default_branch]
  tap_pass(results, n, "RB-2", "client.repo")
rescue => e
  tap_fail(results, n, "RB-2", "client.repo", e)
end

# RB-3: client.issues
n += 1
begin
  issues = client.issues("#{OWNER}/#{REPO}", state: "all", per_page: 5)
  raise "not an array" unless issues.is_a?(Array)
  tap_pass(results, n, "RB-3", "client.issues")
rescue => e
  tap_fail(results, n, "RB-3", "client.issues", e)
end

# RB-4: create + close issue
n += 1
begin
  title = "compat-rb4-#{Time.now.to_i}"
  iss = client.create_issue("#{OWNER}/#{REPO}", title, "auto")
  raise "bad number" unless iss[:number].to_i > 0
  client.close_issue("#{OWNER}/#{REPO}", iss[:number])
  tap_pass(results, n, "RB-4", "create and close issue")
rescue => e
  tap_fail(results, n, "RB-4", "create and close issue", e)
end

# RB-5: client.rate_limit
n += 1
begin
  rl = client.rate_limit
  raise "limit is zero" unless rl[:limit].to_i > 0
  tap_pass(results, n, "RB-5", "client.rate_limit")
rescue => e
  tap_fail(results, n, "RB-5", "client.rate_limit", e)
end

# RB-6: client.pull_requests
n += 1
begin
  prs = client.pull_requests("#{OWNER}/#{REPO}", state: "all", per_page: 5)
  raise "not an array" unless prs.is_a?(Array)
  tap_pass(results, n, "RB-6", "client.pull_requests")
rescue => e
  tap_fail(results, n, "RB-6", "client.pull_requests", e)
end

# RB-7: client.create_hook (webhook)
n += 1
begin
  wh = client.create_hook("#{OWNER}/#{REPO}", "web", {
    url:          "https://example.com/compat-rb7-#{Time.now.to_i}",
    content_type: "json",
    secret:       "test-secret",
    insecure_ssl: "0"
  }, events: ["push"], active: false)
  raise "no id" unless wh[:id].to_i > 0
  client.remove_hook("#{OWNER}/#{REPO}", wh[:id])
  tap_pass(results, n, "RB-7", "client.create_hook")
rescue => e
  tap_fail(results, n, "RB-7", "client.create_hook", e)
end

# RB-8: auto_paginate
n += 1
begin
  auto_client = Octokit::Client.new(
    access_token:  TOKEN,
    api_endpoint:  API_ENDPOINT,
    auto_paginate: true
  )
  issues = auto_client.issues("#{OWNER}/#{REPO}", state: "all", per_page: 2)
  raise "not an array" unless issues.is_a?(Array)
  tap_pass(results, n, "RB-8", "auto_paginate")
rescue => e
  tap_fail(results, n, "RB-8", "auto_paginate", e)
end

puts "TAP version 14"
puts "1..#{TOTAL}"
results.each { |r| puts r }
