import { describe, it, expect, afterEach } from "vitest";
import { render, fireEvent, cleanup } from "@testing-library/svelte";
import NoteForm from "./NoteForm.svelte";
import { CelestialBody } from "$entities/shared/model/celestial-body";

afterEach(() => cleanup());

// COMET-1 этап B: поля расписания показываются только для типа «comet».
describe("NoteForm — comet fields", () => {
  it("shows comet fields when type is comet", () => {
    render(NoteForm, {
      props: { types: CelestialBody.UI_TYPES, type: "comet" },
    });
    expect(document.querySelector('[data-testid="comet-fields"]')).toBeTruthy();
    expect(document.querySelector('[data-testid="comet-due"]')).toBeTruthy();
  });

  it("hides comet fields for non-comet types", () => {
    render(NoteForm, {
      props: { types: CelestialBody.UI_TYPES, type: "star" },
    });
    expect(document.querySelector('[data-testid="comet-fields"]')).toBeNull();
  });

  it("shows the reminder selector only when a due date is set", async () => {
    render(NoteForm, {
      props: { types: CelestialBody.UI_TYPES, type: "comet" },
    });
    expect(document.querySelector('[data-testid="comet-remind"]')).toBeNull();

    const due = document.querySelector('[data-testid="comet-due"]') as HTMLInputElement;
    await fireEvent.input(due, { target: { value: "2026-03-01T12:00" } });
    expect(document.querySelector('[data-testid="comet-remind"]')).toBeTruthy();
  });

  it("shows the custom minutes input only for the custom reminder choice", async () => {
    render(NoteForm, {
      props: { types: CelestialBody.UI_TYPES, type: "comet", dueAtLocal: "2026-03-01T12:00" },
    });
    expect(document.querySelector('[data-testid="comet-custom-minutes"]')).toBeNull();

    const remind = document.querySelector('[data-testid="comet-remind"]') as HTMLSelectElement;
    remind.value = "custom";
    await fireEvent.change(remind);
    expect(document.querySelector('[data-testid="comet-custom-minutes"]')).toBeTruthy();
  });

  it("does not render a done checkbox in the create form", () => {
    render(NoteForm, {
      props: { types: CelestialBody.UI_TYPES, type: "comet" },
    });
    expect(document.querySelector('[data-testid="comet-done"]')).toBeNull();
  });
});
