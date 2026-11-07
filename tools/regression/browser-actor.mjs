import assert from "node:assert/strict";
import { request } from "playwright";

export async function referenceActor(input) {
  const origin = new URL(input.base);
  assert.equal(origin.hostname, "127.0.0.1");
  assert.equal(origin.protocol, "http:");
  const actor = await request.newContext({ baseURL: origin.origin, extraHTTPHeaders: { Origin: origin.origin } });
  try {
    const login = await actor.post("/api/v1/auth/login", { data: input.credentials });
    assert.equal(login.status(), 200);
    const csrfToken = (await login.json()).csrfToken;
    assert.ok(csrfToken);
    return { actor, csrfToken, origin };
  } catch (error) {
    await actor.dispose();
    throw error;
  }
}
