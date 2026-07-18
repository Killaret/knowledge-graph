/**
 * LinkType — Value Object for a graph link type.
 *
 * Encapsulates the visual style (color, dash pattern) and default weight of a
 * link so that the renderer, forms, and tooltips no longer rely on string
 * literals or scattered switch statements.
 */

export interface LinkTypeProps {
  type: string;
  label: string;
  color: string;
  lineDash: number[];
  defaultWeight: number;
  isUi?: boolean;
}

export class LinkType {
  constructor(public readonly props: Readonly<LinkTypeProps>) {}

  get type(): string {
    return this.props.type;
  }

  get label(): string {
    return this.props.label;
  }

  get color(): string {
    return this.props.color;
  }

  get lineDash(): number[] {
    return this.props.lineDash;
  }

  get defaultWeight(): number {
    return this.props.defaultWeight;
  }

  get isUi(): boolean {
    return this.props.isUi ?? true;
  }

  /**
   * Returns an RGBA color for this link type, mixing the base color with the
   * given weight and fade opacity. Weight is expected to be in [0, 1].
   */
  getColor(weight: number, fadeOpacity: number = 1): string {
    const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
    const finalOpacity = baseOpacity * fadeOpacity;

    const r = parseInt(this.color.slice(1, 3), 16);
    const g = parseInt(this.color.slice(3, 5), 16);
    const b = parseInt(this.color.slice(5, 7), 16);
    return `rgba(${r}, ${g}, ${b}, ${finalOpacity})`;
  }

  /**
   * Returns the canvas line dash pattern for this link type, optionally taking
   * the link weight into account for the `related` threshold.
   */
  getLineDash(weight?: number): number[] {
    if (this === LinkType.RELATED) {
      return (weight ?? 0.5) < 0.3 ? [6, 4] : [];
    }
    return [...this.lineDash];
  }

  /** Case-insensitive lookup. Falls back to RELATED. */
  static fromString(type: string | undefined): LinkType {
    const normalized = (type ?? "").toLowerCase().trim();
    return LinkType.MAP.get(normalized) ?? LinkType.RELATED;
  }

  static readonly REFERENCE = new LinkType({
    type: "reference",
    label: "Reference",
    color: "#3366ff",
    lineDash: [],
    defaultWeight: 0.8,
  });

  static readonly DEPENDENCY = new LinkType({
    type: "dependency",
    label: "Dependency",
    color: "#ff6600",
    lineDash: [10, 3],
    defaultWeight: 0.7,
  });

  static readonly RELATED = new LinkType({
    type: "related",
    label: "Related",
    color: "#999999",
    lineDash: [],
    defaultWeight: 0.5,
  });

  static readonly CUSTOM = new LinkType({
    type: "custom",
    label: "Custom",
    color: "#ff66ff",
    lineDash: [2, 6],
    defaultWeight: 0.5,
    isUi: false,
  });

  static readonly PARENT = new LinkType({
    type: "parent",
    label: "Parent",
    color: "#3366ff",
    lineDash: [],
    defaultWeight: 0.9,
  });

  static readonly CHILD = new LinkType({
    type: "child",
    label: "Child",
    color: "#3366ff",
    lineDash: [],
    defaultWeight: 0.9,
  });

  private static readonly ALL = [
    LinkType.REFERENCE,
    LinkType.DEPENDENCY,
    LinkType.RELATED,
    LinkType.CUSTOM,
    LinkType.PARENT,
    LinkType.CHILD,
  ] as const;

  private static readonly MAP = new Map(
    LinkType.ALL.map((linkType) => [linkType.type, linkType]),
  );

  static readonly UI_TYPES = LinkType.ALL.filter((linkType) => linkType.isUi);
}
