// CONFIG-AUDIT-1: frontend precedence — there is no runtime env; the bundle
// reads knowledge-graph.config.json baked at build time. It must stay in sync
// with the editable sources in config/*.json (npm run build-config).
import { describe, it, expect } from "vitest";
import { config } from "./config";
import frontendJson from "../../../../config/frontend.json";
import ciCdJson from "../../../../config/ci_cd.json";

describe("baked frontend config", () => {
  it("mirrors config/frontend.json — edit sources and rebuild, there is no env override", () => {
    expect(config.frontend).toEqual(frontendJson.frontend);
  });

  it("mirrors config/ci_cd.json", () => {
    expect(config.ci_cd).toEqual(ciCdJson.ci_cd);
  });
});
