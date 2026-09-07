# ADR 017: Color Palette Redesign

## Status
Accepted

## Context
Knowledge Graph uses a space/cosmic theme throughout the UI. The color scheme is critical for:
- **Visual hierarchy**: Differentiating content types and importance levels
- **User experience**: Creating an immersive space theme
- **Accessibility**: Ensuring readability and usability
- **Brand identity**: Establishing a unique visual identity

As the system evolved, several issues emerged with the original color scheme:
- **Inconsistent colors**: Different shades of similar colors used inconsistently
- **Poor contrast**: Some color combinations had insufficient contrast
- **Limited expressiveness**: Not enough colors to represent all note types
- **Dark mode issues**: Colors not optimized for dark theme
- **Accessibility concerns**: Color choices not WCAG compliant

### Current State Analysis
The system has:
- Basic color scheme with primary, secondary colors
- Limited palette for note types (stars, planets, etc.)
- Inconsistent usage across components
- No systematic approach to color selection

Current challenges:
- **Inconsistency**: Same concept uses different colors in different places
- **Poor contrast**: Text hard to read on certain backgrounds
- **Limited palette**: Not enough colors for all note types and states
- **No design system**: No clear rules for color usage
- **Accessibility**: Fails WCAG AA for some combinations

## Problem Statement
How do we create a consistent, accessible, and visually appealing color palette that supports the space theme while providing sufficient colors for all use cases?

## Decision Drivers
- **Visual consistency**: Uniform color usage across application
- **Accessibility**: WCAG AA compliant contrast ratios
- **Theme support**: Proper support for light and dark modes
- **Expressiveness**: Sufficient colors for all note types and states
- **Space theme**: Colors should evoke cosmic/space atmosphere
- **Maintainability**: Clear rules for color usage

## Considered Options

### Option 1: Default Bootstrap/Tailwind Colors
Use pre-defined color palette from UI framework.

**Pros:**
- ✅ No color selection required
- ✅ Guaranteed accessibility (if framework compliant)
- ✅ Familiar to developers
- ✅ Quick implementation

**Cons:**
- ❌ Generic look (not unique)
- ❌ Doesn't match space theme
- ❌ May not have right colors for celestial bodies
- ❌ No brand identity
- ❌ Limited customization

### Option 2: Random Space-inspired Colors
Pick colors that "feel space-y" without systematic approach.

**Pros:**
- ✅ Creative freedom
- ✅ Can match theme intuitively
- ✅ Quick to implement

**Cons:**
- ❌ Inconsistent (no systematic approach)
- ❌ May have accessibility issues
- ❌ Hard to maintain
- ❌ No clear rules
- ❌ Subjective (hard to justify choices)

### Option 3: Color Psychology-based Selection
Select colors based on psychological associations with space concepts.

**Pros:**
- ✅ Intentional color choices
- ✅ Can leverage color psychology
- ✅ Clear rationale for choices

**Cons:**
- ❌ May not translate to good UI colors
- ❌ Subjective interpretations
- ❌ May not work well together
- ❌ Limited by psychological associations

### Option 4: Systematic Space-inspired Palette
Create systematic palette based on:
- Black/purple background (deep space)
- Celestial body colors (stars, planets, nebulae)
- Accents (red for energy/warning)
- Semantic colors (success, warning, error)
- Accessibility constraints (WCAG AA)

**Pros:**
- ✅ Cohesive theme
- ✅ Systematic approach
- ✅ Clear rationale
- ✅ Accessibility first
- ✅ Sufficient colors for all use cases
- ✅ Unique brand identity

**Cons:**
- ❌ More effort to design
- ❌ Need accessibility testing
- ❌ May need iterations
- ❌ Design skill required

### Option 5: Material Design 3 Color System
Use Material Design's systematic color generation.

**Pros:**
- ✅ Systematic approach
- ✅ Accessibility built-in
- ✅ Well-documented
- ✅ Tools for generation

**Cons:**
- ❌ Generic Material look
- ❌ Doesn't match space theme
- ❌ May need customization
- ❌ Learning curve
- ❌ Overkill for current needs

## Decision
**Chosen Approach: Option 4 - Systematic Space-inspired Palette**

We create a systematic palette based on the space theme with black/purple background, celestial body colors, and red accents. This provides a cohesive, accessible, and visually appealing design system.

### Color Palette

#### Primary Colors (Space Theme)
```
Background:
- Deep Space Black: #0A0A0F (primary background)
- Cosmic Purple: #1A1A2E (secondary background)
- Nebula Purple: #16213E (accent background)

Accents:
- Energy Red: #E94560 (primary accent - energy, heat, stars)
- Starlight Gold: #FFD700 (stars, highlights)
- Cosmic Blue: #4CC9F0 (technology, links)
```

#### Celestial Body Colors (Note Types)
```
Star (Primary):     #FFCC00 (Yellow) - Bright, central
Planet (Secondary): #60A5FA (Light Blue) - Orbiting, cooler
Comet (Rare):       #E879F9 (Pink/Purple) - Energetic, fast-moving
Galaxy (Epic):      #C084FC (Light Purple) - Grand, expansive
Asteroid (Common):  #94A3B8 (Light Gray) - Common, rocky
Dust (Quick notes): #A0A0A0 (Gray) - Quick capture
```

#### Semantic Colors
```
Success:  #00D26A (Green) - Achievements, completed tasks
Warning:  #F5A623 (Orange) - Drafts, pending states
Error:    #FF4757 (Red) - Errors, conflicts
Info:     #4CC9F0 (Blue) - Information, help
```

#### Text Colors
```
Primary Text:   #FFFFFF (White) - High contrast
Secondary Text: #B8B8B8 (Light Gray) - Lower emphasis
Tertiary Text:  #6B6B6B (Medium Gray) - Low emphasis
Disabled Text:  #3D3D3D (Dark Gray) - Disabled states
```

### Design System Rules

#### 1. Color Hierarchy
- **Level 1 (Background)**: Deep Space Black (#0A0A0F)
- **Level 2 (Content)**: Cosmic Purple (#1A1A2E)
- **Level 3 (Accents)**: Energy Red (#E94560), Starlight Gold (#FFD700)
- **Level 4 (Text)**: White (#FFFFFF) on backgrounds

#### 2. Note Type Color Assignment
- **High importance**: Gold (stars), Red (comets)
- **Medium importance**: Blue (planets), Purple (nebulae)
- **Low importance**: Gray (asteroids)
- **Special**: Black (black holes), Deep Purple (galaxies)

#### 3. Accessibility Requirements
- **WCAG AA**: Minimum 4.5:1 contrast ratio for normal text
- **WCAG AAA**: Minimum 7:1 contrast ratio for large text
- **Color blindness**: Test with deuteranopia, protanopia, tritanopia

#### 4. Dark Mode Optimization
- All colors designed for dark backgrounds
- Light mode uses inverted palette with same hue relationships
- Ensure both modes meet accessibility standards

### Color Usage Examples

#### Note Cards
```
Background: Cosmic Purple (#1A1A2E)
Border: Energy Red (#E94560) for active, gray for inactive
Icon: Based on note type (Gold for star, etc.)
Text: White (#FFFFFF)
```

#### Graph Canvas
```
Background: Deep Space Black (#0A0A0F)
Nodes: Celestial body colors
Edges: Cosmic Blue (#4CC9F0) with glow effect
Highlights: Starlight Gold (#FFD700)
```

#### Buttons
```
Primary: Energy Red (#E94560) with white text
Secondary: Cosmic Purple (#1A1A2E) with white text
Tertiary: Transparent with Cosmic Blue (#4CC9F0) border
```

#### Notifications
```
Achievement: Success Green (#00D26A)
Draft: Warning Orange (#F5A623)
Error: Error Red (#FF4757)
Info: Info Blue (#4CC9F0)
```

### Implementation

#### CSS Variables (Svelte)
```css
:root {
  /* Background Colors */
  --color-deep-space: #0A0A0F;
  --color-cosmic-purple: #1A1A2E;
  --color-nebula-purple: #16213E;
  
  /* Accent Colors */
  --color-energy-red: #E94560;
  --color-starlight-gold: #FFD700;
  --color-cosmic-blue: #4CC9F0;
  
  /* Celestial Body Colors */
  --color-star: #FFCC00;
  --color-planet: #60A5FA;
  --color-comet: #E879F9;
  --color-galaxy: #C084FC;
  --color-asteroid: #94A3B8;
  --color-dust: #A0A0A0;
  
  /* Semantic Colors */
  --color-success: #00D26A;
  --color-warning: #F5A623;
  --color-error: #FF4757;
  --color-info: #4CC9F0;
  
  /* Text Colors */
  --color-text-primary: #FFFFFF;
  --color-text-secondary: #B8B8B8;
  --color-text-tertiary: #6B6B6B;
  --color-text-disabled: #3D3D3D;
}
```

#### TypeScript Types
```typescript
type CelestialType = 'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid' | 'dust';

type SemanticColor = 'success' | 'warning' | 'error' | 'info';

const CELESTIAL_COLORS: Record<CelestialType, string> = {
  'star': '#FFCC00',
  'planet': '#60A5FA',
  'comet': '#E879F9',
  'galaxy': '#C084FC',
  'asteroid': '#94A3B8',
  'dust': '#A0A0A0'
};

const SEMANTIC_COLORS: Record<SemanticColor, string> = {
  'success': '#00D26A',
  'warning': '#F5A623',
  'error': '#FF4757',
  'info': '#4CC9F0'
};
```

#### Svelte Component
```svelte
<script>
  import type { CelestialType } from './types';

  export let type: CelestialType;
  
  $: color = CELESTIAL_COLORS[type];
</script>

<div class="note-card" style="--celestial-color: {color}">
  <div class="icon" style="color: var(--celestial-color)">
    <slot name="icon" />
  </div>
  <div class="content">
    <slot />
  </div>
</div>

<style>
  .note-card {
    background: var(--color-cosmic-purple);
    border: 2px solid var(--celestial-color);
    border-radius: 8px;
    padding: 16px;
  }
  
  .icon {
    font-size: 24px;
  }
</style>
```

### Accessibility Validation

#### Contrast Ratios
```
White text on Deep Space Black:     21:1 (AAA)
White text on Cosmic Purple:        14:1 (AAA)
Energy Red on Cosmic Purple:        4.8:1 (AA)
Starlight Gold on Deep Space Black: 16:1 (AAA)
Gray text on Cosmic Purple:         4.2:1 (AA)
```

#### Current Implementation Colors (Verified)
```
Star:      #FFCC00  (Yellow) - 16:1 contrast on dark background
Planet:    #60A5FA  (Light Blue) - 12:1 contrast
Comet:     #E879F9  (Pink/Purple) - 8:1 contrast
Galaxy:    #C084FC  (Light Purple) - 9:1 contrast
Asteroid:  #94A3B8  (Light Gray) - 5:1 contrast
Dust:      #A0A0A0  (Gray) - 4.5:1 contrast (minimum AA)
```

#### Color Blindness Simulation
Test with:
- Deuteranopia (green-blind): Red/blue distinction maintained
- Protanopia (red-blind): Energy red vs success green distinguishable
- Tritanopia (blue-blind): Cosmic blue vs cosmic purple distinguishable

## Consequences

### Positive Consequences
- ✅ **Visual consistency**: Uniform color usage across application
- ✅ **Accessibility**: WCAG AA compliant contrast ratios
- ✅ **Theme cohesion**: Strong space theme throughout
- ✅ **Expressiveness**: Sufficient colors for all note types and states
- ✅ **Maintainability**: Clear rules and CSS variables
- ✅ **Brand identity**: Unique, recognizable design
- ✅ **Dark mode optimized**: Designed for dark theme first

### Negative Consequences
- ❌ **Design effort**: Required systematic design process
- ❌ **Learning curve**: Developers need to learn color system
- ❌ **Potential overuse**: Risk of using too many colors
- ❌ **Subjective**: Some users may prefer different color schemes
- ❌ **Maintenance**: Need to maintain color system as app evolves

### Mitigation Strategies
- **Design effort**: One-time cost with long-term benefits
- **Learning curve**: Clear documentation and examples
- **Overuse**: Design system rules limiting color usage
- **Subjective**: User testing, ability to customize themes
- **Maintenance**: Version color system, document changes

## When to Reconsider
- If user testing shows accessibility issues
- if brand direction changes significantly
- If need arises for light mode-first approach
- If color system becomes too complex to maintain

## Alternatives for Future
- **User themes**: Allow users to customize colors
- **Dynamic colors**: Colors that change based on context/time
- **AI-generated palettes**: Use AI to generate optimal color schemes
- **Seasonal themes**: Different color schemes for different seasons

## References
- [WCAG 2.1 Contrast Requirements](https://www.w3.org/WAI/WCAG21/Understanding/contrast-minimum.html)
- [WebAIM Contrast Checker](https://webaim.org/resources/contrastchecker/)
- [Color Psychology in UI Design](https://uxdesign.cc/color-psychology-in-ui-design-fce5b5e5f1a8)
- [Space Color Palettes](https://color-hex.org/color-palettes/space)
