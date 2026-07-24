/**
 * Theme — Value Object representing the application's visual/tonal theme.
 *
 * Currently supports `standard` and `galactic` modes. Centralizes the
 * boolean/ternary switching that is duplicated across modals and toasts.
 */

export type ThemeMode = "standard" | "galactic";

export interface ThemeProps {
  mode?: ThemeMode;
}

export class Theme {
  constructor(public readonly mode: ThemeMode = "standard") {}

  get isGalactic(): boolean {
    return this.mode === "galactic";
  }

  get isStandard(): boolean {
    return this.mode === "standard";
  }

  get useGalacticMode(): boolean {
    return this.isGalactic;
  }

  /**
   * Pick a value based on the current theme.
   */
  choose<T>(standard: T, galactic: T): T {
    return this.isGalactic ? galactic : standard;
  }

  /**
   * Return a galactic label only when in galactic mode, otherwise the
   * provided standard label.
   */
  label(standard: string, galactic: string): string {
    return this.choose(standard, galactic);
  }

  /**
   * Optionally transform a standard label when it exactly matches a known
   * default and the theme is galactic.
   */
  transformLabel(standardLabel: string, galacticMap: Record<string, string>): string {
    if (!this.isGalactic) return standardLabel;
    return galacticMap[standardLabel] ?? standardLabel;
  }

  equals(other: Theme): boolean {
    return this.mode === other.mode;
  }

  toString(): ThemeMode {
    return this.mode;
  }

  static standard(): Theme {
    return new Theme("standard");
  }

  static galactic(): Theme {
    return new Theme("galactic");
  }

  static fromBoolean(galactic: boolean): Theme {
    return galactic ? Theme.galactic() : Theme.standard();
  }

  static fromString(value: string): Theme {
    if (value === "galactic") return Theme.galactic();
    return Theme.standard();
  }
}
