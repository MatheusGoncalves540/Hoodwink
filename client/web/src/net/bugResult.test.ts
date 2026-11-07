import { describe, expect, it } from "vitest";
import { classifyBugResult } from "./bugResult";

describe("classifyBugResult", () => {
  it("200 e 201 são ok", () => {
    expect(classifyBugResult(200)).toBe("ok");
    expect(classifyBugResult(201)).toBe("ok");
  });

  it("429 é limite", () => {
    expect(classifyBugResult(429)).toBe("limit");
  });

  it("401 pede login", () => {
    expect(classifyBugResult(401)).toBe("unauthorized");
  });

  it("o resto vira erro, inclusive rede e 5xx", () => {
    expect(classifyBugResult(0)).toBe("error");
    expect(classifyBugResult(400)).toBe("error");
    expect(classifyBugResult(403)).toBe("error");
    expect(classifyBugResult(404)).toBe("error");
    expect(classifyBugResult(500)).toBe("error");
  });
});
