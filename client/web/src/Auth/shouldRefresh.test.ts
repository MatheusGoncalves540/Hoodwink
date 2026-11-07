import { describe, expect, it } from "vitest";
import shouldRefresh from "./shouldRefresh";

const token = (payload: object) => `h.${btoa(JSON.stringify(payload))}.s`;
const HOUR = 3600;

describe("shouldRefresh", () => {
  it("antes da metade não renova", () => {
    const now = 1_000_000;
    expect(shouldRefresh(token({ iat: now - 1000, exp: now + 1000 }), now)).toBe(false);
  });

  it("depois da metade renova", () => {
    const now = 1_000_000;
    expect(shouldRefresh(token({ iat: now - 2000, exp: now + 1000 }), now)).toBe(true);
  });

  it("sem iat e faltando 11h renova", () => {
    const now = 1_000_000;
    expect(shouldRefresh(token({ exp: now + 11 * HOUR }), now)).toBe(true);
  });

  it("sem iat e faltando 13h não renova", () => {
    const now = 1_000_000;
    expect(shouldRefresh(token({ exp: now + 13 * HOUR }), now)).toBe(false);
  });

  it("token inválido não renova", () => {
    expect(shouldRefresh("nada", 1_000_000)).toBe(false);
  });
});
