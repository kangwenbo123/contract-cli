import assert from "node:assert/strict";
import { cp, mkdtemp, readFile, rm } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { prepareTestE2ESkills } from "./prepare-test-e2e-skills.mjs";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const repositorySkillsDir = path.join(repositoryRoot, "skills");

test("creates Test-only auth guidance in an isolated Skill directory", async () => {
  const root = await mkdtemp(path.join(os.tmpdir(), "contract-cli-test-e2e-skills-"));
  const target = path.join(root, "skills");
  try {
    await cp(repositorySkillsDir, target, { recursive: true });
    const sourceBefore = await readFile(path.join(repositorySkillsDir, "auth", "SKILL.md"), "utf8");

    await prepareTestE2ESkills(target);

    const auth = await readFile(path.join(target, "auth", "SKILL.md"), "utf8");
    const shared = await readFile(path.join(target, "contract-cli-shared", "SKILL.md"), "utf8");
    for (const content of [auth, shared]) {
      assert.match(content, /contract-test/);
      assert.match(content, /https:\/\/test-open\.qtech\.cn/);
      assert.match(content, /https:\/\/test-myaccount\.qtech\.cn/);
      assert.doesNotMatch(content, /正式包固定使用 `contract` profile 和 `prod` 环境/);
    }
    assert.doesNotMatch(auth, /--profile contract(?:\s|$)/);
    assert.equal(await readFile(path.join(repositorySkillsDir, "auth", "SKILL.md"), "utf8"), sourceBefore);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

test("refuses to rewrite the repository Skill source", async () => {
  await assert.rejects(
    prepareTestE2ESkills(repositorySkillsDir),
    /refuses to modify repository Skill sources/,
  );
});
