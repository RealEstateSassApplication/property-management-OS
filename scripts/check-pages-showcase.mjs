import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { Script } from "node:vm";
import test from "node:test";

const html = readFileSync(new URL("../site/index.html", import.meta.url), "utf8");

test("static site is a complete HTML document with accessibility metadata", () => {
  for (const expected of [
    "<!doctype html>", 'lang="en"', 'name="viewport"', "<title>",
    'id="tour"', 'id="workflow"', 'id="architecture"', 'id="roadmap"',
    'role="tablist"', 'role="tabpanel"', 'aria-live="polite"'
  ]) assert.ok(html.toLowerCase().includes(expected.toLowerCase()), expected);
});

test("site shows correct live-vs-sample boundaries", () => {
  assert.match(html, /illustrative/i);
  assert.match(html, /not (live|a live|an authenticated production environment)/i);
  assert.match(html, /owner proofbook/i);
  assert.match(html, /planned/i);
});

test("site remains a static demo without back-end requests", () => {
  assert.doesNotMatch(html, /\bfetch\s*\(|\bXMLHttpRequest\b|\/api\/v1\//);
  assert.doesNotMatch(html, /localStorage|sessionStorage/);
  assert.match(html, /https:\/\/github\.com\/RealEstateSassApplication\/property-management-OS/);
});

test("inline sample-switching script parses", () => {
  const match = html.match(/<script>([\s\S]*?)<\/script>/i);
  assert.ok(match, "Expected inline interactive tour script");
  assert.doesNotThrow(() => new Script(match[1]));
});

test("all demo views are defined", () => {
  for (const view of ["portfolio", "leasing", "rent", "maintenance", "owners"]) {
    assert.match(html, new RegExp('data-view="'+view+'"'));
    assert.match(html, new RegExp('id="tab-'+view+'"'));
  }
});
