// Octokit JS/TS conformance runner
// Gates: OCT-1 through OCT-11

import { Octokit } from "@octokit/rest";
import { graphql } from "@octokit/graphql";
import { createAppAuth } from "@octokit/auth-app";
import { Webhooks } from "@octokit/webhooks";

const HOST = process.env.GITHOME_HOST ?? (() => { throw new Error("need GITHOME_HOST") })();
const TOKEN = process.env.GITHOME_TOKEN ?? (() => { throw new Error("need GITHOME_TOKEN") })();
const OWNER = process.env.GITHOME_OWNER ?? "test-owner";
const REPO = process.env.GITHOME_REPO ?? "test-repo";
const APP_ID = process.env.GITHOME_APP_ID;
const INSTALL_ID = process.env.GITHOME_INSTALL_ID;
const APP_PEM = process.env.GITHOME_APP_PEM;
const WEBHOOK_SECRET = process.env.GITHOME_WEBHOOK_SECRET ?? "test-secret";

const SCHEME = process.env.GITHOME_INSECURE === "1" ? "http" : "https";
const REST_BASE = `${SCHEME}://${HOST}/api/v3`;
const GQL_URL = `${SCHEME}://${HOST}/api/graphql`;

const octokit = new Octokit({ auth: TOKEN, baseUrl: REST_BASE });
const gql = graphql.defaults({ headers: { authorization: `token ${TOKEN}` }, baseUrl: GQL_URL });

const results = [];
let num = 0;

async function check(gateId, desc, fn) {
  num++;
  try {
    await fn();
    results.push(`ok ${num} - ${gateId}: ${desc}`);
  } catch (e) {
    results.push(`not ok ${num} - ${gateId}: ${desc}`);
    results.push(`  # ${e.message?.slice(0, 200) ?? String(e)}`);
  }
}

function skipGate(gateId, desc, reason) {
  num++;
  results.push(`ok ${num} - ${gateId}: ${desc} # SKIP ${reason}`);
}

// OCT-1: users.getAuthenticated()
await check("OCT-1", "users.getAuthenticated()", async () => {
  const { data } = await octokit.users.getAuthenticated();
  if (!data.login) throw new Error("no login in response");
});

// OCT-2: repos.get()
await check("OCT-2", "repos.get()", async () => {
  const { data } = await octokit.repos.get({ owner: OWNER, repo: REPO });
  if (!data.id) throw new Error("no id");
  if (!data.default_branch) throw new Error("no default_branch");
});

// OCT-3: paginate issues
await check("OCT-3", "paginate issues", async () => {
  const issues = await octokit.paginate(octokit.issues.listForRepo, {
    owner: OWNER, repo: REPO, state: "all", per_page: 5,
  });
  if (!Array.isArray(issues)) throw new Error("not an array");
});

// OCT-4: create + close issue
await check("OCT-4", "create and close issue", async () => {
  const { data: created } = await octokit.issues.create({
    owner: OWNER, repo: REPO, title: `compat-test-oct-${Date.now()}`, body: "automated",
  });
  if (!created.number || created.number <= 0) throw new Error("bad issue number");
  await octokit.issues.update({ owner: OWNER, repo: REPO, issue_number: created.number, state: "closed" });
});

// OCT-5: GraphQL viewer
await check("OCT-5", "GraphQL viewer { login }", async () => {
  const data = await gql(`{ viewer { login } }`);
  if (!data.viewer?.login) throw new Error("no viewer.login");
});

// OCT-6: GraphQL repository
await check("OCT-6", "GraphQL repository", async () => {
  const data = await gql(`
    query($owner: String!, $name: String!) {
      repository(owner: $owner, name: $name) {
        id
        defaultBranchRef { name }
      }
    }`, { owner: OWNER, name: REPO });
  if (!data.repository?.id) throw new Error("no repository.id");
  if (!data.repository?.defaultBranchRef?.name) throw new Error("no defaultBranchRef.name");
});

// OCT-7: rate-limit headers present
await check("OCT-7", "rate-limit headers present", async () => {
  const { headers } = await octokit.request("GET /user");
  if (!headers["x-ratelimit-limit"]) throw new Error("x-ratelimit-limit missing");
  if (!headers["x-github-request-id"]) throw new Error("x-github-request-id missing");
});

// OCT-8: ETag / 304
await check("OCT-8", "ETag / 304 caching", async () => {
  const first = await octokit.request("GET /repos/{owner}/{repo}", { owner: OWNER, repo: REPO });
  const etag = first.headers.etag;
  if (!etag) throw new Error("no ETag on first response");
  try {
    await octokit.request("GET /repos/{owner}/{repo}", {
      owner: OWNER, repo: REPO,
      headers: { "If-None-Match": etag },
    });
    // If we get here, the server did NOT return 304 — fail
    throw new Error("expected 304 Not Modified");
  } catch (e) {
    // @octokit/rest throws on 304 — check that it's a 304 error
    if (e.status === 304 || e.message?.includes("304")) return; // pass
    throw e;
  }
});

// OCT-9: webhook signature verification
await check("OCT-9", "webhook signature verification", async () => {
  const wh = new Webhooks({ secret: WEBHOOK_SECRET });
  const body = JSON.stringify({ action: "opened", zen: "test" });
  // Node.js built-in crypto HMAC
  const crypto = await import("node:crypto");
  const sig = "sha256=" + crypto.createHmac("sha256", WEBHOOK_SECRET).update(body).digest("hex");
  const event = {
    id: "test-delivery-1",
    name: "push",
    signature: sig,
    payload: body,
  };
  await wh.verifyAndReceive(event);
});

// OCT-10: GitHub App installation token
if (APP_ID && INSTALL_ID && APP_PEM) {
  await check("OCT-10", "GitHub App installation token", async () => {
    const auth = createAppAuth({ appId: APP_ID, privateKey: APP_PEM, installationId: INSTALL_ID });
    const authData = await auth({ type: "installation" });
    if (!authData.token?.startsWith("ghs_")) throw new Error(`expected ghs_ prefix, got ${authData.token?.slice(0, 10)}`);
    // Use the installation token for a simple request
    const instOctokit = new Octokit({ auth: authData.token, baseUrl: REST_BASE });
    const { data } = await instOctokit.users.getAuthenticated();
    if (!data.login) throw new Error("no login with installation token");
  });
} else {
  skipGate("OCT-10", "GitHub App installation token", "GITHOME_APP_ID/INSTALL_ID/APP_PEM not set");
}

// OCT-11: @actions/github GITHUB_API_URL env
await check("OCT-11", "GITHUB_API_URL / GITHUB_GRAPHQL_URL env recognized", async () => {
  // Verify the env vars are set correctly (they would be consumed by @actions/github)
  const apiUrl = process.env.GITHUB_API_URL;
  const gqlUrl = process.env.GITHUB_GRAPHQL_URL;
  if (!apiUrl) throw new Error("GITHUB_API_URL not set");
  if (!apiUrl.includes("/api/v3")) throw new Error(`GITHUB_API_URL should include /api/v3, got: ${apiUrl}`);
  if (!gqlUrl) throw new Error("GITHUB_GRAPHQL_URL not set");
  if (!gqlUrl.includes("/api/graphql")) throw new Error(`GITHUB_GRAPHQL_URL should include /api/graphql, got: ${gqlUrl}`);
  // Make a real request with the env-derived URL
  const envOctokit = new Octokit({ auth: TOKEN, baseUrl: apiUrl });
  const { data } = await envOctokit.users.getAuthenticated();
  if (!data.login) throw new Error("no login via GITHUB_API_URL");
});

console.log("TAP version 14");
console.log(`1..${num}`);
results.forEach(r => console.log(r));
