import { strict as assert } from "node:assert";
import { test } from "node:test";
import type { AddressInfo } from "node:net";
import { quote } from "./pricing.ts";
import { buildServer } from "./server.ts";

test("quote sums lines without discount below the bulk threshold", () => {
  assert.deepEqual(quote([{ sku: "a", unitPriceCents: 250, quantity: 2 }]), {
    subtotalCents: 500,
    discountCents: 0,
    totalCents: 500,
  });
});

test("quote applies 10% bulk discount per qualifying line", () => {
  const q = quote([
    { sku: "a", unitPriceCents: 100, quantity: 10 },
    { sku: "b", unitPriceCents: 999, quantity: 1 },
  ]);
  assert.equal(q.subtotalCents, 1999);
  assert.equal(q.discountCents, 100);
  assert.equal(q.totalCents, 1899);
});

test("quote rejects invalid input", () => {
  assert.throws(() => quote([{ sku: "x", unitPriceCents: -1, quantity: 1 }]), RangeError);
  assert.throws(() => quote([{ sku: "x", unitPriceCents: 1, quantity: 0 }]), RangeError);
  assert.throws(() => quote([{ sku: "x", unitPriceCents: 1.5, quantity: 1 }]), RangeError);
});

test("HTTP API: healthz, quote, errors", async () => {
  const server = buildServer();
  await new Promise<void>((resolve) => server.listen(0, "127.0.0.1", resolve));
  const { port } = server.address() as AddressInfo;
  const base = `http://127.0.0.1:${port}`;
  try {
    const health = await fetch(`${base}/healthz`);
    assert.equal(health.status, 200);
    assert.deepEqual(await health.json(), { status: "ok" });

    const ok = await fetch(`${base}/quote`, {
      method: "POST",
      body: JSON.stringify({ items: [{ sku: "a", unitPriceCents: 100, quantity: 10 }] }),
    });
    assert.equal(ok.status, 200);
    assert.deepEqual(await ok.json(), { subtotalCents: 1000, discountCents: 100, totalCents: 900 });

    const bad = await fetch(`${base}/quote`, { method: "POST", body: "{}" });
    assert.equal(bad.status, 400);

    const invalid = await fetch(`${base}/quote`, { method: "POST", body: "not json" });
    assert.equal(invalid.status, 400);

    const missing = await fetch(`${base}/nope`);
    assert.equal(missing.status, 404);
  } finally {
    await new Promise<void>((resolve) => server.close(() => resolve()));
  }
});
