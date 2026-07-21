import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { createLogger, logger } from "./logger";

describe("logger", () => {
  let debugSpy: ReturnType<typeof vi.spyOn>;
  let infoSpy: ReturnType<typeof vi.spyOn>;
  let warnSpy: ReturnType<typeof vi.spyOn>;
  let errorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    debugSpy = vi.spyOn(console, "debug").mockImplementation(() => {});
    infoSpy = vi.spyOn(console, "info").mockImplementation(() => {});
    warnSpy = vi.spyOn(console, "warn").mockImplementation(() => {});
    errorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("creates loggers with different contexts", () => {
    const testLogger = createLogger("test");
    testLogger.info("hello");
    expect(infoSpy).toHaveBeenCalledWith(
      expect.stringContaining("[test]"),
      "hello",
      "",
    );
  });

  it("logs at all levels", () => {
    const log = createLogger("levels");
    log.debug("debug msg", { a: 1 });
    log.info("info msg");
    log.warn("warn msg");
    log.error("error msg", new Error("boom"));

    expect(debugSpy).toHaveBeenCalledOnce();
    expect(infoSpy).toHaveBeenCalledOnce();
    expect(warnSpy).toHaveBeenCalledOnce();
    expect(errorSpy).toHaveBeenCalledOnce();
  });

  it("buffers logs and exposes them via getLogs", () => {
    const log = createLogger("buffer");
    log.info("one");
    log.warn("two");

    const logs = log.getLogs();
    expect(logs).toHaveLength(2);
    expect(logs[1].message).toBe("two");
    expect(logs[1].level).toBe("warn");
  });

  it("clears the log buffer", () => {
    const log = createLogger("clear");
    log.info("msg");
    expect(log.getLogs()).toHaveLength(1);

    log.clearLogs();
    expect(log.getLogs()).toHaveLength(0);
  });

  it("caps the buffer size at 100 entries", () => {
    const log = createLogger("capacity");
    for (let i = 0; i < 105; i++) {
      log.info(`msg ${i}`);
    }
    expect(log.getLogs()).toHaveLength(100);
    expect(log.getLogs()[0].message).toBe("msg 5");
  });

  it("exposes a global logger", () => {
    logger.info("global");
    expect(infoSpy).toHaveBeenCalledWith(
      expect.stringContaining("[app]"),
      "global",
      "",
    );
  });
});
