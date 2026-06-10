package dev.githome.compat;

import org.junit.jupiter.api.*;
import org.kohsuke.github.*;

import java.io.*;
import java.util.*;

/**
 * hub4j/github-api conformance runner. Gates JAV-1 through JAV-7.
 *
 * Each test method maps to one gate. The TAP output is printed to stdout
 * because maven surefire captures it; add -Dsurefire.useFile=false to
 * see it in the console.
 */
@TestMethodOrder(MethodOrderer.OrderAnnotation.class)
class HubApiTest {

    static GitHub gh;
    static String owner;
    static String repo;
    static int issueNumber;
    static int prNumber;

    @BeforeAll
    static void setUp() throws Exception {
        String host    = requireEnv("GITHOME_HOST");
        String token   = requireEnv("GITHOME_TOKEN");
        owner          = env("GITHOME_OWNER", "test-owner");
        repo           = env("GITHOME_REPO",  "test-repo");
        issueNumber    = Integer.parseInt(env("GITHOME_ISSUE_NUMBER", "1"));
        prNumber       = Integer.parseInt(env("GITHOME_PR_NUMBER", "1"));

        boolean insecure = "1".equals(System.getenv("GITHOME_INSECURE"));
        String scheme = insecure ? "http" : "https";
        String endpoint = scheme + "://" + host + "/api/v3";

        gh = new GitHubBuilder()
                .withEndpoint(endpoint)
                .withOAuthToken(token)
                .build();
    }

    /** JAV-1: checkApiUrlValidity */
    @Test @Order(1)
    void jav1CheckApiValidity() throws Exception {
        Assertions.assertDoesNotThrow(() -> gh.checkApiUrlValidity());
    }

    /** JAV-2: getMyself */
    @Test @Order(2)
    void jav2GetMyself() throws Exception {
        GHMyself me = gh.getMyself();
        Assertions.assertNotNull(me);
        Assertions.assertFalse(me.getLogin().isBlank(), "login is empty");
    }

    /** JAV-3: getRepository */
    @Test @Order(3)
    void jav3GetRepository() throws Exception {
        GHRepository r = gh.getRepository(owner + "/" + repo);
        Assertions.assertNotNull(r);
        Assertions.assertFalse(r.getName().isBlank(), "name is empty");
        Assertions.assertFalse(r.getDefaultBranch().isBlank(), "default_branch is empty");
    }

    /** JAV-4: issues pagination */
    @Test @Order(4)
    void jav4IssuesPagination() throws Exception {
        GHRepository r = gh.getRepository(owner + "/" + repo);
        PagedIterable<GHIssue> issues = r.getIssues(GHIssueState.ALL);
        Iterator<GHIssue> it = issues.iterator();
        // Just check iteration doesn't throw
        int count = 0;
        while (it.hasNext() && count < 5) {
            GHIssue issue = it.next();
            Assertions.assertTrue(issue.getNumber() > 0);
            count++;
        }
    }

    /** JAV-5: createHook */
    @Test @Order(5)
    void jav5CreateHook() throws Exception {
        GHRepository r = gh.getRepository(owner + "/" + repo);
        Map<String, String> config = new HashMap<>();
        config.put("url", "https://example.com/compat-jav5-" + System.currentTimeMillis());
        config.put("content_type", "json");
        GHHook hook = r.createHook("web", config, List.of("push"), false);
        Assertions.assertTrue(hook.getId() > 0);
        hook.delete();
    }

    /** JAV-6: createCommitStatus */
    @Test @Order(6)
    void jav6CreateCommitStatus() throws Exception {
        GHRepository r = gh.getRepository(owner + "/" + repo);
        // Use the head SHA from env or the latest commit
        String sha = System.getenv("GITHOME_HEAD_SHA");
        if (sha == null || sha.isBlank()) {
            var branches = r.getBranches();
            String defaultBranch = r.getDefaultBranch();
            sha = branches.get(defaultBranch).getSHA1();
        }
        GHCommitStatus status = r.createCommitStatus(sha,
                GHCommitState.SUCCESS, "https://example.com", "compat test", "compat/jav6");
        Assertions.assertNotNull(status);
    }

    /** JAV-7: createCheckRun — skip (requires GitHub App) */
    @Test @Order(7)
    void jav7CreateCheckRun() {
        Assumptions.assumeTrue(false, "JAV-7: skipped — requires GitHub App auth");
    }

    // helpers

    static String requireEnv(String key) {
        String v = System.getenv(key);
        if (v == null || v.isBlank()) throw new IllegalStateException("need env: " + key);
        return v;
    }

    static String env(String key, String def) {
        String v = System.getenv(key);
        return (v != null && !v.isBlank()) ? v : def;
    }
}
