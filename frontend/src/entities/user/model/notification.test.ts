import { describe, it, expect } from "vitest";
import { Notification } from "./notification";

describe("Notification", () => {
  it("uses default type and duration", () => {
    const n = new Notification({ message: "Hello" });
    expect(n.type).toBe("info");
    expect(n.duration).toBe(5000);
    expect(n.icon).toBe("ℹ️");
  });

  it("selects standard icons by type", () => {
    expect(new Notification({ message: "", type: "success" }).icon).toBe("✅");
    expect(new Notification({ message: "", type: "error" }).icon).toBe("❌");
    expect(new Notification({ message: "", type: "warning" }).icon).toBe("⚠️");
  });

  it("selects galactic icons when requested", () => {
    expect(new Notification({ message: "", type: "success", useGalacticMode: true }).icon).toBe(
      "⭐"
    );
    expect(new Notification({ message: "", type: "error", useGalacticMode: true }).icon).toBe("💥");
  });

  it("produces correct CSS classes", () => {
    const n = new Notification({ message: "", type: "warning" });
    expect(n.typeClass).toBe("toast-warning");
    expect(n.cssClass()).toContain("toast-notification");
    expect(n.cssClass()).toContain("toast-warning");
    expect(n.cssClass(false)).not.toContain("visible");
  });

  it("creates achievement notification", () => {
    const n = Notification.achievement("Explorer");
    expect(n.message).toBe("Achievement unlocked: Explorer");
    expect(n.type).toBe("success");
  });
});
