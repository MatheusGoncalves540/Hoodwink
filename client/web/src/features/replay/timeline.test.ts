import { readFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { ReplayInvalidError, type ReplayFileLike } from "../../replay/loader";
import { decodeDeltas, readReplayDeltas } from "./timeline";

const here = dirname(fileURLToPath(import.meta.url));
const goldenPath = join(resolve(here, "../../../../.."), "server/replay/engines/v1/testdata/golden_1.hwr");
const golden = new Uint8Array(readFileSync(goldenPath));

const fileOf = (bytes: Uint8Array): ReplayFileLike => ({
  size: bytes.byteLength,
  arrayBuffer: async () =>
    bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength) as ArrayBuffer,
});

describe("deltas do .hwr", () => {
  it("lê os 105 deltas do golden_1 sem tocar no motor", async () => {
    const deltas = await readReplayDeltas(fileOf(golden));
    expect(deltas).toHaveLength(105);
    expect(deltas[0]).toBe(0);
    expect(deltas.reduce((sum, delta) => sum + delta, 0)).toBe(80);
  });

  it("corpo vazio ou truncado não vira lista silenciosa", async () => {
    expect(decodeDeltas(new Uint8Array())).toEqual([]);
    expect(() => decodeDeltas(golden.subarray(0, 10))).toThrow(ReplayInvalidError);
    await expect(readReplayDeltas(fileOf(new Uint8Array([1, 2, 3])))).rejects.toBeInstanceOf(
      ReplayInvalidError,
    );
  });
});
