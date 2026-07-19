/**
 * GraphMode — Value Object for the graph interaction/visualization mode.
 *
 * Currently the only modes are `normal` and `focus`. The object centralizes
 * the styling, iconography and toggle logic that was duplicated between
 * GraphCanvasControls and the canvas renderer.
 */

export type GraphModeType = "normal" | "focus";

export interface GraphModeProps {
  mode?: GraphModeType;
}

export class GraphMode {
  constructor(public readonly mode: GraphModeType = "normal") {}

  get isFocus(): boolean {
    return this.mode === "focus";
  }

  get isNormal(): boolean {
    return this.mode === "normal";
  }

  /** Icon for the primary mode toggle. */
  get icon(): string {
    return this.isFocus ? "👁" : "⚡";
  }

  /** Icon for the focus toggle button. */
  get focusIcon(): string {
    return this.isFocus ? "🎯" : "🔘";
  }

  get borderColor(): string {
    return this.isFocus
      ? "rgba(139, 92, 246, 0.8)"
      : "rgba(255,255,255,0.2)";
  }

  get textColor(): string {
    return this.isFocus ? "#a78bfa" : "white";
  }

  get label(): string {
    return this.isFocus ? "Focus mode" : "Normal mode";
  }

  get focusLabel(): string {
    return this.isFocus ? "Focused" : "Focus off";
  }

  toggle(): GraphMode {
    return new GraphMode(this.isFocus ? "normal" : "focus");
  }

  equals(other: GraphMode): boolean {
    return this.mode === other.mode;
  }

  toString(): GraphModeType {
    return this.mode;
  }

  static normal(): GraphMode {
    return new GraphMode("normal");
  }

  static focus(): GraphMode {
    return new GraphMode("focus");
  }

  static fromFocus(focus: boolean): GraphMode {
    return focus ? GraphMode.focus() : GraphMode.normal();
  }

  static fromString(value: string): GraphMode {
    if (value === "focus") return GraphMode.focus();
    return GraphMode.normal();
  }
}
