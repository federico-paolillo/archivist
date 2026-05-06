---
name: Archivist
colors:
  surface: '#131313'
  surface-dim: '#131313'
  surface-bright: '#393939'
  surface-container-lowest: '#0e0e0e'
  surface-container-low: '#1b1b1b'
  surface-container: '#1f1f1f'
  surface-container-high: '#2a2a2a'
  surface-container-highest: '#353535'
  on-surface: '#e2e2e2'
  on-surface-variant: '#c4c7c8'
  inverse-surface: '#e2e2e2'
  inverse-on-surface: '#303030'
  outline: '#8e9192'
  outline-variant: '#444748'
  surface-tint: '#c6c6c7'
  primary: '#ffffff'
  on-primary: '#2f3131'
  primary-container: '#e2e2e2'
  on-primary-container: '#636565'
  inverse-primary: '#5d5f5f'
  secondary: '#c8c6c5'
  on-secondary: '#313030'
  secondary-container: '#474746'
  on-secondary-container: '#b7b5b4'
  tertiary: '#ffffff'
  on-tertiary: '#303030'
  tertiary-container: '#e4e2e1'
  on-tertiary-container: '#656464'
  error: '#ffb4ab'
  on-error: '#690005'
  error-container: '#93000a'
  on-error-container: '#ffdad6'
  primary-fixed: '#e2e2e2'
  primary-fixed-dim: '#c6c6c7'
  on-primary-fixed: '#1a1c1c'
  on-primary-fixed-variant: '#454747'
  secondary-fixed: '#e5e2e1'
  secondary-fixed-dim: '#c8c6c5'
  on-secondary-fixed: '#1c1b1b'
  on-secondary-fixed-variant: '#474746'
  tertiary-fixed: '#e4e2e1'
  tertiary-fixed-dim: '#c8c6c6'
  on-tertiary-fixed: '#1b1c1c'
  on-tertiary-fixed-variant: '#474747'
  background: '#131313'
  on-background: '#e2e2e2'
  surface-variant: '#353535'
typography:
  headline-xl:
    fontFamily: Space Grotesk
    fontSize: 64px
    fontWeight: '700'
    lineHeight: '1.0'
    letterSpacing: -0.04em
  headline-lg:
    fontFamily: Space Grotesk
    fontSize: 32px
    fontWeight: '700'
    lineHeight: '1.1'
    letterSpacing: -0.02em
  body-md:
    fontFamily: Public Sans
    fontSize: 16px
    fontWeight: '400'
    lineHeight: '1.5'
    letterSpacing: 0em
  body-sm:
    fontFamily: Public Sans
    fontSize: 14px
    fontWeight: '400'
    lineHeight: '1.4'
    letterSpacing: 0em
  label-mono:
    fontFamily: Space Grotesk
    fontSize: 12px
    fontWeight: '500'
    lineHeight: '1.0'
    letterSpacing: 0.1em
spacing:
  unit: 4px
  gutter: 16px
  margin: 32px
  container-max: 1440px
---

## Brand & Style

This design system is built upon the aesthetic of "Digital Forensics" and "Institutional Permanence." It evokes the cold, unyielding atmosphere of a high-security archive or a government data vault. The personality is uncompromising, utilitarian, and intentionally imposing, designed to make the user feel like a silent operator within a vast, monolithic system.

The style is a strict interpretation of **Brutalism**. It rejects soft curves, decorative flourishes, and friendly affordances in favor of raw structural honesty. Every element serves a functional purpose, organized within a rigid hierarchy that suggests the data contained within is heavy, permanent, and potentially dangerous.

**Target Audience:** Power users, data archivists, and individuals who prioritize systemic order and security over aesthetic comfort.
**Emotional Response:** Awe, seriousness, focus, and a sense of "Information Gravity."

## Colors

The palette is restricted to a monochromatic spectrum to maintain institutional coldness and high legibility. 

- **Neutral (Black):** Used for the primary background surfaces. It should feel like an infinite void.
- **Primary (Stark White):** Reserved for critical text and high-priority interactive states. It should appear to "pierce" the darkness.
- **Secondary (Deep Gray):** Used for structural elements, containers, and inactive states. 
- **Tertiary (Mid Gray):** Used for borders and secondary data visualizations.

Avoid gradients, transparency, or soft blurs. All color transitions must be immediate and high-contrast.

## Typography

Typography is the primary structural tool of this design system. 

- **Headlines:** Space Grotesk is used in heavy weights. It provides a technical, geometric, and slightly futuristic edge that feels industrial and precise.
- **Body:** Public Sans is used for its "Institutional/Government" clarity. It is neutral, legible, and devoid of personality, which fits the cold, archival theme.
- **Labels:** Small, uppercase Space Grotesk with expanded letter spacing is used for metadata, status indicators, and system IDs to mimic serial numbers and technical readouts.

Text should be left-aligned whenever possible to reinforce the rigid, vertical grid.

## Layout & Spacing

This design system utilizes a **Fixed Grid** philosophy. Content is locked into a strict 12-column system with heavy gutters. The rhythm is dictated by a 4px baseline unit.

- **Margins:** Large, generous margins (32px+) are used to frame content, making it feel like a specimen on a slide.
- **Gutters:** 16px gutters act as "fissures" between data blocks.
- **Alignment:** Elements must snap to the grid. Avoid centering; prefer hard left-alignment or justified blocks for text-heavy data to create "solid walls" of information.
- **Density:** High information density is preferred. The UI should feel like a dashboard of critical metrics rather than a consumer-friendly experience.

## Elevation & Depth

Depth is conveyed through **Bold Borders** and tonal stacking rather than shadows. 

- **Borders:** All containers must have a solid border (minimum 2px). Secondary borders are used to nest information within primary blocks.
- **Tonal Layers:** Deepening shades of gray (#000000 to #1A1A1A to #333333) are used to indicate nesting or focus.
- **No Shadows:** Shadows are strictly forbidden. The UI is 2D and structural.
- **Focus States:** High-contrast inversion is the primary method of showing focus. When an element is selected, it should invert (White background with Black text) to demand immediate attention.

## Shapes

The shape language is strictly **Sharp**. 

- **0px Corner Radius:** Every button, card, input, and container must have a 90-degree corner. 
- **Heavy Strokes:** Use 2px or 4px strokes for all containers to emphasize the "built" nature of the interface.
- **Geometric Rigidity:** Icons must follow the same sharp-edged, geometric rules. Avoid curves unless they are mathematically necessary for the icon's recognition.

## Components

Components are designed to be utilitarian and imposing.

- **Buttons:** Rectangular blocks with 2px borders. In the default state, they are Black with White borders and White text. On hover, they invert to solid White with Black text. No transitions or easing; the change must be instantaneous.
- **Inputs:** Simple boxes with a 2px border and a monospaced-style label above. Active state is indicated by a thicker border (4px).
- **Cards:** Heavy, bordered containers with a "Serial Number" or "ID Label" in the top-left corner using the `label-mono` style.
- **Lists:** Rows separated by 2px horizontal rules. Each row should feel like a record in a ledger.
- **Checkboxes/Radios:** Square boxes with 0px radius. The "Selected" state is a solid white fill.
- **Data Headers:** Large, bold headlines that span the full width of their container, acting as physical barriers between sections.
- **Additional Components:**
    - **System Log:** A dedicated area for real-time technical feedback.
    - **Status Indicators:** Simple, high-contrast geometric shapes (Square = Active, Empty Square = Inactive).