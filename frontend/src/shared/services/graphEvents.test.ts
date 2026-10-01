import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { subscribeGraphEvents, type EventSourceLike } from "./graphEvents";

vi.mock("$app/environment", () => ({ browser: true }));

class FakeEventSource implements EventSourceLike {
  static instances: FakeEventSource[] = [];
  readyState = 0; // CONNECTING
  onopen: ((e: Event) => void) | null = null;
  onerror: ((e: Event) => void) | null = null;
  closed = false;
  private listeners = new Map<string, Array<(e: MessageEvent<string>) => void>>();

  constructor(public url: string) {
    FakeEventSource.instances.push(this);
  }

  addEventListener(type: string, cb: (e: MessageEvent<string>) => void) {
    const arr = this.listeners.get(type) ?? [];
    arr.push(cb);
    this.listeners.set(type, arr);
  }

  close() {
    this.closed = true;
    this.readyState = 2;
  }

  emit(type: string, data: string) {
    for (const cb of this.listeners.get(type) ?? []) {
      cb(new MessageEvent(type, { data }));
    }
  }

  open() {
    this.readyState = 1;
    this.onopen?.(new Event("open"));
  }

  fail() {
    this.onerror?.(new Event("error"));
  }
}

function createSource(url: string): EventSourceLike {
  return new FakeEventSource(url);
}

describe("subscribeGraphEvents", () => {
  beforeEach(() => {
    FakeEventSource.instances = [];
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("connects to the events endpoint with the access token", () => {
    subscribeGraphEvents({
      onGraphChanged: vi.fn(),
      createSource,
      getToken: () => "tok-123",
    });
    expect(FakeEventSource.instances).toHaveLength(1);
    expect(FakeEventSource.instances[0].url).toContain("/v1/graph/events");
    expect(FakeEventSource.instances[0].url).toContain("access_token=tok-123");
  });

  it("does not connect without a token", () => {
    const handle = subscribeGraphEvents({
      onGraphChanged: vi.fn(),
      createSource,
      getToken: () => null,
    });
    expect(FakeEventSource.instances).toHaveLength(0);
    expect(handle.isConnected()).toBe(false);
  });

  it("delivers a burst of change events as a single callback", () => {
    const onGraphChanged = vi.fn();
    subscribeGraphEvents({ onGraphChanged, createSource, getToken: () => "t" });
    const es = FakeEventSource.instances[0];
    es.open();

    es.emit("graph", "{}");
    es.emit("graph", "{}");
    es.emit("graph", "{}");
    vi.advanceTimersByTime(400);
    expect(onGraphChanged).toHaveBeenCalledTimes(1);
  });

  it("reconnects after the stream dies", () => {
    subscribeGraphEvents({ onGraphChanged: vi.fn(), createSource, getToken: () => "t" });
    const es = FakeEventSource.instances[0];
    es.open();
    es.fail();

    expect(es.closed).toBe(true);
    vi.advanceTimersByTime(2000);
    expect(FakeEventSource.instances).toHaveLength(2);
    FakeEventSource.instances[1].open();
  });

  it("backs off on repeated failures", () => {
    subscribeGraphEvents({ onGraphChanged: vi.fn(), createSource, getToken: () => "t" });
    FakeEventSource.instances[0].open();
    FakeEventSource.instances[0].fail();
    vi.advanceTimersByTime(2000); // reconnect #1 scheduled at min backoff
    expect(FakeEventSource.instances).toHaveLength(2);
    FakeEventSource.instances[1].fail();
    vi.advanceTimersByTime(2000); // backoff doubled — too early for #2
    expect(FakeEventSource.instances).toHaveLength(2);
    vi.advanceTimersByTime(2000); // 4 s total — reconnect #2
    expect(FakeEventSource.instances).toHaveLength(3);
  });

  it("stop() prevents reconnects and closes the stream", () => {
    const handle = subscribeGraphEvents({
      onGraphChanged: vi.fn(),
      createSource,
      getToken: () => "t",
    });
    const es = FakeEventSource.instances[0];
    es.open();
    es.fail();

    handle.stop();
    vi.advanceTimersByTime(60000);
    expect(FakeEventSource.instances).toHaveLength(1);
  });
});
